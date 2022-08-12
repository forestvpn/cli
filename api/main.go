// api is package containg the ApiClientWrapper wich is used as wgrest client.
// See https://github.com/suquant/wgrest for more information.
package api

import (
	"context"
	"fmt"

	forestvpn_api "github.com/forestvpn/api-client-go"
)

// ApiClientWrapper is a astructure that wraps forestvpn_api.APIClient to extend it.
// See https://github.com/forestvpn/api-client-go for more information.
type ApiClientWrapper struct {
	APIClient   *forestvpn_api.APIClient
	AccessToken string
}

// CreateDevice sends a POST request to create a new device on the back-end for the user being logged in.
// See https://github.com/forestvpn/api-client-go/blob/main/docs/DeviceApi.md#createdevice for more information.
func (w ApiClientWrapper) CreateDevice() (*forestvpn_api.Device, error) {
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	request := *forestvpn_api.NewCreateOrUpdateDeviceRequest()
	j, _ := request.MarshalJSON()
	fmt.Println(string(j))
	resp, _, err := w.APIClient.DeviceApi.CreateDevice(auth).CreateOrUpdateDeviceRequest(request).Execute()
	return resp, err

}

// UpdateDevice updates an existing device for the user on the back-end.
// See https://github.com/forestvpn/api-client-go/blob/main/docs/DeviceApi.md#updatedevice for more information.
func (w ApiClientWrapper) UpdateDevice(deviceID string, locationID string) (*forestvpn_api.Device, error) {
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	request := *forestvpn_api.NewCreateOrUpdateDeviceRequest()
	request.SetLocation(locationID)
	resp, _, err := w.APIClient.DeviceApi.UpdateDevice(auth, deviceID).CreateOrUpdateDeviceRequest(request).Execute()
	return resp, err
}

// GetLocations gets the locations available for the user.
// See https://github.com/forestvpn/api-client-go/blob/main/docs/GeoApi.md#listlocations for more information.
func (w ApiClientWrapper) GetLocations() ([]forestvpn_api.Location, error) {
	resp, _, err := w.APIClient.GeoApi.ListLocations(context.Background()).Execute()
	return resp, err
}

// GetBillingFeatures gets the servers available for the user.
// See https://github.com/forestvpn/api-client-go/blob/main/docs/BillingApi.md#listbillingfeatures for more information.
func (w ApiClientWrapper) GetBillingFeatures() ([]forestvpn_api.BillingFeature, error) {
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	resp, _, err := w.APIClient.BillingApi.ListBillingFeatures(auth).Execute()
	return resp, err
}

// GetApiClient is a fabric function that returns the ApiClientWrapper structure.
func GetApiClient(accessToken string, apiHost string) ApiClientWrapper {
	configuration := forestvpn_api.NewConfiguration()
	configuration.Host = apiHost
	client := forestvpn_api.NewAPIClient(configuration)
	wrapper := ApiClientWrapper{APIClient: client, AccessToken: accessToken}
	return wrapper
}
