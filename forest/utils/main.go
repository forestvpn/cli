package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-resty/resty/v2"
)

func resolveAppDir() (string, error) {
	confDir, err := os.UserConfigDir()

	if err != nil {
		homeDir, err := os.UserHomeDir()

		if err != nil {
			panic(err)
		}

		return homeDir + "/.forestvpn/", err
	}
	return confDir + "/.forestvpn/", err
}

var AppDir, _ = resolveAppDir()
var AuthDir = AppDir + "auth/"
var FirebaseAuthFile = AuthDir + "firebase.json"
var DeviceDir = AppDir + "device/"
var DeviceFile = DeviceDir + "device.json"

// Creates directories structure
func Init() {
	for _, path := range []string{AppDir, AuthDir, DeviceDir} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(path, 0755)
		}
	}
}

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

func ReadFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)

	if err != nil {
		return []byte(""), err
	}

	defer file.Close()
	return ioutil.ReadAll(file)

}

func JsonLoad(filepath string) (map[string]string, error) {
	var data map[string]string
	byteStream, err := ReadFile(filepath)

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
	var body map[string]map[string]string
	json.Unmarshal(response.Body(), &body)

	if body["error"] != nil {
		respError := body["error"]
		return errors.New(respError["message"])
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

func HandleApiResponse(response *resty.Response) error {
	if response.IsError() {
		var body map[string]string
		json.Unmarshal(response.Body(), &body)
		return errors.New(body["message"])
	}
	return nil
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
	return "", errors.New("access token not found")
}

func IsRefreshTokenExists() bool {
	_, err := LoadRefreshToken()
	return err == nil
}

func IsDeviceRegistered() bool {
	_, err := os.ReadFile(DeviceFile)
	return err == nil
}

func LoadDeviceID() (string, error) {
	key, err := loadKey("id", DeviceFile)

	if err != nil {
		return "", err
	}
	return key, nil
}
