// actions is a package containing a high-level structure that implements the functions to use as CLI Actions.
//
// See https://cli.urfave.org/v2/ for more information.

package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
	"gopkg.in/ini.v1"

	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/forestvpn/cli/api"
	"github.com/forestvpn/cli/auth"
	"github.com/forestvpn/cli/utils"
)

// AuthClientWrapper is a structure that is used as a high-level wrapper for both AuthClient and ApiClient.
// It is used as main wgrest and Firebase REST API client as both of wrapped structures share the same AccessToken for authentication purposes.
type AuthClientWrapper struct {
	AuthClient auth.AuthClient
	ApiClient  api.ApiClientWrapper
}

// Register is a method to perform a user registration on Firebase.
//
// See https://firebase.google.com/docs/reference/rest/auth#section-create-email-password for more information.
func (w AuthClientWrapper) Register(email string, password string) error {
	signinform := auth.SignInForm{}
	emailfield, err := auth.GetEmailField(email)

	if err != nil {
		return err
	}

	signinform.EmailField = emailfield

	if err != nil {
		return err
	}

	response, err := w.AuthClient.GetUserData(emailfield.Value)

	if err != nil {
		return err
	}

	var data map[string]bool
	json.Unmarshal(response.Body(), &data)

	if data["registered"] {
		return errors.New("a profile for this user already exists")
	}

	validate := true
	passwordfield, err := auth.GetPasswordField([]byte(password), validate)

	if err != nil {
		return err
	}

	err = passwordfield.Validate()

	if err != nil {
		return err
	}

	signinform.PasswordField = passwordfield
	signupform := auth.SignUpForm{}
	fmt.Print("Confirm password: ")
	passwordConfirmation, err := term.ReadPassword(0)
	fmt.Println()

	if err != nil {
		return err
	}

	signupform.PasswordConfirmationField.Value = passwordConfirmation
	signupform.SignInForm = signinform
	err = signupform.ValidatePasswordConfirmation()

	if err != nil {
		return err
	}

	response, err = w.AuthClient.SignUp(signupform)

	if err != nil {
		return err
	}

	err = auth.HandleFirebaseAuthResponse(response)

	if err != nil {
		return err
	}

	jsonresponse := make(map[string]string)
	json.Unmarshal(response.Body(), &jsonresponse)
	refreshToken := jsonresponse["refreshToken"]
	response, err = w.AuthClient.ExchangeRefreshForIdToken(refreshToken)

	if err != nil {
		return err
	}

	err = auth.JsonDump(response.Body(), auth.FirebaseAuthFile)

	if err != nil {
		return err
	}

	accessToken, err := auth.LoadAccessToken()

	if err != nil {
		return err
	}

	w.ApiClient.AccessToken = accessToken
	resp, err := w.ApiClient.CreateDevice()

	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(resp, "", "    ")

	if err != nil {
		return err
	}

	err = auth.JsonDump(b, auth.DeviceFile)

	if err == nil {
		color.Green("Signed up")
	}

	return err
}

// Login is a method for logging in a user on the Firebase.
// Accepts the deviceID (coming from local file) which indicates wether the device was created on previous login.
// If the deviceID is empty, then should create a new device on login.
//
// See https://firebase.google.com/docs/reference/rest/auth#section-sign-in-email-password for more information.
func (w AuthClientWrapper) Login(email string, password string, deviceID string) error {
	if !auth.IsRefreshTokenExists() {
		signinform := auth.SignInForm{}
		emailfield, err := auth.GetEmailField(email)

		if err != nil {
			return err
		}

		signinform.EmailField = emailfield
		response, err := w.AuthClient.GetUserData(emailfield.Value)

		if err != nil {
			return err
		}

		var data map[string]bool
		json.Unmarshal(response.Body(), &data)

		if !data["registered"] {
			return errors.New("the user doesn't exist")
		}

		validate := false
		passwordfield, err := auth.GetPasswordField([]byte(password), validate)

		if err != nil {
			return err
		}

		signinform.PasswordField = passwordfield
		response, err = w.AuthClient.SignIn(signinform)

		if err != nil {
			return err
		}

		if response.IsError() {
			// This is stupid! We know that email is ok on the assumption of the code above.
			// I don't want to show this error message, but nobody cares about my opinion here.
			// We even have a Firebase error codes to determine exact error. E.g. INVALID_PASSWORD, INVALID_EMAIL, etc.
			return errors.New("invalid email or password")
		}

		err = auth.HandleFirebaseSignInResponse(response)

		if err != nil {
			return err
		}

		response, err = w.AuthClient.GetAccessToken()

		if err != nil {
			return err
		}

		err = auth.JsonDump(response.Body(), auth.FirebaseAuthFile)

		if err != nil {
			return err
		}
	}

	if len(deviceID) == 0 {
		accessToken, err := auth.LoadAccessToken()

		if err != nil {
			return err
		}

		w.ApiClient.AccessToken = accessToken
		response, err := w.ApiClient.CreateDevice()

		if err != nil {
			return err
		}

		b, err := json.MarshalIndent(response, "", "    ")

		if err != nil {
			return err
		}

		err = auth.JsonDump(b, auth.DeviceFile)

		if err != nil {
			return err
		}

	}

	color.Green("Logged in")
	return nil
}

