package api

import (
	"forest/auth/forms"
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
	body := signInUpRequestBody{Email: form.EmailField.Value, Password: string(form.PasswordField.Value), ReturnSecureToken: true}
	req := make(map[string]any)
	req["email"] = body.Email
	req["password"] = body.Password
	req["returnSecureToken"] = body.ReturnSecureToken
	jsonRequest, err := json.Marshal(req)

	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"key": firebaseApiKey,
		}).
		SetBody(jsonRequest).
		Post("https://identitytoolkit.googleapis.com/v1/accounts:signUp")

	return resp, err
}

func SignIn(firebaseApiKey string, form forms.SignInForm) (*resty.Response, error) {
	body := signInUpRequestBody{Email: form.EmailField.Value, Password: string(form.PasswordField.Value), ReturnSecureToken: true}
	req := make(map[string]any)
	req["email"] = body.Email
	req["password"] = body.Password
	req["returnSecureToken"] = body.ReturnSecureToken
	jsonRequest, err := json.Marshal(req)

	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"key": firebaseApiKey,
		}).
		SetBody(jsonRequest).
		Post("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword")

	return resp, err
}
