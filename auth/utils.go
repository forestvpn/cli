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
// Read more: https://github.com/forestvpn/api-client-go.
var DeviceFile = AppDir + "device.json"

// WireguardConfig is a Wireguard configuration file.
// It's being rewrittten per location change.
var WireguardConfig = AppDir + "fvpn0.conf"

// The SessionFile is a file for storing the last session information.
// It's used to track down the status of connection.
var SessionFile = AppDir + "session.json"

// Init creates directories structure for Forest CLI
func Init() error {
	if _, err := os.Stat(AppDir); os.IsNotExist(err) {
		os.Mkdir(AppDir, 0755)
	}
	return nil
}

// JsonDump dumps the json data into the file at filepath
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

// readFile reads the content of a file at filepath
func readFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)

	if err != nil {
		return []byte(""), err
	}

	defer file.Close()
	return ioutil.ReadAll(file)

}

// JsonLoad calls the readFile function to read the content of a file and unmarshals it into the map.
func JsonLoad(filepath string) (map[string]string, error) {
	var data map[string]string
	byteStream, err := readFile(filepath)

	if err == nil {
		json.Unmarshal(byteStream, &data)
	}

	return data, err
}

// loadKey calls JsonLoad function to get the map
// It accepts the key to search in this map and the file to read the json data from.
// Returns the value of a key if it exist in the map loaded.
func loadKey(key string, file string) (string, error) {
	data, err := JsonLoad(file)
	return data[key], err
}

// LoadAccessToken calls loadKey function to find the access token in the FirebaseAuthFile.
func LoadAccessToken() (string, error) {
	return loadKey("access_token", FirebaseAuthFile)
}

// HandleFirebaseAuthResponse extracts the error message from Firebase response when the status is ok.
func HandleFirebaseAuthResponse(response *resty.Response) error {
	if response.IsError() {
		var body map[string]map[string]string
		json.Unmarshal(response.Body(), &body)
		msg := body["error"]
		return errors.New(msg["message"])
	}
	return nil
}

// HandleFirebaseSignInResponse calls HandleFirebaseAuthResponse and, if there's no any error, dumps the response body into FirebaseAuthFile calling JsonDump.
func HandleFirebaseSignInResponse(response *resty.Response) error {
	err := HandleFirebaseAuthResponse(response)

	if err != nil {
		return err
	}

	return JsonDump(response.Body(), FirebaseAuthFile)
}

// LoadRefreshToken loads and returns refresh token from FirebaseAuthFile using JsonLoad.
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

// IsRefreshTokenExists returns true if LoadRefreshToken is called without errors, i.e. refresh token is found.
func IsRefreshTokenExists() bool {
	_, err := LoadRefreshToken()
	return err == nil
}

// IsDeviceCreated checks if the DeviceFile exist using the readFile.
func IsDeviceCreated() bool {
	_, err := readFile(DeviceFile)
	return err == nil
}

// LoadDeviceID calls a loadKey to find the device ID in the DeviceFile.
func LoadDeviceID() (string, error) {
	key, err := loadKey("id", DeviceFile)

	if err != nil {
		return "", err
	}
	return key, nil
}

// BuyPremiumDialog prompts the user if he or she want to by premium subscrition on Forest VPN.
// If user prompts 'yes', then it opens https://forestvpn.com/pricing/ page in the default browser.
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

// IsAuthenticated is a helper function to quickly check wether user is authenticated.
// It calls the LoadAccessToken to check if an access token exist.
func IsAuthenticated() bool {
	accessToken, err := LoadAccessToken()

	if err != nil {
		return false
	} else if len(accessToken) < 1 {
		return false
	}
	return true

}
