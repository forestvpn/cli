package forms

import (
	"errors"
	"strings"
)

type EmailField struct {
	Value string
}

type PasswordField struct {
	Value []byte
}

// Password confirmation is used in the *forest.auth.forms.SignUpForm
type PasswordConfirmationField struct {
	Value []byte
}

func (f EmailField) Validate() error {
	if len(f.Value) < 5 || !strings.Contains(f.Value, "@") || !strings.Contains(f.Value, ".") {
		return errors.New("invalid email address")
	}
	return nil
}

func (p PasswordField) Validate() error {
	if len(p.Value) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}
