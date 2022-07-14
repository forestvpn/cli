package actions

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/forestvpn/cli/api"
	"github.com/forestvpn/cli/auth"
	"github.com/forestvpn/cli/utils"
	"github.com/getsentry/sentry-go"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/ini.v1"
)

func ListLocations(locations []forestvpn_api.Location, country string) error {
	var data [][]string

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

func SetLocation(location forestvpn_api.Location, includeHostIP bool) error {

	accessToken, err := auth.LoadAccessToken()

	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	resp, err := api.GetBillingFeatures(accessToken)

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

	device, err := api.UpdateDevice(accessToken, deviceID, location.GetId())

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

		allowedIps, err := utils.GetAllowedIpsLocal(peer)
		fmt.Println(allowedIps)

		if err != nil {
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
