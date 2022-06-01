package forms

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

type SignInForm struct {
	Email    string
	Password []byte
}

type SignUpForm struct {
	SignInForm
	PasswordConfirmation []byte
}

func (s SignInForm) IsFilled() bool {
	return len(s.Email)+len(s.Password) > 0
}

func (s SignUpForm) ValidatePasswordConfirmation() error {
	if !bytes.Equal(s.Password, s.PasswordConfirmation) {
		return errors.New("password confirmation doesn't match")
	}
	return nil
}

func (s SignInForm) ValidatePassword() error {
	if len(s.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

func (s SignInForm) ValidateEmail() error {
	if len(s.Email) < 5 && !strings.Contains(s.Email, "@") {
		return errors.New("invalid email address")
	}
	return nil
}

func PromptSignInForm(email string, password string) (SignInForm, error) {
	signinform := SignInForm{Email: email, Password: []byte(password)}
	var reader = bufio.NewReader(os.Stdin)
	var err error

	for !signinform.IsFilled() {
		fmt.Print("Enter email: ")
		email, err := reader.ReadString('\n')

		if err != nil {
			return signinform, err
		}

		signinform.Email = email[:len(email)-1]
		err = signinform.ValidateEmail()

		if err != nil {
			return signinform, err
		}

		fmt.Print("Enter password: ")
		password, err := term.ReadPassword(0)
		fmt.Println()

		if err != nil {
			return signinform, err
		}

		signinform.Password = password

		err = signinform.ValidatePassword()

		if err != nil {
			return signinform, err
		}
	}
	return signinform, err
}
