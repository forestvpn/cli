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

var reader = bufio.NewReader(os.Stdin)

type SignInForm struct {
	Email    string
	Password []byte
}

type SignUpForm struct {
	SignInForm
	PasswordConfirmation []byte
}

func (s SignInForm) PromptEmail() error {
	fmt.Print("Enter email: ")
	email, err := reader.ReadString('\n')

	if err == nil {
		s.Email = email[:len(email)-1]
	}
	return err
}

func (s SignInForm) PromptPassword() error {
	fmt.Print("Enter password: ")
	password, err := term.ReadPassword(0)
	fmt.Println()

	if err == nil {
		s.Password = password
	}
	return err

}

func (s SignUpForm) PromptPasswordConfirmation() error {
	fmt.Print("Confirm password: ")
	password, err := term.ReadPassword(0)
	fmt.Println()

	if err == nil {
		s.PasswordConfirmation = password
	}
	return err
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
	if len(s.Email) > 5 && strings.Index(s.Email, "@") > 0 {
		return errors.New("invalid email address")
	}
	return nil
}
