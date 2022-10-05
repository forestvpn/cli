// actions is a package containing a high-level structure that implements the functions to use as CLI Actions.
//
// See https://cli.urfave.org/v2/ for more information.

package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/go-resty/resty/v2"
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

	validate := true
	passwordfield, err := auth.GetPasswordField([]byte(password), validate)

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

	response, err := w.AuthClient.SignUp(signupform)

	if err != nil {
		return err
	}

	err = auth.HandleFirebaseSignUpResponse(response)

	if err != nil {
		return err
	}

	return w.SetUpProfile(response)
}

func (w AuthClientWrapper) SetUpProfile(response *resty.Response) error {
	var user_id string
	var accessToken string
	var refreshToken string
	var expires_in string
	var data map[string]any

	err := json.Unmarshal(response.Body(), &data)

	if err != nil {
		return err
	}

	var x interface{} = data["refreshToken"]
	switch v := x.(type) {
	case string:
		refreshToken = v
	}

	response, err = w.AuthClient.GetAccessToken(refreshToken)

	if err != nil {
		return err
	}

	err = auth.HandleFirebaseSignInResponse(response)

	if err != nil {
		return err
	}

	data = make(map[string]any)
	err = json.Unmarshal(response.Body(), &data)

	if err != nil {
		return err
	}

	var y interface{} = data["user_id"]
	switch v := y.(type) {
	case string:
		user_id = v
	}

	if len(user_id) > 0 {
		exists := auth.ProfileExists(user_id)

		if !exists {
			var y interface{} = data["access_token"]
			switch v := y.(type) {
			case string:
				accessToken = v
			}
			w.ApiClient.AccessToken = accessToken
			device, err := w.ApiClient.CreateDevice()

			if err != nil {
				return err
			}

			path := auth.ProfilesDir + user_id
			err = os.Mkdir(path, 0755)

			if err != nil {
				return err
			}

			data, err := json.MarshalIndent(device, "", "    ")

			if err != nil {
				return err
			}

			path = auth.ProfilesDir + user_id + auth.DeviceFile
			err = auth.JsonDump(data, path)

			if err != nil {
				return err
			}

			err = w.SetLocation(device)

			if err != nil {
				return err
			}
		}

		path := auth.ProfilesDir + user_id + auth.FirebaseAuthFile
		err := auth.JsonDump(response.Body(), path)

		if err != nil {
			return err
		}

		active := auth.IsActiveProfile(user_id)

		if !active {
			err = auth.SetActiveProfile(user_id)

			if err != nil {
				return err
			}

		}

	} else {
		return errors.New("error parsing firebase sign in response: invalid user_id")
	}

	var z interface{} = data["expires_in"]
	switch v := z.(type) {
	case string:
		expires_in = v
	}

	return auth.DumpAccessTokenExpireDate(user_id, expires_in)
}

// Login is a method for logging in a user on the Firebase.
// Accepts the deviceID (coming from local file) which indicates wether the device was created on previous login.
// If the deviceID is empty, then should create a new device on login.
//
// See https://firebase.google.com/docs/reference/rest/auth#section-sign-in-email-password for more information.
func (w AuthClientWrapper) Login(email string, password string) error {
	validate := false
	signinform := auth.SignInForm{}
	emailfield, err := auth.GetEmailField(email)

	if err != nil {
		return err
	}

	signinform.EmailField = emailfield
	passwordfield, err := auth.GetPasswordField([]byte(password), validate)

	if err != nil {
		return err
	}

	signinform.PasswordField = passwordfield
	response, err := w.AuthClient.SignIn(signinform)

	if err != nil {
		return err
	}

	err = auth.HandleFirebaseSignInResponse(response)

	if err != nil {
		return err
	}

	return w.SetUpProfile(response)

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
func (w AuthClientWrapper) SetLocation(device *forestvpn_api.Device) error {
	os := runtime.GOOS
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

		var allowedIps []string

		if os == "darwin" {
			allowedIps = append(allowedIps, "0.0.0.0/0")
		} else {
			allowedIps = peer.GetAllowedIps()
			activeSShClients, err := utils.GetActiveSshClients()

			if err != nil {
				return err
			}

			if len(activeSShClients) > 0 {
				allowedIps, err = utils.ExcludeDisallowedIps(allowedIps, activeSShClients)

				if err != nil {
					return err
				}
			}
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

func (w AuthClientWrapper) LoadOrGetBillingFeature(user_id string) (forestvpn_api.BillingFeature, error) {
	var billingFeature forestvpn_api.BillingFeature
	var err error
	update := false

	if auth.BillingFeautureExists(user_id) {
		billingFeature, err = auth.LoadBillingFeature(user_id)

		if err != nil {
			return billingFeature, err
		}

		if auth.BillingFeatureExpired(billingFeature) {
			update = true
		}
	} else {
		update = true
	}

	if update {
		resp, err := w.ApiClient.GetBillingFeatures()

		if err != nil {
			return billingFeature, err
		}

		billingFeature = resp[0]
		data, err := json.MarshalIndent(billingFeature, "", "    ")

		if err != nil {
			return billingFeature, err
		}

		path := auth.ProfilesDir + user_id + auth.BillingFeatureFile
		err = auth.JsonDump(data, path)

		if err != nil {
			return billingFeature, err
		}
	}
	return billingFeature, nil
}
