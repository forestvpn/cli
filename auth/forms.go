// forms is package contaning structures to work with user input during Firebase authentication process.
package auth

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

// SignInForm is used to store user's email and password.
type SignInForm struct {
	EmailField
	PasswordField
}

// SignUpForm is carries the SignInForm and the password confirmation
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

// ValidatePasswordConfirmation compares password and the password confirmation values.
func (s SignUpForm) ValidatePasswordConfirmation() error {
	if !bytes.Equal(s.PasswordField.Value, s.PasswordConfirmationField.Value) {
		return errors.New("password confirmation doesn't match")
	}
	return nil
}

// getPasswordField prompts the user a password and then validates it.
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

// GetEmailField prompts a user an email and then validates it.
func GetEmailField(email string) (EmailField, error) {
	var reader = bufio.NewReader(os.Stdin)
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

// Prompts user both email and password and returns the SignInForm.
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
