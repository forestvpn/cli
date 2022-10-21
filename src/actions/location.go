package actions

import (
	"os"
	"runtime"
	"sort"
	"strings"

	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/forestvpn/cli/auth"
	"github.com/forestvpn/cli/utils"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/ini.v1"
)

// ListLocations is a function to get the list of locations available for user.
//
// See https://github.com/forestvpn/api-client-go/blob/main/docs/GeoApi.md#listlocations for more information.
func (w AuthClientWrapper) ListLocations(country string) error {
	var data [][]string
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

	wrappedLocations := GetLocationWrappers(locations)

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
func (w AuthClientWrapper) SetLocation(device *forestvpn_api.Device, user_id string) error {
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

	path := auth.ProfilesDir + user_id + auth.WireguardConfig
	err = config.SaveTo(path)

	if err != nil {
		return err
	}

	return nil
}

type LocationWrapper struct {
	Location forestvpn_api.Location
	Premium  bool
}

func GetLocationWrappers(locations []forestvpn_api.Location) []LocationWrapper {
	var wrappers []LocationWrapper

	for _, location := range locations {
		wrapper := LocationWrapper{Location: location, Premium: IsPremiumLocation(location)}
		wrappers = append(wrappers, wrapper)
	}

	return wrappers
}

func IsPremiumLocation(location forestvpn_api.Location) bool {
	if location.GetId() == "7fc5b17c-eddf-413f-8b37-9d36eb5e33ec" || location.GetId() == "b134d679-8697-4dc6-b629-c4c189392fca" {
		return false
	}
	return true
}