// Logout is a method that removes FirebaseAuthFile, i.e. logs out the user.
func (w AuthClientWrapper) Logout() error {
	if _, err := os.Stat(auth.FirebaseAuthFile); !os.IsNotExist(err) {
		err = os.Remove(auth.FirebaseAuthFile)

		if err != nil {
			return err
		}
	}
	color.Green("Logged out")
	return nil
}

// ListLocations is a function to get the list of locations available for user.
//
// See https://github.com/forestvpn/api-client-go/blob/main/docs/GeoApi.md#listlocations for more information.
func (w AuthClientWrapper) ListLocations(country string) error {
	var data [][]string
	resp, err := w.ApiClient.GetBillingFeatures()

	if err != nil {
		return err
	}

	billingFeature := resp[0]
	locations, err := w.ApiClient.GetLocations()

	if err != nil {
		return err
	}

	if len(country) > 0 {
		var locationsByCountry []forestvpn_api.Location

		for _, location := range locations {
			if strings.EqualFold(location.Country.GetName(), country) {
				locationsByCountry = append(locationsByCountry, location)
			}
		}

		if len(locationsByCountry) > 0 {
			locations = locationsByCountry
		}
	}

	sort.Slice(locations, func(i, j int) bool {
		return locations[i].GetName() < locations[j].GetName() && locations[i].Country.GetName() < locations[j].Country.GetName()
	})

	wrappedLocations := GetWrappedLocations(billingFeature, locations)

	for _, loc := range wrappedLocations {
		premiumMark := ""

		if loc.Premium {
			premiumMark = "*"
		}
		data = append(data, []string{loc.Location.GetName(), loc.Location.Country.GetName(), loc.Location.GetId(), premiumMark})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"City", "Country", "UUID", "Premium"})
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()
	return nil
}

// SetLocation is a function that writes the location data into the Wireguard configuration file.
// It uses gopkg.in/ini.v1 package to form Woreguard compatible configuration file from the location data.
// If the user subscrition on the Forest VPN services is out of date, it calls BuyPremiumDialog.
//
// See https://github.com/forestvpn/api-client-go/blob/main/docs/BillingFeature.md for more information.
func (w AuthClientWrapper) SetLocation(billingFeature forestvpn_api.BillingFeature, location LocationWrapper, includeHostIP bool) error {
	expireDate := billingFeature.GetExpiryDate()
	now := time.Now()

	if !expireDate.After(now) && location.Premium {
		return auth.BuyPremiumDialog()
	}

	deviceID, err := auth.LoadDeviceID()

	if err != nil {
		return err
	}

	device, err := w.ApiClient.UpdateDevice(deviceID, location.Location.GetId())

	if err != nil {
		return err
	}

	config := ini.Empty()
	interfaceSection, err := config.NewSection("Interface")

	if err != nil {
		return err
	}

	_, err = interfaceSection.NewKey("Address", strings.Join(device.GetIps()[:], ","))

	if err != nil {
		return err
	}

	_, err = interfaceSection.NewKey("PrivateKey", device.Wireguard.GetPrivKey())

	if err != nil {
		return err
	}

	_, err = interfaceSection.NewKey("DNS", strings.Join(device.GetDns()[:], ","))

	if err != nil {
		return err
	}

	for _, peer := range device.Wireguard.GetPeers() {
		peerSection, err := config.NewSection("Peer")

		if err != nil {
			return err
		}

		existingRoutes, err := utils.GetExistingRoutes()

		if err != nil {
			return err
		}

		activeSShClientIps, err := utils.GetActiveSshClientIps()

		if err != nil {
			return err
		}

		disallowed := append(existingRoutes, activeSShClientIps...)
		allowed := peer.GetAllowedIps()
		allowedIps, err := utils.ExcludeDisallowedIps(allowed, disallowed)

		if err != nil {
			return err
		}

		_, err = peerSection.NewKey("AllowedIPs", strings.Join(allowedIps, ", "))

		if err != nil {
			return err
		}

		_, err = peerSection.NewKey("Endpoint", peer.GetEndpoint())

		if err != nil {
			return err
		}

		_, err = peerSection.NewKey("PublicKey", peer.GetPubKey())

		if err != nil {
			return err
		}

		presharedKey := peer.GetPsKey()

		if len(presharedKey) > 0 {
			_, err = peerSection.NewKey("PresharedKey", presharedKey)
		}

		if err != nil {
			return err
		}
	}

	err = config.SaveTo(auth.WireguardConfig)

	if err != nil {
		return err
	}

	return nil
}
