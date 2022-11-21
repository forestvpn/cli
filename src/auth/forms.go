// forms is package contaning structures to work with user input during Firebase authentication process.
package auth

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// SignInForm is a structure to store user's email and password.
type SignInForm struct {
	EmailField
	PasswordField
}

// SignUpForm is a structure that holds the SignInForm and the password confirmation.
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

// ValidatePasswordConfirmation is a method that compares password and the password confirmation values.
func (s SignUpForm) ValidatePasswordConfirmation() error {
	if !bytes.Equal(s.PasswordField.Value, s.PasswordConfirmationField.Value) {
		return errors.New("password confirmation doesn't match")
	}
	return nil
}

// GetPasswordField is a method that prompts the user a password and then validates it.
// validate is a boolean that allows to enable or disable password validation.
// E.g. when password validation is needed on registration but not on login.
func GetPasswordField(password []byte, validate bool) (PasswordField, error) {
	passwordfield := PasswordField{Value: password}

	for !(len(passwordfield.Value) > 0) {
		fmt.Print("Enter password: ")
		password, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()

		if err != nil {
			return passwordfield, err
		}

		passwordfield.Value = password
	}

	if validate {
		err := passwordfield.Validate()

		if err != nil {
			return passwordfield, err
		}
	}

	return passwordfield, nil
}

// GetEmailField is a method that prompts a user an email and then validates it.
func GetEmailField(email string) (EmailField, error) {
	var reader = bufio.NewReader(os.Stdin)
	emailfield := EmailField{Value: email}

	for !(len(emailfield.Value) > 0) {
		fmt.Print("Enter email: ")
		email, err := reader.ReadString('\n')

		if err != nil {
			return emailfield, err
		}

		email = strings.TrimSuffix(email, "\n")
		email = strings.TrimSuffix(email, "\r")
		emailfield.Value = strings.TrimSuffix(email, "\n")
	}

	err := emailfield.Validate()
	return emailfield, err
}
