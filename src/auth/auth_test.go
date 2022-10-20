package auth_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/forestvpn/cli/actions"
	"github.com/forestvpn/cli/api"
	"github.com/forestvpn/cli/auth"
)

const filepath = "/tmp/test.json"

var (
	email    = os.Getenv("STAGING_EMAIL")
	password = os.Getenv("STAGING_PASSWORD")
	apiKey   = os.Getenv("STAGING_FIREBASE_API_KEY")
	apiHost  = os.Getenv("STAGING_API_URL")
)

func logout() error {
	exists, err := auth.IsRefreshTokenExists()

	if err != nil {
		return err
	}

	if exists {
		user_id, err := auth.LoadUserID()

		if err != nil {
			return err
		}

		if len(user_id) > 0 {
			err = auth.RemoveFirebaseAuthFile(user_id)

			if err != nil {
				return err
			}

			err = auth.RemoveActiveUserLockFile()

			if err != nil {
				return err
			}

			m := auth.GetAccountMap(auth.AccountsMapFile)
			err = m.RemoveAccount(user_id)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func TestInit(t *testing.T) {
	err := os.RemoveAll(auth.AppDir)

	if err != nil {
		t.Error(err)
	}

	err = auth.Init()

	if err != nil {
		t.Error(err)
	}

	for _, dir := range []string{auth.AppDir, auth.ProfilesDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Error(err)
		}
	}

	err = auth.Init()

	if err != nil {
		t.Errorf("init: %s != nil; want == ", err)
	}
}

func TestJsonDump(t *testing.T) {
	var data = map[string]string{"test": "data"}

	if _, err := os.Stat(filepath); os.IsExist(err) {
		err = os.Remove(filepath)

		if err != nil {
			t.Error(err)
		}
	}

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

func TestLoadAccessTokenWhileLoggedIn(t *testing.T) {
	emailfield := auth.EmailField{Value: email}
	passwordfield := auth.PasswordField{Value: []byte(password)}
	signinform := auth.SignInForm{EmailField: emailfield, PasswordField: passwordfield}
	authClient := auth.AuthClient{ApiKey: apiKey}
	response, err := authClient.SignIn(signinform)

	if err != nil {
		t.Error(err)
	}

	err = auth.HandleFirebaseSignInResponse(response)

	if err != nil {
		t.Error(err)
	}

	refreshToken, err := auth.LoadRefreshToken()

	if err != nil {
		t.Error(err)
	}

	response, err = authClient.GetAccessToken(refreshToken)

	if err != nil {
		t.Error(err)
	}

	err = auth.JsonDump(response.Body(), auth.FirebaseAuthFile)

	if err != nil {
		t.Error(err)
	}

	accessToken, err := auth.LoadAccessToken()

	if err != nil {
		t.Error(err)
	}

	var body map[string]string
	json.Unmarshal(response.Body(), &body)
	accessToken1 := body["access_token"]

	if !strings.EqualFold(accessToken, accessToken1) {
		t.Errorf("%s != %s; want ==", accessToken, accessToken1)
	}
}

func TestLoadAccessTokenWhileLoggedOut(t *testing.T) {
	authClient := auth.AuthClient{ApiKey: apiKey}
	accessToken, _ := auth.LoadAccessToken()
	wrapper := api.GetApiClient(accessToken, apiHost)
	apiClient := actions.AuthClientWrapper{AuthClient: authClient, ApiClient: wrapper}
	err := apiClient.Logout()

	if err != nil {
		t.Error(err)
	}

	accessToken, err = auth.LoadAccessToken()

	if err == nil {
		t.Error(err)
	}

	tokenLength := len(accessToken)

	if tokenLength > 0 {
		t.Errorf("%d > 0; want <=", tokenLength)
	}
}

func TestHandleFirebaseSignInResponseWithNormalParams(t *testing.T) {
	emailfield := auth.EmailField{Value: email}
	passwordfield := auth.PasswordField{Value: []byte(password)}
	signinform := auth.SignInForm{EmailField: emailfield, PasswordField: passwordfield}
	authClient := auth.AuthClient{ApiKey: os.Getenv("STAGING_FIREBASE_API_KEY")}
	accessToken, _ := auth.LoadAccessToken()
	wrapper := api.GetApiClient(accessToken, apiHost)
	apiClient := actions.AuthClientWrapper{AuthClient: authClient, ApiClient: wrapper}
	response, err := apiClient.AuthClient.SignIn(signinform)

	if err != nil {
		t.Error(err)
	}

	err = auth.HandleFirebaseSignInResponse(response)

	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(auth.FirebaseAuthFile); os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestHandleFirebaseSignInResponseWithBlankParams(t *testing.T) {
	email := ""
	password := ""
	emailfield := auth.EmailField{Value: email}
	passwordfield := auth.PasswordField{Value: []byte(password)}
	signinform := auth.SignInForm{EmailField: emailfield, PasswordField: passwordfield}
	authClient := auth.AuthClient{ApiKey: apiKey}
	accessToken, _ := auth.LoadAccessToken()
	wrapper := api.GetApiClient(accessToken, apiHost)
	apiClient := actions.AuthClientWrapper{AuthClient: authClient, ApiClient: wrapper}
	response, err := apiClient.AuthClient.SignIn(signinform)

	if err != nil {
		t.Error(err)
	}

	if err != nil {
		t.Error(err)
	} else if _, err := os.Stat(auth.FirebaseAuthFile); os.IsNotExist(err) {
		t.Error(err)
	}

	err = auth.HandleFirebaseSignInResponse(response)

	if err == nil {
		t.Errorf("sign in: %s == nil; want !=", err)
	}

}

// func TestLoadRefreshTokenWhileLoggedIn(t *testing.T) {
// 	authClient := auth.AuthClient{ApiKey: apiKey}
// 	emailfield := auth.EmailField{Value: email}
// 	passwordfield := auth.PasswordField{Value: []byte(password)}
// 	signinform := auth.SignInForm{EmailField: emailfield, PasswordField: passwordfield}
// 	response, err := authClient.SignIn(signinform)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	jsonresponse := make(map[string]string)
// 	err = json.Unmarshal(response.Body(), &jsonresponse)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	refreshToken := jsonresponse["refresh_token"]

// 	if len(refreshToken) == 0 {
// 		t.Error("empty refresh token")
// 	}

// 	response, err = authClient.ExchangeRefreshForIdToken(refreshToken)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	jsonresponse = make(map[string]string)
// 	err = json.Unmarshal(response.Body(), &jsonresponse)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	accessToken := jsonresponse["access_token"]

// 	if len(accessToken) == 0 {
// 		t.Error("empty access token")
// 	}

// 	deviceID, err := auth.LoadDeviceID()

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	wrapper := api.GetApiClient(accessToken, apiHost)
// 	apiClient := actions.AuthClientWrapper{AuthClient: authClient, ApiClient: wrapper}
// 	err = apiClient.Login(email, password, deviceID)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if _, err := os.Stat(auth.FirebaseAuthFile); os.IsNotExist(err) {
// 		t.Error(err)
// 	}

// 	refreshToken, err = auth.LoadRefreshToken()

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if len(refreshToken) == 0 {
// 		t.Error("failed to load refresh token")
// 	}
// }

func TestLoadRefreshTokenWhileLoggedOut(t *testing.T) {
	err := logout()

	if err != nil {
		t.Error(err)
	}

	refreshToken, err := auth.LoadRefreshToken()

	if err == nil {
		t.Error(err)
	}

	if len(refreshToken) > 0 {
		t.Error("Non empty refresh token")
	}
}

func TestIsRefreshTokenExistsWhileLoggedOut(t *testing.T) {
	err := logout()

	if err != nil {
		t.Error(err)
	}

	if auth.IsRefreshTokenExists() {
		t.Error("refresh token exists")
	}

}

func TestIsDeviceCreatedWhileLoggedOut(t *testing.T) {
	os.Remove(auth.DeviceFile)

	if auth.IsDeviceCreated() {
		t.Errorf("device exists: %s", auth.DeviceFile)
	}

}

func TestIsAuthenticatedWhileLoggedOut(t *testing.T) {
	err := logout()

	if err != nil {
		t.Error(err)
	}

	auth := auth.IsAuthenticated()

	if auth {
		t.Error("auth.IsAuthenticated() == true; want false")
	}
}

func TestEmailFieldValidateWithWrongValue(t *testing.T) {
	emailfield := auth.EmailField{Value: "wrongemail.com"}

	if emailfield.Validate() == nil {
		t.Error("emailfield.Validate() == nil; want error")
	}
}

func TestEmailFieldValidateWithRightValue(t *testing.T) {
	emailfield := auth.EmailField{Value: email}

	if emailfield.Validate() != nil {
		t.Error("emailfield.Validate() == error; want nil")
	}
}

func TestPasswordFieldValidateWithWrongValue(t *testing.T) {
	passwordfield := auth.PasswordField{Value: []byte("12345")}

	if passwordfield.Validate() == nil {
		t.Error("passwordfield.Validate() == nil; want error")
	}
}

func TestPasswordFieldValidateWithRightValue(t *testing.T) {
	passwordfield := auth.PasswordField{Value: []byte(password)}

	if passwordfield.Validate() != nil {
		t.Error("passwordfield.Validate() == error; want nil")
	}
}

func TestValidatePasswordConfirmationWhileMatch(t *testing.T) {
	emailfield := auth.EmailField{Value: email}
	passwordfield := auth.PasswordField{Value: []byte(password)}
	confirmation := auth.PasswordConfirmationField{Value: []byte(password)}
	signinform := auth.SignInForm{EmailField: emailfield, PasswordField: passwordfield}
	signupform := auth.SignUpForm{SignInForm: signinform, PasswordConfirmationField: confirmation}
	err := signupform.ValidatePasswordConfirmation()

	if err != nil {
		t.Error(err)
	}
}

func TestValidatePasswordConfirmationWhileNotMatch(t *testing.T) {
	emailfield := auth.EmailField{Value: email}
	passwordfield := auth.PasswordField{Value: []byte(password)}
	confirmation := auth.PasswordConfirmationField{Value: []byte("otherpass")}
	signinform := auth.SignInForm{EmailField: emailfield, PasswordField: passwordfield}
	signupform := auth.SignUpForm{SignInForm: signinform, PasswordConfirmationField: confirmation}
	err := signupform.ValidatePasswordConfirmation()

	if err == nil {
		t.Error("signupform.ValidatePasswordConfirmation() == nil; want error")
	}
}
