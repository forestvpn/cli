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
var ApiURL = os.Getenv("API_URL")

type signInUpRequestBody struct {
	Email             string
	Password          string
	ReturnSecureToken bool
}

func SignUp(firebaseApiKey string, form forms.SignUpForm) (*resty.Response, error) {
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
			"key": firebaseApiKey,
		}).
		SetBody(jsonRequest).
		Post(url)
}

func SignIn(firebaseApiKey string, form forms.SignInForm) (*resty.Response, error) {
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
			"key": firebaseApiKey,
		}).
		SetBody(jsonRequest).
		Post(url)
}

func ExchangeRefreshForIdToken(firebaseApiKey string) (*resty.Response, error) {
	url := "https://securetoken.googleapis.com/v1/token"
	refreshToken, err := utils.LoadRefreshToken()

	if err != nil {
		return nil, err
	}

	body := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", refreshToken)

	return client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryParams(map[string]string{
			"key": firebaseApiKey,
		}).
		SetBody(body).
		Post(url)
}

func GetAccessToken(firebaseApiKey string) (*resty.Response, error) {
	response, err := ExchangeRefreshForIdToken(firebaseApiKey)

	if err != nil {
		return nil, err
	}

	err = utils.JsonDump(response.Body(), utils.FirebaseAuthFile)
	return response, err
}
