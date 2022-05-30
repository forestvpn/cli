package api

import (
	"forest/forms"

	"encoding/json"

	"github.com/go-resty/resty/v2"
)

var client = resty.New()

type signUpRequestBody struct {
	Email             string
	Password          string
	returnSecureToken bool
}

func SignUp(firebaseApiKey string, form forms.SignUpForm) (*resty.Response, error) {
	body := signUpRequestBody{form.Email, string(form.Password), true}
	req, err := json.Marshal(body)

	if err != nil {
		return nil, err
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"key": firebaseApiKey,
		}).
		SetBody(req).
		Post("https://identitytoolkit.googleapis.com/v1/accounts:signUp")

	return resp, err
}
