package auth

import (
	"fmt"
	"forest/auth/forms"
	"forest/utils"
	"os"

	"encoding/json"

	"github.com/go-resty/resty/v2"
)

var client = resty.New()
var FirebaseApiKey = os.Getenv("STAGING_FIREBASE_API_KEY")

type signInUpRequestBody struct {
	Email             string
	Password          string
	ReturnSecureToken bool
}

func SignUp(form forms.SignUpForm) (*resty.Response, error) {
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

	return client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"key": FirebaseApiKey,
		}).
		SetBody(jsonRequest).
		Post(url)
}

func SignIn(form forms.SignInForm) (*resty.Response, error) {
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

	return client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"key": FirebaseApiKey,
		}).
		SetBody(jsonRequest).
		Post(url)
}

func ExchangeRefreshForIdToken() (*resty.Response, error) {
	url := "https://securetoken.googleapis.com/v1/token"
	refreshToken, err := utils.LoadRefreshToken()

	if err != nil {
		return nil, err
	}

	body := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", refreshToken)

	return client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParams(map[string]string{
			"key": FirebaseApiKey,
		}).
		SetBody(body).
		Post(url)
}

func GetAccessToken() (*resty.Response, error) {
	response, err := ExchangeRefreshForIdToken()

	if err != nil {
		return nil, err
	}

	err = utils.JsonDump(response.Body(), utils.FirebaseAuthFile)
	return response, err
}
