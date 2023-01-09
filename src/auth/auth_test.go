package auth_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/forestvpn/cli/actions"
	"github.com/forestvpn/cli/auth"
	"github.com/forestvpn/cli/utils"
)

const filepath = "/tmp/test.json"

func logout() error {
	userID, err := auth.LoadUserID()

	if err != nil || len(userID) == 0 {
		return nil
	}

	auth.RemoveFirebaseAuthFile(userID)
	auth.RemoveActiveUserLockFile()
	m := auth.GetAccountsMap(auth.AccountsMapFile)
	m.RemoveAccount(userID)
	return nil
}

func login(email string, password string) (actions.AuthClientWrapper, error) {
	client := actions.AuthClientWrapper{}
	err := auth.Init()

	if err != nil {
		return client, err
	}

	client, err = actions.GetAuthClientWrapper(utils.ApiHost, utils.FirebaseApiKey)

	if err != nil {
		return client, err
	}

	err = client.Login(email, password)

	if err != nil {
		return client, err
	}

	return actions.GetAuthClientWrapper(utils.ApiHost, utils.FirebaseApiKey)
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

func TestAccountMap(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	accountsmap := auth.GetAccountsMap(auth.AccountsMapFile)
	v := accountsmap.GetEmail(userID)

	if len(v) == 0 {
		t.Errorf("%s not found in local accounts", email)
	}

	err = accountsmap.RemoveAccount(userID)

	if err != nil {
		t.Error(err)
	}

	v = accountsmap.GetEmail(email)

	if len(v) != 0 {
		t.Errorf("%s is not removed from local accounts", email)
	}
}

func TestLoadAccessTokenWhileLoggedIn(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	accessToken, err := auth.LoadAccessToken(userID)

	if err != nil {
		t.Error(err)
	}

	if len(accessToken) == 0 {
		t.Error("access token not loaded while logged in")
	}
}

func TestLoadAccessTokenWhileLoggedOut(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	err = logout()

	if err != nil {
		t.Error(err)
	}

	accessToken, err := auth.LoadAccessToken(userID)

	if err == nil || len(accessToken) > 0 {
		t.Error(err)
	}
}

func TestHandleFirebaseSignInResponseWithNormalParams(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	authclient := auth.AuthClient{ApiKey: utils.FirebaseApiKey}
	emailfield := auth.EmailField{Value: email}
	passwordfield := auth.PasswordField{Value: password}
	signinform := auth.SignInForm{EmailField: emailfield, PasswordField: passwordfield}
	response, err := authclient.SignIn(signinform)

	if err != nil {
		t.Error(err)
	}

	err = auth.HandleFirebaseSignInResponse(response)

	if err != nil {
		t.Error(err)
	}

}

func TestHandleFirebaseSignInResponseWithBlankParams(t *testing.T) {
	email := ""
	password := ""
	authclient := auth.AuthClient{ApiKey: os.Getenv("STAGING_FIREBASE_API_KEY")}
	emailfield := auth.EmailField{Value: email}
	passwordfield := auth.PasswordField{Value: password}
	signinform := auth.SignInForm{EmailField: emailfield, PasswordField: passwordfield}
	response, err := authclient.SignIn(signinform)

	if err != nil {
		t.Error(err)
	}

	if err != nil {
		t.Error(err)
	}

	err = auth.HandleFirebaseSignInResponse(response)

	if err == nil {
		t.Errorf("sign in: %s == nil; want !=", err)
	}

}

func TestLoadRefreshTokenWhileLoggedIn(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	refreshToken, err := auth.LoadRefreshToken()

	if err != nil {
		t.Error(err)
	}

	if len(refreshToken) == 0 {
		t.Error("failed to load refresh token")
	}
}

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

	exists, _ := auth.IsRefreshTokenExists()

	if exists {
		t.Error("refresh token exists")
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

func TestEmailFieldValidateWithIncorrectValue(t *testing.T) {
	email := ""
	emailfield := auth.EmailField{Value: email}

	if emailfield.Validate() == nil {
		t.Error("emailfield.Validate() == nil; want error")
	}
}

func TestEmailFieldValidateWithCorrectValue(t *testing.T) {
	email := "x@x.xx"
	emailfield := auth.EmailField{Value: email}

	if emailfield.Validate() != nil {
		t.Error("emailfield.Validate() == error; want nil")
	}
}

func TestPasswordFieldValidateWithIncorrectValue(t *testing.T) {
	passwordfield := auth.PasswordField{Value: "12345"}

	if passwordfield.Validate() == nil {
		t.Error("passwordfield.Validate() == nil; want error")
	}
}

func TestPasswordFieldValidateWithRightValue(t *testing.T) {
	passwordfield := auth.PasswordField{Value: "123456"}

	if passwordfield.Validate() != nil {
		t.Error("passwordfield.Validate() == error; want nil")
	}
}

func TestValidatePasswordConfirmationWhileMatch(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	emailfield := auth.EmailField{Value: email}
	passwordfield := auth.PasswordField{Value: password}
	confirmation := auth.PasswordConfirmationField{Value: password}
	signinform := auth.SignInForm{EmailField: emailfield, PasswordField: passwordfield}
	signupform := auth.SignUpForm{SignInForm: signinform, PasswordConfirmationField: confirmation}
	err := signupform.ValidatePasswordConfirmation()

	if err != nil {
		t.Error(err)
	}
}

func TestValidatePasswordConfirmationWhileNotMatch(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	emailfield := auth.EmailField{Value: email}
	passwordfield := auth.PasswordField{Value: password}
	confirmation := auth.PasswordConfirmationField{Value: "otherpass"}
	signinform := auth.SignInForm{EmailField: emailfield, PasswordField: passwordfield}
	signupform := auth.SignUpForm{SignInForm: signinform, PasswordConfirmationField: confirmation}
	err := signupform.ValidatePasswordConfirmation()

	if err == nil {
		t.Error("signupform.ValidatePasswordConfirmation() == nil; want error")
	}
}

func TestBillingFeatureExpired(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	client, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	billingFeatures, err := client.ApiClient.GetBillingFeatures()

	if err != nil {
		t.Error(err)
	}

	data, err := json.Marshal(billingFeatures)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	path := auth.ProfilesDir + userID + auth.BillingFeatureFile
	err = auth.JsonDump(data, path)

	if err != nil {
		t.Error(err)
	}

	billingFeature, err := client.GetUnexpiredOrMostRecentBillingFeature(userID)

	if err != nil {
		t.Error(err)
	}

	now := time.Now()
	expiryDate := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, now.Minute(), now.Second(), now.Nanosecond(), now.Location())
	billingFeature.SetExpiryDate(expiryDate)

	if !auth.BillingFeatureExpired(billingFeature) {
		t.Error("billing feature is not expired; want expired")
	}
}

func TestBillingFeautureExists(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	if !auth.BillingFeautureExists(userID) {
		t.Error("billing feature file not found after login.")
	}
}

func TestLoadBillingFeatures(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	billingFeatures, err := auth.LoadBillingFeatures(userID)

	if err != nil {
		t.Error(err)
	}

	if len(billingFeatures) == 0 {
		t.Error("billing features not loaded while logged in")
	}

}

func TestDumpAccessTokenExpireDate(t *testing.T) {
	var expiresIn string
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	data, err := auth.LoadFirebaseAuthFile(userID)

	if err != nil {
		t.Error(err)
	}

	var y interface{} = data["expires_in"]
	switch v := y.(type) {
	case string:
		expiresIn = v
	}

	expireDate, err := auth.GetAccessTokenExpireDate(expiresIn)

	if err != nil {
		t.Error(err)
	}

	err = auth.DumpAccessTokenExpireDate(userID, expiresIn)

	if err != nil {
		t.Error(err)
	}

	expireDateLoaded, err := auth.LoadAccessTokenExpireDate(userID)

	if err != nil {
		t.Error(err)
	}

	if expireDateLoaded.Year() != expireDate.Year() || expireDateLoaded.Month() != expireDate.Month() || expireDateLoaded.Day() != expireDate.Day() || expireDateLoaded.Hour() != expireDate.Hour() || expireDateLoaded.Minute() != expireDate.Minute() {
		t.Error("wrong access token expire date dump")
	}
}

func TestIsAccessTokenExpiredWithExpiredToken(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	now := time.Now()
	expiryDate := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-1, now.Minute(), now.Second(), now.Nanosecond(), now.Location())
	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	b, err := auth.Date2Json(expiryDate)

	if err != nil {
		t.Error(err)
	}

	path := auth.ProfilesDir + userID + auth.FirebaseExtensionFile
	err = auth.JsonDump(b, path)

	if err != nil {
		t.Error(err)
	}

	expired, err := auth.IsAccessTokenExpired(userID)

	if err != nil {
		t.Error(err)
	}

	if !expired {
		t.Error("access token not expired; want expired")
	}

}

