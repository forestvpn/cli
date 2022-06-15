package api

import (
	"context"
	"os"

	forestAPI "github.com/forestvpn/api-client-go"
)

func CreateDevice(accessToken string) (*forestAPI.Device, error) {
	configuration := forestAPI.NewConfiguration()
	configuration.Host = os.Getenv("STAGING_API_URL")
	apiClient := forestAPI.NewAPIClient(configuration)
	auth := context.WithValue(context.Background(), forestAPI.ContextAccessToken, accessToken)
	request := *forestAPI.NewCreateOrUpdateDeviceRequest()
	resp, _, err := apiClient.DeviceApi.CreateDevice(auth).CreateOrUpdateDeviceRequest(request).Execute()
	return resp, err

}

func UpdateDevice(accessToken string, deviceID string, locationID string) (*forestAPI.Device, error) {
	configuration := forestAPI.NewConfiguration()
	configuration.Host = os.Getenv("STAGING_API_URL")
	apiClient := forestAPI.NewAPIClient(configuration)
	auth := context.WithValue(context.Background(), forestAPI.ContextAccessToken, accessToken)
	request := *forestAPI.NewCreateOrUpdateDeviceRequest()
	request.SetLocation(locationID)
	resp, _, err := apiClient.DeviceApi.UpdateDevice(auth, deviceID).CreateOrUpdateDeviceRequest(request).Execute()
	return resp, err
}

func GetLocations() ([]forestAPI.Location, error) {
	configuration := forestAPI.NewConfiguration()
	configuration.Host = os.Getenv("STAGING_API_URL")
	apiClient := forestAPI.NewAPIClient(configuration)
	resp, _, err := apiClient.GeoApi.ListLocations(context.Background()).Execute()
	return resp, err
}

func GetBillingFeatures(accessToken string) ([]forestAPI.BillingFeature, error) {
	configuration := forestAPI.NewConfiguration()
	configuration.Host = os.Getenv("STAGING_API_URL")
	apiClient := forestAPI.NewAPIClient(configuration)
	auth := context.WithValue(context.Background(), forestAPI.ContextAccessToken, accessToken)
	resp, _, err := apiClient.BillingApi.ListBillingFeatures(auth).Execute()
	return resp, err
}
