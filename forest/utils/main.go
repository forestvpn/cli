package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

var AppDIR = os.Getenv("HOME") + "/.forestvpn"
var AuthDIR = AppDIR + "/auth"
var FirebaseAuthDIR = AuthDIR + "/firebase"
var FirebaseAuthFile = FirebaseAuthDIR + "/firebase.json"

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

func loadIDToken() string {
	var token string
	data, err := JsonLoad(FirebaseAuthFile)

	if err == nil {
		token = data["idToken"]
	}
	return token
}

func IsAuthenticated() (bool, string) {
	idToken := loadIDToken()
	return len(idToken) > 0, idToken
}
