// api is package containg the ApiClientWrapper wich is used as wgrest client.
//
// See https://github.com/suquant/wgrest for more information.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/forestvpn/cli/auth"
)

// ApiClientWrapper is a structure that wraps forestvpn_api.APIClient to extend it.
//
// See https://github.com/forestvpn/api-client-go for more information.
type ApiClientWrapper struct {
	APIClient   *forestvpn_api.APIClient
	AccessToken string
}

// CreateDevice sends a POST request to create a new device on the back-end after the user successfully logged in.
//
// See https://github.com/forestvpn/api-client-go/blob/main/docs/DeviceApi.md#createdevice for more information.
func (w ApiClientWrapper) CreateDevice() (*forestvpn_api.Device, error) {
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	request := *forestvpn_api.NewCreateOrUpdateDeviceRequest()
	dev, _, err := w.APIClient.DeviceApi.CreateDevice(auth).CreateOrUpdateDeviceRequest(request).Execute()
	return dev, err

}

// UpdateDevice updates an existing device for the user on the back-end.
//
// See https://github.com/forestvpn/api-client-go/blob/main/docs/DeviceApi.md#updatedevice for more information.
func (w ApiClientWrapper) UpdateDevice(deviceID string, locationID string) (*forestvpn_api.Device, error) {
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	request := *forestvpn_api.NewCreateOrUpdateDeviceRequest()
	request.SetLocation(locationID)
	resp, _, err := w.APIClient.DeviceApi.UpdateDevice(auth, deviceID).CreateOrUpdateDeviceRequest(request).Execute()
	return resp, err
}

// GetLocations is a method for getting all the locations available at back-end.
//
// See https://github.com/forestvpn/api-client-go/blob/main/docs/GeoApi.md#listlocations for more information.
func (w ApiClientWrapper) GetLocations() ([]forestvpn_api.Location, error) {
	resp, _, err := w.APIClient.GeoApi.ListLocations(context.Background()).Execute()
	return resp, err
}

// GetBillingFeatures is a method for getting locations available to the user.
//
// See https://github.com/forestvpn/api-client-go/blob/main/docs/BillingApi.md#listbillingfeatures for more information.
func (w ApiClientWrapper) GetBillingFeatures() ([]forestvpn_api.BillingFeature, error) {
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	resp, _, err := w.APIClient.BillingApi.ListBillingFeatures(auth).Execute()
	return resp, err
}

// GetApiClient is a factory function that returns the ApiClientWrapper structure.
// It configures and wraps an instance of forestvpn_api.APIClient.
//
// See https://github.com/forestvpn/api-client-go for more information.
func GetApiClient(accessToken string, apiHost string) ApiClientWrapper {
	configuration := forestvpn_api.NewConfiguration()
	configuration.Host = apiHost
	client := forestvpn_api.NewAPIClient(configuration)
	wrapper := ApiClientWrapper{APIClient: client, AccessToken: accessToken}
	return wrapper
}

// GetDevice is a method to get the device created on the registraton of the user.
//
// See https://github.com/forestvpn/api-client-go/blob/main/docs/DeviceApi.md#getdevice for more information.
func (w ApiClientWrapper) GetDevice(id string) (*forestvpn_api.Device, error) {
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	resp, _, err := w.APIClient.DeviceApi.GetDevice(auth, id).Execute()
	return resp, err
}

func IsActiveDevice(deviceID string, ApiHost string, accessToken string) (bool, error) {
	url := "https://" + ApiHost + "/v2/devices/" + deviceID + "/"
	auth.Client.SetTimeout(time.Duration(1 * time.Second))
	resp, err := auth.Client.R().
		SetAuthToken(accessToken).
		Get(url)

	if err != nil {
		return false, err
	}

	jsonResp := make(map[string]time.Time)
	json.Unmarshal(resp.Body(), &jsonResp)
	lastActiveAt := jsonResp["last_active_at"]
	fmt.Println(lastActiveAt)
	now := time.Now()
	year, month, day := lastActiveAt.Date()
	nowYear, nowMonth, nowDay := now.Date()
	hours, minutes, _ := lastActiveAt.Clock()
	nowHours, nowMinutes, _ := now.Clock()

	if year != nowYear || month != nowMonth || day != nowDay || hours != nowHours || minutes != nowMinutes {
		return false, nil
	}
	return true, nil
}
