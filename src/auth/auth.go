// auth is a package containing an authentication client built around Firebase REST API.
//
// See https://firebase.google.com/docs/reference/rest for more information.
package auth

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/go-resty/resty/v2"
)

// Client is a REST client for Go.
//
// See https://github.com/go-resty/resty for more information.
var Client = resty.New()

// AuthClient is a structure used as a Firebase REST client.
//
// See https://firebase.google.com/docs/reference/rest for more information.
type AuthClient struct {
	ApiKey string
}

// signInUpRequestBody is a structure that is used as a data holder for both Firebase sign in and sign up requests.
type signInUpRequestBody struct {
	Email             string
	Password          string
	ReturnSecureToken bool
}

// SignUp is a method to perform a Firebase sign up request. It accepts an instance of a SignUpForm that holds validated data for request.
//
// See https://firebase.google.com/docs/reference/rest/auth#section-create-email-password for more information.
func (c AuthClient) SignUp(form SignUpForm) (*resty.Response, error) {
	url := "https://identitytoolkit.googleapis.com/v1/accounts:signUp"
	body := signInUpRequestBody{Email: form.EmailField.Value, Password: string(form.PasswordField.Value), ReturnSecureToken: true}
	request := make(map[string]any)
	request["email"] = body.Email
	request["password"] = body.Password
	request["returnSecureToken"] = body.ReturnSecureToken
	jsonRequest, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	return Client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"key": c.ApiKey,
		}).
		SetBody(jsonRequest).
		Post(url)
}

// SignIn is a method to perform a Firebase sign in request. It accepts an instance of a SignInForm that holds validated data for request.
//
// See https://firebase.google.com/docs/reference/rest/auth#section-sign-in-email-password for more information.
func (c AuthClient) SignIn(form SignInForm) (*resty.Response, error) {
	url := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword"
	body := signInUpRequestBody{Email: form.EmailField.Value, Password: string(form.PasswordField.Value), ReturnSecureToken: true}
	request := make(map[string]any)
	request["email"] = body.Email
	request["password"] = body.Password
	request["returnSecureToken"] = body.ReturnSecureToken
	jsonRequest, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	return Client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"key": c.ApiKey,
		}).
		SetBody(jsonRequest).
		Post(url)
}

// See https://firebase.google.com/docs/reference/rest/auth#section-refresh-token for more information.
func (c AuthClient) ExchangeRefreshForIdToken(refreshToken string) (*resty.Response, error) {
	url := "https://securetoken.googleapis.com/v1/token"
	body := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", refreshToken)
	Client.SetTimeout(time.Duration(1 * time.Second))

	return Client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParams(map[string]string{
			"key": c.ApiKey,
		}).
		SetBody(body).
		Post(url)
}

// GetAccessToken is a method to obtain a new access token from Firebase REST API and dump the response into FirebaseAuthFile.
func (c AuthClient) GetAccessToken(refreshToken string) (*resty.Response, error) {
	response, err := c.ExchangeRefreshForIdToken(refreshToken)

	if err != nil {
		return nil, err
	}

	err = JsonDump(response.Body(), FirebaseAuthFile)
	return response, err
}

// GetUserInfo is used to check if the user already exists in the Firebase database during the registration.
//
// See https://firebase.google.com/docs/reference/rest/auth#section-get-account-info for more information.
func (c AuthClient) GetUserData(email string) (*resty.Response, error) {
	url := "https://identitytoolkit.googleapis.com/v1/accounts:createAuthUri"
	request := make(map[string]any)
	request["identifier"] = email
	request["continueUri"] = "http://localhost:8080/app"
	body, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	return Client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"key": c.ApiKey,
		}).
		SetBody(body).
		Post(url)
}
