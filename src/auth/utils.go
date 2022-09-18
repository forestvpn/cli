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

	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/getsentry/sentry-go"
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
const DeviceFile string = "/device.json"

// WireguardConfig is a Wireguard configuration file.
//
// It's being rewrittten per location change.
var WireguardConfig = AppDir + "fvpn0.conf"

// Deprecated: The SessionFile is a file for storing the last session information.
//
// It's used to track down the status of connection.
var SessionFile = AppDir + "session.json"

var ProfilesDir = AppDir + "profiles/"

const ActiveUserLockFile string = "/.active.lock"

func loadFirebaseAuthFile() (map[string]any, error) {
	var firebaseAuthFile map[string]any

	data, err := readFile(FirebaseAuthFile)

	if err != nil {
		return firebaseAuthFile, err
	}

	err = json.Unmarshal(data, &firebaseAuthFile)
	return firebaseAuthFile, err
}

func LoadAccessToken() (string, error) {
	var accessToken string
	firebaseAuthFile, err := loadFirebaseAuthFile()

	if err != nil {
		return accessToken, err
	}

	var y interface{} = firebaseAuthFile["access_token"]
	switch v := y.(type) {
	case string:
		accessToken = v
	}

	return accessToken, err
}

func AddProfile(user_id string, device *forestvpn_api.Device, activate bool) error {
	path := ProfilesDir + user_id

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)

		if err != nil {
			sentry.CaptureException(err)
		}
	}

	data, err := json.MarshalIndent(device, "", "    ")

	if err != nil {
		return err
	}

	err = JsonDump(data, path+DeviceFile)

	if err != nil {
		return err
	}

	if activate {
		err = SetActiveProfile(user_id)
	}
	return err
}

func LoadUserID() (string, error) {
	var id string
	auth, err := loadFirebaseAuthFile()

	if err != nil {
		return id, err
	}

	var y interface{} = auth["user_id"]
	switch v := y.(type) {
	case string:
		id = v
	}

	return id, err
}

// Init is a function that creates directories structure for Forest CLI.
func Init() error {
	for _, dir := range []string{AppDir, ProfilesDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.Mkdir(dir, 0755)

			if err != nil {
				sentry.CaptureException(err)
			}
		}
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

// HandleFirebaseAuthResponse is a function to extracts the error message from the Firebase REST API response when the status is ok.
func HandleFirebaseAuthResponse(response *resty.Response) error {
	var body map[string]any
	err := json.Unmarshal(response.Body(), &body)

	if err != nil {
		return err
	}

	var x interface{} = body["error"]
	switch v := x.(type) {
	case map[string]any:
		var x interface{} = v["message"]
		switch v := x.(type) {
		case string:
			return errors.New(v)
		}
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
	firebaseAuthFile, err := loadFirebaseAuthFile()

	if err != nil {
		return "", err
	}

	var refreshToken string

	var y interface{} = firebaseAuthFile["refresh_token"]
	switch v := y.(type) {
	case string:
		refreshToken = v
	}

	if len(refreshToken) == 0 {
		var y interface{} = firebaseAuthFile["refreshToken"]
		switch v := y.(type) {
		case string:
			refreshToken = v
		}
	}

	return refreshToken, err
}

// IsRefreshTokenExists is a function to quickly check refresh token exists locally.
func IsRefreshTokenExists() (bool, error) {
	refreshToken, err := LoadRefreshToken()

	if err != nil {
		return false, err
	}

	return len(refreshToken) > 0, err
}

// IsDeviceCreated is a function that checks if the DeviceFile exist, i.e device is created.
func IsDeviceCreated(user_id string) bool {
	_, err := os.Stat(ProfilesDir + user_id + DeviceFile)
	return !os.IsNotExist(err)
}

func LoadDevice(user_id string) (*forestvpn_api.Device, error) {
	var device *forestvpn_api.Device
	data, err := readFile(ProfilesDir + user_id + DeviceFile)

	if err != nil {
		return device, err
	}

	err = json.Unmarshal(data, &device)
	return device, err

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
	authenticated := false
	accessToken, _ := LoadAccessToken()

	if len(accessToken) > 0 {
		authenticated = true
	}
	return authenticated

}

// Deprecated: IsLocationSet is a function to check wether Wireguard configuration file created after location selection.
func IsLocationSet() bool {
	_, err := os.Stat(WireguardConfig)
	return !os.IsNotExist(err)
}

func ProfileExists(user_id string) bool {
	if _, err := os.Stat(ProfilesDir + user_id); os.IsNotExist(err) {
		return true
	}

	return false
}

func IsActiveProfile(user_id string) bool {
	path := ProfilesDir + user_id + ActiveUserLockFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true

}

func SetActiveProfile(user_id string) error {
	files, err := ioutil.ReadDir(ProfilesDir)

	if err != nil {
		return err
	}

	removed := false
	created := false

	for _, f := range files {
		if f.IsDir() {
			path := ProfilesDir + f.Name() + ActiveUserLockFile

			if _, err := os.Stat(path); !removed && !os.IsNotExist(err) {
				err := os.Remove(path)

				if err != nil {
					sentry.CaptureException(err)
				}
				removed = true
			} else if !created && f.Name() == user_id {
				_, err := os.Create(ProfilesDir + f.Name() + ActiveUserLockFile)

				if err != nil {
					return err
				}

				created = true
			} else if created && removed {
				break
			}
		}
	}
	return nil
}

func RemoveFirebaseAuthFile() error {
	if _, err := os.Stat(FirebaseAuthFile); !os.IsNotExist(err) {
		err = os.Remove(FirebaseAuthFile)

		if err != nil {
			return err
		}
	}

	return nil
}

func RemoveActiveUserLockFile() error {
	if _, err := os.Stat(ActiveUserLockFile); !os.IsNotExist(err) {
		err = os.Remove(ActiveUserLockFile)

		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateProfileDevice(device *forestvpn_api.Device) error {
	data, err := json.MarshalIndent(device, "", "    ")

	if err != nil {
		return err
	}

	user_id, err := LoadUserID()

	if err != nil {
		return err
	}

	return JsonDump(data, ProfilesDir+user_id+DeviceFile)
}
