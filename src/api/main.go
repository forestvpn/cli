// api is package containg the ApiClientWrapper wich is used as wgrest client.
//
// See https://github.com/suquant/wgrest for more information.
package api

import (
	"context"
	"errors"
	"io"
	"os"
	"runtime"
	"strings"

	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/forestvpn/cli/utils"
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
	hostname, err := os.Hostname()

	if err != nil {
		return nil, err
	}

	info := map[string]string{"arch": runtime.GOARCH}
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	request := *forestvpn_api.NewCreateOrUpdateDeviceRequest()
	request.SetName(hostname)
	createOrUpdateDeviceRequestInfo := request.GetInfo()
	createOrUpdateDeviceRequestInfo.SetType(runtime.GOOS)
	createOrUpdateDeviceRequestInfo.SetInfo(info)
	request.SetInfo(createOrUpdateDeviceRequestInfo)
	dev, resp, err := w.APIClient.DeviceApi.CreateDevice(auth).CreateOrUpdateDeviceRequest(request).Execute()

	if utils.Verbose {
		buf := new(strings.Builder)
		n, err := io.Copy(buf, resp.Body)

		if err != nil {
			return dev, err
		}

		if len(buf.String()) != int(n) {
			return dev, errors.New("response body read mismatch")
		}

		utils.InfoLogger.Printf("%s %s \n %s\n", resp.Request.Method, resp.Request.URL.String(), buf.String())
	}

	return dev, err
}

// UpdateDevice updates an existing device for the user on the back-end.
//
// See https://github.com/forestvpn/api-client-go/blob/main/docs/DeviceApi.md#updatedevice for more information.
func (w ApiClientWrapper) UpdateDevice(deviceID string, locationID string) (*forestvpn_api.Device, error) {
	hostname, err := os.Hostname()

	if err != nil {
		return nil, err
	}

	info := map[string]string{"arch": runtime.GOARCH}
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	request := *forestvpn_api.NewCreateOrUpdateDeviceRequest()
	request.SetName(hostname)
	createOrUpdateDeviceRequestInfo := request.GetInfo()
	createOrUpdateDeviceRequestInfo.SetType(runtime.GOOS)
	createOrUpdateDeviceRequestInfo.SetInfo(info)
	request.SetInfo(createOrUpdateDeviceRequestInfo)
	request.SetLocation(locationID)
	dev, resp, err := w.APIClient.DeviceApi.UpdateDevice(auth, deviceID).CreateOrUpdateDeviceRequest(request).Execute()

	if utils.Verbose {
		buf := new(strings.Builder)
		n, err := io.Copy(buf, resp.Body)

		if err != nil {
			return dev, err
		}

		if len(buf.String()) != int(n) {
			return dev, errors.New("response body read mismatch")
		}

		utils.InfoLogger.Printf("%s %s \n %s\n", resp.Request.Method, resp.Request.URL.String(), buf.String())
	}

	return dev, err
}

// GetLocations is a method for getting all the locations available at back-end.
//
// See https://github.com/forestvpn/api-client-go/blob/main/docs/GeoApi.md#listlocations for more information.
func (w ApiClientWrapper) GetLocations() ([]forestvpn_api.Location, error) {
	loc, resp, err := w.APIClient.GeoApi.ListLocations(context.Background()).Execute()

	if utils.Verbose {
		buf := new(strings.Builder)
		n, err := io.Copy(buf, resp.Body)

		if err != nil {
			return loc, err
		}

		if len(buf.String()) != int(n) {
			return loc, errors.New("response body read mismatch")
		}

		utils.InfoLogger.Printf("%s %s \n %s\n", resp.Request.Method, resp.Request.URL.String(), buf.String())
	}

	return loc, err
}

// GetBillingFeatures is a method for getting locations available to the user.
//
// See https://github.com/forestvpn/api-client-go/blob/main/docs/BillingApi.md#listbillingfeatures for more information.
func (w ApiClientWrapper) GetBillingFeatures() ([]forestvpn_api.BillingFeature, error) {
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	b, resp, err := w.APIClient.BillingApi.ListBillingFeatures(auth).Execute()

	if utils.Verbose {
		buf := new(strings.Builder)
		n, err := io.Copy(buf, resp.Body)

		if err != nil {
			return b, err
		}

		if len(buf.String()) != int(n) {
			return b, errors.New("response body read mismatch")
		}

		utils.InfoLogger.Printf("%s %s \n %s\n", resp.Request.Method, resp.Request.URL.String(), buf.String())
	}

	return b, err
}

// GetApiClient is a factory function that returns the ApiClientWrapper structure.
// It configures and wraps an instance of forestvpn_api.APIClient.
//
// See https://github.com/forestvpn/api-client-go for more information.
func GetApiClient(accessToken string, apiHost string) ApiClientWrapper {
	configuration := forestvpn_api.NewConfiguration()
	configuration.Host = apiHost
	configuration.HTTPClient = utils.GetHttpClient(10)
	client := forestvpn_api.NewAPIClient(configuration)
	wrapper := ApiClientWrapper{APIClient: client, AccessToken: accessToken}
	return wrapper
}

// GetDevice is a method to get the device created on the registraton of the user.
//
// See https://github.com/forestvpn/api-client-go/blob/main/docs/DeviceApi.md#getdevice for more information.
func (w ApiClientWrapper) GetDevice(id string) (*forestvpn_api.Device, error) {
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	dev, resp, err := w.APIClient.DeviceApi.GetDevice(auth, id).Execute()

	if utils.Verbose {
		buf := new(strings.Builder)
		n, err := io.Copy(buf, resp.Body)

		if err != nil {
			return dev, err
		}

		if len(buf.String()) != int(n) {
			return dev, errors.New("response body read mismatch")
		}

		utils.InfoLogger.Printf("%s %s \n %s\n", resp.Request.Method, resp.Request.URL.String(), buf.String())
	}

	return dev, err
}

func (w ApiClientWrapper) DeleteDevice(id string) error {
	auth := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, w.AccessToken)
	resp, err := w.APIClient.DeviceApi.DeleteDevice(auth, id).Execute()

	if utils.Verbose {
		buf := new(strings.Builder)
		n, err := io.Copy(buf, resp.Body)

		if err != nil {
			return err
		}

		if len(buf.String()) != int(n) {
			return errors.New("response body read mismatch")
		}

		utils.InfoLogger.Printf("%s %s \n %s\n", resp.Request.Method, resp.Request.URL.String(), buf.String())
	}

	return err
}
