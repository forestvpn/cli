package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var APP_DIR = os.Getenv("HOME") + "/.forest"
var AUTH_DIR = APP_DIR + "/auth"
var FB_AUTH_DIR = AUTH_DIR + "/firebase"

func init() {
	for _, path := range []string{APP_DIR, AUTH_DIR, FB_AUTH_DIR} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(APP_DIR, 0755)
		}
	}

}

func jsonDump(data map[string]any, filename string) error {
	file, err := json.MarshalIndent(data, "", " ")
	if err == nil {
		err = ioutil.WriteFile(filename+".json", file, 0755)
	}
	return err
}

func jsonLoad(filepath string) map[string]any {
	var data map[string]any
	file, _ := os.Open(filepath)
	defer file.Close()
	byteStream, _ := ioutil.ReadAll(file)
	json.Unmarshal(byteStream, &data)
	return data
}
