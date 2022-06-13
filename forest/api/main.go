package api

import (
	"context"
	"os"

	forestAPI "github.com/forestvpn/api-client-go"
)

func CreateDevice(accessToken string) (*forestAPI.Device, error) {
	configuration := forestAPI.NewConfiguration()
	configuration.Host = os.Getenv("STAGING_API_URL")
	var apiClient = forestAPI.NewAPIClient(configuration)
	auth := context.WithValue(context.Background(), forestAPI.ContextAccessToken, accessToken)
	request := *forestAPI.NewCreateOrUpdateDeviceRequest()
	resp, _, err := apiClient.DeviceApi.CreateDevice(auth).CreateOrUpdateDeviceRequest(request).Execute()
	return resp, err

}

func UpdateDevice(accessToken string, deviceID string, locationID string) {
	// request := *forestApiClient.NewCreateOrUpdateDeviceRequest()
	// apiClient := forestApiClient.NewAPIClient(configuration)
	// resp, r, err := apiClient.DeviceApi.UpdateDevice(context.Background(), deviceID).CreateOrUpdateDeviceRequest(request).Execute()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error when calling `DeviceApi.UpdateDevice``: %v\n", err)
	// 	fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	// }
	// fmt.Fprintf(os.Stdout, "Response from `DeviceApi.UpdateDevice`: %v\n", resp)
}

func GetLocations() ([]forestAPI.Location, error) {
	configuration := forestAPI.NewConfiguration()
	configuration.Host = os.Getenv("STAGING_API_URL")
	apiClient := forestAPI.NewAPIClient(configuration)
	resp, _, err := apiClient.GeoApi.ListLocations(context.Background()).Execute()
	return resp, err
}
