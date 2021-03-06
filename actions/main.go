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
	"github.com/getsentry/sentry-go"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
	"gopkg.in/ini.v1"

	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/forestvpn/cli/api"
	"github.com/forestvpn/cli/auth"
	"github.com/forestvpn/cli/utils"
)

type AuthClientWrapper struct {
	AuthClient auth.AuthClient
	ApiClient  api.ApiClientWrapper
}

func (w AuthClientWrapper) Register(email string, password string) error {
	signinform, err := auth.GetSignInForm(email, []byte(password))

	if err != nil {
		return err
	}

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

	err = auth.HandleFirebaseAuthResponse(response)

	if err == nil {
		color.Green("Signed up")
	}

	return err
}

func (w AuthClientWrapper) Login(email string, password string) error {
	if !auth.IsRefreshTokenExists() {
		signinform, err := auth.GetSignInForm(email, []byte(password))

		if err != nil {
			return err
		}

		response, err := w.AuthClient.SignIn(signinform)

		if err != nil {
			return err
		}

		if response.StatusCode() != 200 {
			var data map[string]map[string]string
			json.Unmarshal(response.Body(), &data)
			err := data["error"]
			return errors.New(err["message"])
		}

		err = auth.HandleFirebaseSignInResponse(response)

		if err != nil {
			return err
		}

		err = auth.JsonDump(response.Body(), auth.FirebaseAuthFile)

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

	if !auth.IsDeviceCreated() {
		resp, err := w.ApiClient.CreateDevice()

		if err != nil {
			return err
		}

		b, err := json.MarshalIndent(resp, "", "    ")

		if err != nil {
			return err
		}

		err = auth.JsonDump(b, auth.DeviceFile)

		if err != nil {
			return err
		}
	}

	color.Green("Signed in")

	return nil
}

func (w AuthClientWrapper) Logout() error {
	err := os.Remove(auth.FirebaseAuthFile)

	if err == nil {
		color.Red("Signed out")
	}

	return err
}

func (w AuthClientWrapper) ListLocations(country string) error {
	var data [][]string
	locations, err := w.ApiClient.GetLocations()

	if err != nil {
		sentry.CaptureException(err)
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

	for _, loc := range locations {
		data = append(data, []string{loc.GetName(), loc.Country.GetName(), loc.GetId()})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"City", "Country", "UUID"})
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()
	return nil
}

func (w AuthClientWrapper) SetLocation(location forestvpn_api.Location, includeHostIP bool) error {
	resp, err := w.ApiClient.GetBillingFeatures()

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	billingFeature := resp[0]
	constraint := billingFeature.GetConstraints()[0]
	subject := constraint.GetSubject()
	expireDate := billingFeature.GetExpiryDate()
	now := time.Now()

	if !expireDate.After(now) {
		return auth.BuyPremiumDialog()
	}

	if !strings.Contains(strings.Join(subject[:], " "), location.GetId()) {
		return auth.BuyPremiumDialog()

	}

	deviceID, err := auth.LoadDeviceID()

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	device, err := w.ApiClient.UpdateDevice(deviceID, location.GetId())

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	config := ini.Empty()
	interfaceSection, err := config.NewSection("Interface")

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	_, err = interfaceSection.NewKey("Address", strings.Join(device.GetIps()[:], ","))

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	_, err = interfaceSection.NewKey("PrivateKey", device.Wireguard.GetPrivKey())

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	_, err = interfaceSection.NewKey("DNS", strings.Join(device.GetDns()[:], ","))

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	for _, peer := range device.Wireguard.GetPeers() {
		peerSection, err := config.NewSection("Peer")

		if err != nil {
			sentry.CaptureException(err)
			return err
		}

		existingRoutes, err := utils.GetExistingRoutes()

		if err != nil {
			sentry.CaptureException(err)
			return err
		}

		activeSShClientIps, err := utils.GetActiveSshClientIps()

		if err != nil {
			sentry.CaptureException(err)
			return err
		}

		disallowed := append(existingRoutes, activeSShClientIps...)
		allowed := peer.GetAllowedIps()
		allowedIps, err := utils.ExcludeDisallowedIpds(allowed, disallowed)

		if err != nil {
			sentry.CaptureException(err)
			return err
		}

		_, err = peerSection.NewKey("AllowedIPs", strings.Join(allowedIps, ","))

		if err != nil {
			sentry.CaptureException(err)
			return err
		}

		_, err = peerSection.NewKey("Endpoint", peer.GetEndpoint())

		if err != nil {
			sentry.CaptureException(err)
			return err
		}

		_, err = peerSection.NewKey("PublicKey", peer.GetPubKey())

		if err != nil {
			sentry.CaptureException(err)
			return err
		}

		presharedKey := peer.GetPsKey()

		if len(presharedKey) > 0 {
			_, err = peerSection.NewKey("PresharedKey", presharedKey)
		}

		if err != nil {
			sentry.CaptureException(err)
			return err
		}
	}

	err = config.SaveTo(auth.WireguardConfig)

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	color.New(color.FgGreen).Println(fmt.Sprintf("Default location is set to %s", location.GetId()))

	return nil
}
