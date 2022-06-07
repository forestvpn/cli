package api

import (
	"fmt"
	"forest/utils"
	"os"

	"github.com/go-resty/resty/v2"
)

var client = resty.New()
var ApiURL = os.Getenv("API_URL")

func RegisterDevice(accessToken string) (*resty.Response, error) {
	url := ApiURL + "devices/"
	return client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(accessToken).
		Post(url)
}

func GetWireguards(accessToken string) (*resty.Response, error) {
	deviceId, err := utils.LoadDeviceID()

	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%sdevices/%s/wireguards/", ApiURL, deviceId)
	return client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(accessToken).
		Get(url)
}

func GetLocations() (*resty.Response, error) {
	url := ApiURL + "locations/"
	return client.R().
		// SetHeader("Content-Type", "application/json").
		Get(url)
}
