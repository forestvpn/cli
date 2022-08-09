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
	"github.com/go-resty/resty/v2"
)

var home, _ = os.UserHomeDir()
var AppDir = home + "/.forestvpn/"

// var AuthDir = AppDir + "auth/"
var FirebaseAuthFile = AppDir + "firebase.json"

// var DeviceDir = AppDir + "device/"
var DeviceFile = AppDir + "device.json"

// var WireguardDir = AppDir + "wireguard/"
var WireguardConfig = AppDir + "fvpn0.conf"
var SessionFile = AppDir + "session.json"

// Creates directories structure
func Init() error {
	if _, err := os.Stat(AppDir); os.IsNotExist(err) {
		os.Mkdir(AppDir, 0755)
	}
	return nil
}

func JsonDump(data []byte, filepath string) error {
	file, err := os.Create(filepath)

	if err != nil {
		return err
	}

	// err = os.Chmod(filepath, 0755)

	// if err != nil {
	// 	return err
	// }

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

func readFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)

	if err != nil {
		return []byte(""), err
	}

	defer file.Close()
	return ioutil.ReadAll(file)

}

func JsonLoad(filepath string) (map[string]string, error) {
	var data map[string]string
	byteStream, err := readFile(filepath)

	if err == nil {
		json.Unmarshal(byteStream, &data)
	}

	return data, err
}

func loadKey(key string, file string) (string, error) {
	data, err := JsonLoad(file)
	return data[key], err
}

func LoadAccessToken() (string, error) {
	return loadKey("access_token", FirebaseAuthFile)
}

func HandleFirebaseAuthResponse(response *resty.Response) error {
	if response.IsError() {
		var body map[string]map[string]string
		json.Unmarshal(response.Body(), &body)
		msg := body["error"]
		return errors.New(msg["message"])
	}
	return nil
}

func HandleFirebaseSignInResponse(response *resty.Response) error {
	err := HandleFirebaseAuthResponse(response)

	if err != nil {
		return err
	}

	return JsonDump(response.Body(), FirebaseAuthFile)
}

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

func IsRefreshTokenExists() bool {
	_, err := LoadRefreshToken()
	return err == nil
}

func IsDeviceCreated() bool {
	_, err := readFile(DeviceFile)
	return err == nil
}

func LoadDeviceID() (string, error) {
	key, err := loadKey("id", DeviceFile)

	if err != nil {
		return "", err
	}
	return key, nil
}

type LocationWrapper struct {
	Location forestvpn_api.Location
	Type     string
}

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

func IsAuthenticated() bool {
	accessToken, err := LoadAccessToken()

	if err != nil {
		return false
	} else if len(accessToken) < 1 {
		return false
	}
	return true

}
