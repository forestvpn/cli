package utils

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

var filepath = "/tmp/test.json"
var data = make(map[string]string)

func TestInit(t *testing.T) {
	Init()

	for _, path := range []string{AppDir, AuthDir, DeviceDir} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error(err.Error())
		}
	}
}

func TestJsonDump(t *testing.T) {
	data["test"] = "data"
	jsonData, err := json.Marshal(data)

	if err != nil {
		t.Error(err.Error())
	}

	err = JsonDump(jsonData, filepath)

	if err != nil {
		t.Error(err.Error())
	}

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Error(err.Error())
	}
}

func TestJsonLoad(t *testing.T) {
	data["test"] = "data"
	loadedData, err := JsonLoad(filepath)

	if err != nil {
		t.Error(err.Error())
	}

	jsonData1, err := json.Marshal(loadedData)

	if err != nil {
		t.Error(err.Error())
	}

	jsonData2, err := json.Marshal(data)

	if err != nil {
		t.Error(err.Error())
	}

	if !bytes.Equal(jsonData1, jsonData2) {
		t.Errorf("%b != %b; want ==", jsonData1, jsonData2)
	}
}
