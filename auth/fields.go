// fields is a package containing structures for storing user prompted data such as email or password.
package auth

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

// PasswordConfirmationField is used in the SignUpForm to store the confirmation of a password.
type PasswordConfirmationField struct {
	Value []byte
}

// Validate checks if the EmailField.Value is a valid email address.
func (f EmailField) Validate() error {
	if len(f.Value) < 5 || !strings.Contains(f.Value, "@") || !strings.Contains(f.Value, ".") {
		return errors.New("invalid email address")
	}
	return nil
}

// Validate checks if the PasswordField.Value is a valid password.
func (p PasswordField) Validate() error {
	if len(p.Value) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}
