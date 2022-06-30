package actions

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"golang.org/x/term"

	"github.com/forestvpn/cli/api"
	"github.com/forestvpn/cli/auth"
)

func Register(email string, password string) error {
	signinform, err := auth.GetSignInForm(email, []byte(password))

	if err != nil {
		return err
	}

	signupform := auth.SignUpForm{}
	fmt.Print("Confirm password: ")
	passwordConfirmation, err := term.ReadPassword(0)
	fmt.Println()

	if err != nil {
		return err
	}

	signupform.PasswordConfirmationField.Value = passwordConfirmation
	signupform.SignInForm = signinform
	err = signupform.ValidatePasswordConfirmation()

	if err != nil {
		return err
	}

	response, err := auth.SignUp(signupform)

	if err != nil {
		return err
	}

	err = auth.HandleFirebaseAuthResponse(response)

	if err == nil {
		color.Green("Signed up")
	}

	return err
}

func Login(email string, password string) error {
	if !auth.IsRefreshTokenExists() {
		signinform, err := auth.GetSignInForm(email, []byte(password))

		if err != nil {
			return err
		}

		response, err := auth.SignIn(signinform)

		if err != nil {
			return err
		}

		err = auth.HandleFirebaseSignInResponse(response)

		if err != nil {
			return err
		}

		err = auth.JsonDump(response.Body(), auth.FirebaseAuthFile)

		if err != nil {
			return err
		}

		response, err = auth.GetAccessToken()

		if err != nil {
			return err
		}

		err = auth.JsonDump(response.Body(), auth.FirebaseAuthFile)

		if err != nil {
			return err
		}
	}

	if !auth.IsDeviceCreated() {
		accessToken, err := auth.LoadAccessToken()

		if err != nil {
			return err
		}

		resp, err := api.CreateDevice(accessToken)

		if err != nil {
			return err
		}

		b, err := json.MarshalIndent(resp, "", "    ")

		if err != nil {
			return err
		}

		err = auth.JsonDump(b, auth.DeviceFile)

		if err != nil {
			return err
		}

		return err

	}

	color.Green("Signed in")

	return nil
}

func Logout() error {
	err := os.Remove(auth.FirebaseAuthFile)

	if err == nil {
		color.Red("Signed out")
	}

	return err
}
