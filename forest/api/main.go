package api

import (
	"os"

	"github.com/go-resty/resty/v2"
)

var client = resty.New()
var ApiURL = os.Getenv("STAGING_API_URL")

func CreateDevice(accessToken string) (*resty.Response, error) {
	url := ApiURL + "devices/"
	return client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(accessToken).
		Post(url)
}

func UpdateDevice(accessToken string, DeviceId string, LocationId string) (*resty.Response, error) {
	url := ApiURL + "devices/{deviceID}/"
	return client.R().
		SetAuthToken(accessToken).
		SetPathParam("deviceID", DeviceId).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"location": LocationId}).
		Patch(url)
}

// func GetWireguards(accessToken string) (*resty.Response, error) {
// 	deviceId, err := utils.LoadDeviceID()

// 	if err != nil {
// 		return nil, err
// 	}
// 	url := fmt.Sprintf("%sdevices/%s/wireguards/", ApiURL, deviceId)
// 	return client.R().
// 		SetHeader("Content-Type", "application/json").
// 		SetAuthToken(accessToken).
// 		Get(url)
// }

func GetLocations() (*resty.Response, error) {
	url := ApiURL + "locations/"
	return client.R().Get(url)
}
