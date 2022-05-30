package forms

import (
	"bytes"
	"errors"
	"strings"
)

type form interface {
	correctEmailAddress() bool
	passwordConfirmationMatch() bool
}

type SignUpForm struct {
	Email                string
	Password             []byte
	PasswordConfirmation []byte
}

func (s SignUpForm) passwordConfirmationMatch() bool {
	return bytes.Equal(s.Password, s.PasswordConfirmation)
}

func (s SignUpForm) correctEmailAddress() bool {
	if len(s.Email) > 5 && strings.Index(s.Email, "@") > 0 {
		return true
	}
	return false
}

func validateForm(f form) error {
	if !f.correctEmailAddress() {
		return errors.New("incorrect email address")
	}

	if !f.passwordConfirmationMatch() {
		return errors.New("password confirmation doesn't match")
	}
	return nil
}

func IsValidForm(f form) (bool, error) {
	err := validateForm(f)

	if err != nil {
		return false, err
	}
	return true, err
}