func TestIsAccessTokenExpiredWithUnExpiredToken(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	now := time.Now()
	expiryDate := time.Date(now.Year()+1, now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	b, err := auth.Date2Json(expiryDate)

	if err != nil {
		t.Error(err)
	}

	path := auth.ProfilesDir + userID + auth.FirebaseExtensionFile
	err = auth.JsonDump(b, path)

	if err != nil {
		t.Error(err)
	}

	expired, err := auth.IsAccessTokenExpired(userID)

	if err != nil {
		t.Error(err)
	}

	if expired {
		t.Error("access token is expired; want not expired")
	}

}

func TestGetAccessTokenExpireDate(t *testing.T) {
	var seconds int
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	data, err := auth.LoadFirebaseAuthFile(userID)

	if err != nil {
		t.Error(err)
	}

	var z interface{} = data["expires_in"]
	switch v := z.(type) {
	case string:
		seconds, err = strconv.Atoi(v)

		if err != nil {
			t.Error(err)
		}
	}

	now := time.Now()
	expireDate, err := auth.GetAccessTokenExpireDate(fmt.Sprint(seconds))

	if err != nil {
		t.Error(err)
	}

	expireDate1 := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+seconds, now.Nanosecond(), now.Location())

	// t.Error(expireDate.String(), expireDate1.String())

	if expireDate.Year() != expireDate1.Year() || expireDate.Month() != expireDate1.Month() || expireDate.Day() != expireDate1.Day() || expireDate.Hour() != expireDate1.Hour() || expireDate.Minute() != expireDate1.Minute() {
		t.Error("wrong expire date")
	}
}

