package auth

import (
	"fmt"

	"encoding/json"

	"github.com/go-resty/resty/v2"
)

var Client = resty.New()

type AuthClient struct {
	ApiKey string
}

type signInUpRequestBody struct {
	Email             string
	Password          string
	ReturnSecureToken bool
}

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

func (c AuthClient) GetAccessToken() (*resty.Response, error) {
	response, err := c.ExchangeRefreshForIdToken()

	if err != nil {
		return nil, err
	}

	err = JsonDump(response.Body(), FirebaseAuthFile)
	return response, err
}
