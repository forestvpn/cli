package auth

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

var reader = bufio.NewReader(os.Stdin)

type SignInForm struct {
	EmailField
	PasswordField
}

type SignUpForm struct {
	SignInForm
	PasswordConfirmationField
}

type Info struct {
	AdditionalProperties string
}

type InfoForm struct {
	Type string
	Info Info
}

// Compares *forest.auth.forms.SignUpForm.SignInForm.PasswordField and *forest.auth.forms.SignUpForm.PasswordConfirmationField
func (s SignUpForm) ValidatePasswordConfirmation() error {
	if !bytes.Equal(s.PasswordField.Value, s.PasswordConfirmationField.Value) {
		return errors.New("password confirmation doesn't match")
	}
	return nil
}

func getPasswordField(password []byte) (PasswordField, error) {
	passwordfield := PasswordField{Value: password}

	for !(len(passwordfield.Value) > 0) {
		fmt.Print("Enter password: ")
		password, err := term.ReadPassword(0)
		fmt.Println()

		if err != nil {
			return passwordfield, err
		}

		passwordfield.Value = password
	}

	err := passwordfield.Validate()

	if err != nil {
		return passwordfield, err
	}
	return passwordfield, nil
}

func GetEmailField(email string) (EmailField, error) {
	emailfield := EmailField{Value: email}

	for !(len(emailfield.Value) > 0) {
		fmt.Print("Enter email: ")
		email, err := reader.ReadString('\n')

		if err != nil {
			return emailfield, err
		}

		emailfield.Value = email[:len(email)-1]
	}

	err := emailfield.Validate()
	return emailfield, err
}

// Prompts user and fills the *forest.auth.forms.SignInForm with *forest.auth.forms.fields.EmailField and *forest.auth.forms.fields.PasswordField
func GetSignInForm(email string, password []byte) (SignInForm, error) {
	signinform := SignInForm{}
	emailfield, err := GetEmailField(email)

	if err != nil {
		return signinform, err
	}

	signinform.EmailField = emailfield
	passwordfield, err := getPasswordField(password)

	if err != nil {
		return signinform, err
	}

	signinform.PasswordField = passwordfield
	return signinform, err
}
