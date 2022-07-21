package auth_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/forestvpn/cli/actions"
	"github.com/forestvpn/cli/api"
	"github.com/forestvpn/cli/auth"
)

const filepath = "/tmp/test.json"

var data = make(map[string]string)
var email = os.Getenv("STAGING_EMAIL")
var password = os.Getenv("STAGING_PASSWORD")

func TestInit(t *testing.T) {
	err := os.RemoveAll(auth.AppDir)

	if err != nil {
		t.Error(err)
	}

	err = auth.Init()

	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(auth.AppDir); os.IsNotExist(err) {
		t.Error(err)
	}

	err = auth.Init()

	if err != nil {
		t.Errorf("init: %s != nil; want == ", err)
	}
}

func TestJsonDump(t *testing.T) {
	if _, err := os.Stat(filepath); os.IsExist(err) {
		err = os.Remove(filepath)

		if err != nil {
			t.Error(err)
		}
	}

	data["test"] = "data"
	jsonData, err := json.Marshal(data)

	if err != nil {
		t.Error(err.Error())
	}

	err = auth.JsonDump(jsonData, filepath)

	if err != nil {
		t.Error(err.Error())
	}

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Error(err.Error())
	}

}

func TestJsonLoad(t *testing.T) {
	data["test"] = "data"
	loadedData, err := auth.JsonLoad(filepath)

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

func TestLoadAccessToken(t *testing.T) {
	auth.FirebaseAuthFile = "/tmp/data.json"
	fakeData := make(map[string]string)
	fakeData["access_token"] = "0123456789"
	jsonData, err := json.Marshal(fakeData)

	if err != nil {
		t.Error(err)
	}

	err = auth.JsonDump(jsonData, auth.FirebaseAuthFile)

	if err != nil {
		t.Error(err)
	}

	token, err := auth.LoadAccessToken()

	if err != nil {
		t.Error(err)
	}

	token2 := fakeData["access_token"]

	if token != token2 {
		t.Errorf("%s != %s; want ==", token, token2)
	}

}

func TestHandleFirebaseAuthResponse(t *testing.T) {
	for i := 0; i < 2; i++ {

		if i > 0 {
			email = ""
			password = ""
		}

		emailfield := auth.EmailField{Value: email}
		passwordfield := auth.PasswordField{Value: []byte(password)}
		signinform := auth.SignInForm{EmailField: emailfield, PasswordField: passwordfield}
		authClient := auth.AuthClient{ApiKey: os.Getenv("STAGING_FIREBASE_API_KEY")}
		accessToken, _ := auth.LoadAccessToken()
		wrapper := api.GetApiClient(accessToken, os.Getenv("STAGING_API_URL"))
		apiClient := actions.AuthClientWrapper{AuthClient: authClient, ApiClient: wrapper}
		response, err := apiClient.AuthClient.SignIn(signinform)

		if err != nil {
			t.Error(err)
		}

		err = auth.HandleFirebaseAuthResponse(response)

		if i > 0 && err == nil {
			t.Errorf("sign in: %s == nil; want !=", err)
		}

	}

}

func TestHandleFirebaseSignInResponse(t *testing.T) {
	for i := 0; i < 2; i++ {

		if i > 0 {
			email = ""
			password = ""
		}

		emailfield := auth.EmailField{Value: email}
		passwordfield := auth.PasswordField{Value: []byte(password)}
		signinform := auth.SignInForm{EmailField: emailfield, PasswordField: passwordfield}
		authClient := auth.AuthClient{ApiKey: os.Getenv("STAGING_FIREBASE_API_KEY")}
		accessToken, _ := auth.LoadAccessToken()
		wrapper := api.GetApiClient(accessToken, os.Getenv("STAGING_API_URL"))
		apiClient := actions.AuthClientWrapper{AuthClient: authClient, ApiClient: wrapper}
		response, err := apiClient.AuthClient.SignIn(signinform)

		if err != nil {
			t.Error(err)
		}

		if i < 1 {
			if err != nil {
				t.Error(err)
			} else if _, err := os.Stat(auth.FirebaseAuthFile); os.IsNotExist(err) {
				t.Error(err)
			}
		}

		err = auth.HandleFirebaseSignInResponse(response)

		if i > 0 && err == nil {
			t.Errorf("sign in: %s == nil; want !=", err)
		}

	}

}