func TestLoadIdTokenWhileLoggedIn(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	idToken, err := auth.LoadIdToken(userID)

	if err != nil {
		t.Error(err)
	}

	if len(idToken) == 0 {
		t.Error("empty id token")
	}
}

func TestLoadIdTokenWhileLoggedOut(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	err = logout()

	if err != nil {
		t.Error(err)
	}

	idToken, err := auth.LoadIdToken(userID)

	if err == nil {
		t.Error("error is not nil")
	}

	if len(idToken) > 0 {
		t.Error("not empty id token")
	}

}

// func TestUpdateProfileDevice(t *testing.T) {
// 	email := "x@x.xx"
// 	password := "123456"
// 	client, err := login(email, password)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	locations, err := client.ApiClient.GetLocations()

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	location := locations[rand.Intn(len(locations))]
// 	userID, err := auth.LoadUserID()

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	device, err := auth.LoadDevice(userID)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	device, err = client.ApiClient.UpdateDevice(device.GetId(), location.GetId())

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	err = auth.UpdateProfileDevice(device)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	device, err = auth.LoadDevice(userID)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if device.Location.GetId() != location.GetId() {
// 		t.Error("wrong device id")
// 	}
// }

func TestRemoveActiveUserLockFile(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	path := auth.ProfilesDir + userID + auth.ActiveUserLockFile
	_, err = os.Stat(path)

	if os.IsNotExist(err) {
		t.Error(err)
	}

	err = auth.RemoveActiveUserLockFile()

	if err != nil {
		t.Error(err)
	}

	if os.IsExist(err) {
		t.Error(err)
	}

}

func TestRemoveFirebaseAuthFile(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	path := auth.ProfilesDir + userID + auth.FirebaseAuthFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error(err)
	}

	err = auth.RemoveFirebaseAuthFile(userID)

	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(path); os.IsExist(err) {
		t.Error(err)
	}
}

func TestSetActiveProfile(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	email = "z@z.zz"
	_, err = login(email, password)

	if err != nil {
		t.Error(err)
	}

	if auth.IsActiveProfile(userID) {
		t.Errorf("logged out profile with uuid %s is active", userID)
	}

	inactiveUserId := userID
	userID, err = auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	if !auth.IsActiveProfile(userID) {
		t.Errorf("logged in profile with uuid %s is inactive", userID)
	}

	err = auth.SetActiveProfile(inactiveUserId)

	if err != nil {
		t.Error(err)
	}

	if !auth.IsActiveProfile(inactiveUserId) {
		t.Errorf("logged in profile with uuid %s is inactive", inactiveUserId)
	}
}

func TestIsActiveProfile(t *testing.T) {
	email := "x@x.xx"
	password := "123456"
	_, err := login(email, password)

	if err != nil {
		t.Error(err)
	}

	userID, err := auth.LoadUserID()

	if err != nil {
		t.Error(err)
	}

	if !auth.IsActiveProfile(userID) {
		t.Errorf("logged in profile with uuid %s is inactive", userID)
	}

	err = logout()

	if err != nil {
		t.Error(err)
	}

	if auth.IsActiveProfile(userID) {
		t.Errorf("logged out profile with uuid %s is active", userID)
	}

}
