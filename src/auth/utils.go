// Authentication related utilities around firebase REST authentication workflow.
// See https://firebase.google.com/docs/reference/rest for more information.
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/go-resty/resty/v2"
)

var home, _ = os.UserHomeDir()

// AppDir is Forest CLI application directory.
var AppDir = home + "/.forestvpn/"

// FirebaseAuthFile is a file to dump Firebase responses.
var FirebaseAuthFile = AppDir + "firebase.json"

// The DeviceFile represents the device created for the user.
//
// Read more: https://github.com/forestvpn/api-client-go.
var DeviceFile = AppDir + "device.json"

// WireguardConfig is a Wireguard configuration file.
//
// It's being rewrittten per location change.
var WireguardConfig = AppDir + "fvpn0.conf"

// The SessionFile is a file for storing the last session information.
//
// It's used to track down the status of connection.
var SessionFile = AppDir + "session.json"

// Init is a function that creates directories structure for Forest CLI.
func Init() error {
	if _, err := os.Stat(AppDir); os.IsNotExist(err) {
		os.Mkdir(AppDir, 0755)
	}
	return nil
}

// JsonDump is a function that dumps the json data into the file at filepath.
func JsonDump(data []byte, filepath string) error {
	file, err := os.Create(filepath)

	if err != nil {
		return err
	}

	defer file.Close()
	n, err := file.WriteString(string(data))

	if err != nil {
		return err
	}

	if n != len(string(data)) {
		return fmt.Errorf("error dumping %s to %s", string(data), filepath)
	}
	return err
}

// readFile is a function that reads the content of a file at filepath
func readFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)

	if err != nil {
		return []byte(""), err
	}

	defer file.Close()
	return ioutil.ReadAll(file)

}

// JsonLoad  is a function that reads the content of a file and unmarshals it into the map.
// If there's no file or it is empty, returns an empty map.
func JsonLoad(filepath string) (map[string]string, error) {
	data := make(map[string]string)
	byteStream, err := readFile(filepath)

	if err == nil {
		json.Unmarshal(byteStream, &data)
	}

	return data, err
}

// loadKey is a function to quickly find some key in the json encoded file.
func loadKey(key string, file string) (string, error) {
	data, err := JsonLoad(file)
	return data[key], err
}

// LoadAccessToken is a function to quickly get the local access token from FirebaseAuthFile.
func LoadAccessToken() (string, error) {
	return loadKey("access_token", FirebaseAuthFile)
}

// HandleFirebaseAuthResponse is a function to extracts the error message from the Firebase REST API response when the status is ok.
func HandleFirebaseAuthResponse(response *resty.Response) error {
	if response.IsError() {
		var body map[string]map[string]string
		json.Unmarshal(response.Body(), &body)
		msg := body["error"]
		return errors.New(msg["message"])
	}
	return nil
}

// HandleFirebaseSignInResponse is a function that dumps the Firebase REST API response into FirebaseAuthFile.
func HandleFirebaseSignInResponse(response *resty.Response) error {
	err := HandleFirebaseAuthResponse(response)

	if err != nil {
		return err
	}

	return JsonDump(response.Body(), FirebaseAuthFile)
}

// LoadRefreshToken is a function to quickly get the local refresh token from FirebaseAuthFile.
func LoadRefreshToken() (string, error) {
	data, err := JsonLoad(FirebaseAuthFile)

	if err != nil {
		return "", err
	}

	token := data["refresh_token"]

	if len(token) > 0 {
		return token, err
	}

	token = data["refreshToken"]

	if len(token) > 0 {
		return token, err
	}
	return "", errors.New("refresh token not found")
}

// IsRefreshTokenExists is a function to quickly check refresh token exists locally.
func IsRefreshTokenExists() bool {
	_, err := LoadRefreshToken()
	return err == nil
}

// Deprecated: IsDeviceCreated is a function that checks if the DeviceFile exist, i.e device is created.
func IsDeviceCreated() bool {
	_, err := readFile(DeviceFile)
	return err == nil
}

// LoadDeviceID is a function to quickly get the device ID from the DeviceFile.
func LoadDeviceID() (string, error) {
	key, err := loadKey("id", DeviceFile)

	if err != nil {
		return "", err
	}
	return key, nil
}

// BuyPremiumDialog is a function that prompts the user to by premium subscrition on Forest VPN.
// If the user prompts 'yes', then it opens https://forestvpn.com/pricing/ page in the default browser.
func BuyPremiumDialog() error {
	var answer string
	var openCommand string
	os := runtime.GOOS
	switch os {
	case "windows":
		openCommand = "start"
	case "darwin":
		openCommand = "open"
	case "linux":
		openCommand = "xdg-open"
	}
	fmt.Println("Buy Premium? ([Y]es/[N]o)")
	fmt.Scanln(&answer)

	if strings.Contains("YESyesYesYEsyEsyeSyES", answer) {
		err := exec.Command(openCommand, "https://forestvpn.com/pricing/").Run()

		if err != nil {
			return err
		}
	}
	return nil
}

// IsAuthenticated is a helper function to quickly check wether user is authenticated by checking existance of an access token.
func IsAuthenticated() bool {
	accessToken, err := LoadAccessToken()

	if err != nil {
		return false
	} else if len(accessToken) < 1 {
		return false
	}
	return true

}

// IsLocationSet is a function to check wether Wireguard configuration file created after location selection.
func IsLocationSet() bool {
	_, err := os.Stat(WireguardConfig)
	return !os.IsNotExist(err)
}
