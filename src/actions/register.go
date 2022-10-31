package actions

import (
	"errors"
	"fmt"

	"github.com/forestvpn/cli/auth"
	"golang.org/x/term"
)

// Register is a method to perform a user registration on Firebase.
// See https://firebase.google.com/docs/reference/rest/auth#section-create-email-password for more information.
func (w AuthClientWrapper) Register(email string, password string) error {
	signinform := auth.SignInForm{}
	emailfield, err := auth.GetEmailField(email)

	if err != nil {
		return err
	}

	signinform.EmailField = emailfield

	if err != nil {
		return err
	}

	signinform.PasswordField.Value = []byte("12345678")
	response, err := w.AuthClient.SignIn(signinform)

	if err != nil {
		return err
	}

	message, err := auth.HandleFirebaseAuthResponse(response)

	if err != nil {
		return err
	}

	if message == "EMAIL_NOT_FOUND" {
		validate := true
		passwordfield, err := auth.GetPasswordField([]byte(password), validate)

		if err != nil {
			return err
		}

		signinform.PasswordField = passwordfield
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

		response, err := w.AuthClient.SignUp(signupform)

		if err != nil {
			return err
		}

		err = auth.HandleFirebaseSignUpResponse(response)

		if err != nil {
			return err
		}

		user_id, err := w.SetUpProfile(response)

		if err != nil {
			return err
		}

		err = auth.SetActiveProfile(user_id)

		if err != nil {
			return err
		}

		return w.AccountsMap.AddAccount(signinform.EmailField.Value, user_id)

	}

	return errors.New("a profile for this user already exists")

}
