package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

var APP_DIR = os.Getenv("HOME") + "/.forestvpn"
var AUTH_DIR = APP_DIR + "/auth"
var FB_AUTH_DIR = AUTH_DIR + "/firebase"

// Creates directories structure
func Init() {
	for _, path := range []string{APP_DIR, AUTH_DIR, FB_AUTH_DIR} {
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

func JsonLoad(filepath string) (map[string]any, error) {
	var data map[string]any
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
