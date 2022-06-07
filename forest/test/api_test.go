package main

import (
	"bytes"
	"encoding/json"
	"forest/api"
	"forest/utils"
	"testing"
)

func TestRegisterDevice(t *testing.T) {
	response, err := api.RegisterDevice()

	if err != nil {
		t.Error(err.Error())
	}

	err = utils.HandleApiResponse(response)

	if err != nil {
		t.Error(err.Error())
	}

	body := make(map[string]string)
	json.Unmarshal(response.Body(), &body)
	id := body["id"]
	filepath := utils.DeviceDir + id + ".json"
	err = utils.JsonDump(response.Body(), filepath)

	if err != nil {
		t.Error(err.Error())
	}

	byteStream, err := utils.ReadFile(filepath)

	if err != nil {
		t.Error(err.Error())
	}

	if bytes.Equal(byteStream, response.Body()) {
		t.Errorf("%d != %d; wanted ==", byteStream, response.Body())
	}

}
