// auth is a package containing an authentication client built around Firebase REST API.
// See https://firebase.google.com/docs/reference/rest for more information.
package auth

import (
	"fmt"

	"encoding/json"

	"github.com/go-resty/resty/v2"
)

// CLient is a REST client for Go.
// See https://github.com/go-resty/resty for more information.
var Client = resty.New()

// AuthClient is a structure used as a Firebase REST client.
type AuthClient struct {
	ApiKey string
}

// signInUpRequestBody is a structure that is used as a data holder for both SignIn and SignUp requests.
type signInUpRequestBody struct {
	Email             string
	Password          string
	ReturnSecureToken bool
}

// SignUp performs a Firebase sign up request.
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

// SignIn performs a Firebase sign in request.
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
func (c AuthClient) ExchangeRefreshForIdToken() (*resty.Response, error) {
	url := "https://securetoken.googleapis.com/v1/token"
	refreshToken, err := LoadRefreshToken()

	if err != nil {
		return nil, err
	}

	body := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", refreshToken)

	return Client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParams(map[string]string{
			"key": c.ApiKey,
		}).
		SetBody(body).
		Post(url)
}

// GetAccessToken calls ExchangeRefreshForIdToken and dumps the response into FirebaseAuthFile.
func (c AuthClient) GetAccessToken() (*resty.Response, error) {
	response, err := c.ExchangeRefreshForIdToken()

	if err != nil {
		return nil, err
	}

	err = JsonDump(response.Body(), FirebaseAuthFile)
	return response, err
}
