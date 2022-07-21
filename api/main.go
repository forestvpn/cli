package api

import (
	"context"

	forestvpn_api "github.com/forestvpn/api-client-go"
)

type ApiClientWrapper struct {
	APIClient   *forestvpn_api.APIClient
	AccessToken string
}

func (w ApiClientWrapper) CreateDevice() (*forestvpn_api.Device, error) {
	// configuration.AddDefaultHeader("X-App-Name", "com.forestvpn.web")
	// configuration.AddDefaultHeader("X-App-Version", "1.0.0")
	// id, err := auth.LoadDeviceID()

	// if err != nil {
	// 	return nil, err
	// }

	// configuration.AddDefaultHeader("X-Device-ID", id)

	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	request := *forestvpn_api.NewCreateOrUpdateDeviceRequest()
	resp, _, err := w.APIClient.DeviceApi.CreateDevice(auth).CreateOrUpdateDeviceRequest(request).Execute()
	return resp, err

}

func (w ApiClientWrapper) UpdateDevice(deviceID string, locationID string) (*forestvpn_api.Device, error) {
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	request := *forestvpn_api.NewCreateOrUpdateDeviceRequest()
	request.SetLocation(locationID)
	resp, _, err := w.APIClient.DeviceApi.UpdateDevice(auth, deviceID).CreateOrUpdateDeviceRequest(request).Execute()
	return resp, err
}

func (w ApiClientWrapper) GetLocations() ([]forestvpn_api.Location, error) {
	resp, _, err := w.APIClient.GeoApi.ListLocations(context.Background()).Execute()
	return resp, err
}

func (w ApiClientWrapper) GetBillingFeatures() ([]forestvpn_api.BillingFeature, error) {
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	resp, _, err := w.APIClient.BillingApi.ListBillingFeatures(auth).Execute()
	return resp, err
}

func GetApiClient(accessToken string, apiHost string) ApiClientWrapper {
	configuration := forestvpn_api.NewConfiguration()
	configuration.Host = apiHost
	client := forestvpn_api.NewAPIClient(configuration)
	wrapper := ApiClientWrapper{APIClient: client, AccessToken: accessToken}
	return wrapper
}
