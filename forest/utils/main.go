package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
)

var AppDIR = os.Getenv("HOME") + "/.forestvpn"
var AuthDIR = AppDIR + "/auth"
var FirebaseAuthDIR = AuthDIR + "/firebase"
var firebaseAuthFile = FirebaseAuthDIR + "/firebase.json"

// Creates directories structure
func Init() {
	for _, path := range []string{AppDIR, AuthDIR, FirebaseAuthDIR} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(path, 0755)
		}
	}
}

func JsonDump(data []byte, filepath string) error {
	var localError error
	file, err := os.Create(filepath)
	if err == nil {
		defer file.Close()
		n, err := file.WriteString(string(data))

		if err == nil {
			if n != len(string(data)) {
				localError = errors.New("error writing json to file")
			}
		} else {
			localError = err
		}
	} else {
		localError = err
	}
	return localError
}

func JsonLoad(filepath string) (map[string]string, error) {
	var data map[string]string
	var localError error
	file, err := os.Open(filepath)
	if err == nil {
		defer file.Close()
		byteStream, err := ioutil.ReadAll(file)

		if err == nil {
			json.Unmarshal(byteStream, &data)
		} else {
			localError = err
		}
	} else {
		localError = err
	}
	return data, localError
}

func LoadIDToken() string {
	var token string
	data, err := JsonLoad(firebaseAuthFile)

	if err == nil {
		token = data["idToken"]
	}
	return token
}

func HandleFirebaseSignInUpResponse(response *resty.Response, successMessage string) error {
	var body map[string]map[string]string
	json.Unmarshal(response.Body(), &body)

	if body["error"] != nil {
		respError := body["error"]
		return errors.New(respError["message"])
	}

	err := JsonDump(response.Body(), firebaseAuthFile)

	if err == nil {
		color.Green(successMessage)
	}
	return err
}
