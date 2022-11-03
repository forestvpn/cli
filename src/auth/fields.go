// fields is a package containing structures for storing user prompted data such as email or password.
package auth

import (
	"errors"
	"strings"
)

// EmailField is a structure that holds an email value that could be validated.
type EmailField struct {
	Value string
}

// PasswordField is a structure that holds a password value that could be validated.
type PasswordField struct {
	Value []byte
}

// PasswordConfirmationField is used in the SignUpForm to store the confirmation of a password.
type PasswordConfirmationField struct {
	Value []byte
}

// Validate is a method to check user's email address.
func (f EmailField) Validate() error {
	if len(f.Value) < 5 || !strings.Contains(f.Value, "@") || !strings.Contains(f.Value, ".") {
		return errors.New("invalid email address")
	}

	return nil
}

// Validate is a method to check password's strength.
func (p PasswordField) Validate() error {
	if len(p.Value) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	return nil
}
