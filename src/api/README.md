<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# api

```go
import "github.com/forestvpn/cli/api"
```

api is package containg the ApiClientWrapper wich is used as wgrest client.

See https://github.com/suquant/wgrest for more information.

## Index

- [type ApiClientWrapper](<#type-apiclientwrapper>)
  - [func GetApiClient(accessToken string, apiHost string) ApiClientWrapper](<#func-getapiclient>)
  - [func (w ApiClientWrapper) CreateDevice() (*forestvpn_api.Device, error)](<#func-apiclientwrapper-createdevice>)
  - [func (w ApiClientWrapper) GetBillingFeatures() ([]forestvpn_api.BillingFeature, error)](<#func-apiclientwrapper-getbillingfeatures>)
  - [func (w ApiClientWrapper) GetConnectedLocation() (forestvpn_api.Location, error)](<#func-apiclientwrapper-getconnectedlocation>)
  - [func (w ApiClientWrapper) GetDevice(id string) (*forestvpn_api.Device, error)](<#func-apiclientwrapper-getdevice>)
  - [func (w ApiClientWrapper) GetLocations() ([]forestvpn_api.Location, error)](<#func-apiclientwrapper-getlocations>)
  - [func (w ApiClientWrapper) GetStatus() (bool, error)](<#func-apiclientwrapper-getstatus>)
  - [func (w ApiClientWrapper) UpdateDevice(deviceID string, locationID string) (*forestvpn_api.Device, error)](<#func-apiclientwrapper-updatedevice>)


## type [ApiClientWrapper](<https://github.com/forestvpn/cli/blob/main/src/api/main.go#L17-L20>)

ApiClientWrapper is a structure that wraps forestvpn\_api.APIClient to extend it.

See https://github.com/forestvpn/api-client-go for more information.

```go
type ApiClientWrapper struct {
    APIClient   *forestvpn_api.APIClient
    AccessToken string
}
```

### func [GetApiClient](<https://github.com/forestvpn/cli/blob/main/src/api/main.go#L65>)

```go
func GetApiClient(accessToken string, apiHost string) ApiClientWrapper
```

GetApiClient is a factory function that returns the ApiClientWrapper structure. It configures and wraps an instance of forestvpn\_api.APIClient.

See https://github.com/forestvpn/api-client-go for more information.

### func \(ApiClientWrapper\) [CreateDevice](<https://github.com/forestvpn/cli/blob/main/src/api/main.go#L25>)

```go
func (w ApiClientWrapper) CreateDevice() (*forestvpn_api.Device, error)
```

CreateDevice sends a POST request to create a new device on the back\-end after the profile successfully logged in.

See https://github.com/forestvpn/api-client-go/blob/main/docs/DeviceApi.md#createdevice for more information.

### func \(ApiClientWrapper\) [GetBillingFeatures](<https://github.com/forestvpn/cli/blob/main/src/api/main.go#L55>)

```go
func (w ApiClientWrapper) GetBillingFeatures() ([]forestvpn_api.BillingFeature, error)
```

GetBillingFeatures is a method for getting locations available to the profile.

See https://github.com/forestvpn/api-client-go/blob/main/docs/BillingApi.md#listbillingfeatures for more information.

### func \(ApiClientWrapper\) [GetConnectedLocation](<https://github.com/forestvpn/cli/blob/main/src/api/main.go#L109>)

```go
func (w ApiClientWrapper) GetConnectedLocation() (forestvpn_api.Location, error)
```

### func \(ApiClientWrapper\) [GetDevice](<https://github.com/forestvpn/cli/blob/main/src/api/main.go#L76>)

```go
func (w ApiClientWrapper) GetDevice(id string) (*forestvpn_api.Device, error)
```

GetDevice is a method to get the device created on the registraton of the profile.

See https://github.com/forestvpn/api-client-go/blob/main/docs/DeviceApi.md#getdevice for more information.

### func \(ApiClientWrapper\) [GetLocations](<https://github.com/forestvpn/cli/blob/main/src/api/main.go#L47>)

```go
func (w ApiClientWrapper) GetLocations() ([]forestvpn_api.Location, error)
```

GetLocations is a method for getting all the locations available at back\-end.

See https://github.com/forestvpn/api-client-go/blob/main/docs/GeoApi.md#listlocations for more information.

### func \(ApiClientWrapper\) [GetStatus](<https://github.com/forestvpn/cli/blob/main/src/api/main.go#L82>)

```go
func (w ApiClientWrapper) GetStatus() (bool, error)
```

### func \(ApiClientWrapper\) [UpdateDevice](<https://github.com/forestvpn/cli/blob/main/src/api/main.go#L36>)

```go
func (w ApiClientWrapper) UpdateDevice(deviceID string, locationID string) (*forestvpn_api.Device, error)
```

UpdateDevice updates an existing device for the profile on the back\-end.

See https://github.com/forestvpn/api-client-go/blob/main/docs/DeviceApi.md#updatedevice for more information.



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
