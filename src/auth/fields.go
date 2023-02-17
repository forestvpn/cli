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
	Value string
}

// PasswordConfirmationField is used in the SignUpForm to store the confirmation of a password.
type PasswordConfirmationField struct {
	Value string
}

// Validate is a method to check user's email address.
func (f EmailField) Validate() error {
	if len(f.Value) < 5 {
		return errors.New("password must be at least 5 characters long")
	}

	if !strings.Contains(f.Value, "@") || !strings.Contains(f.Value, ".") {
		return errors.New("invalid email address")
	}

	return nil
}

// Validate is a method to check password's strength.
func (p PasswordField) Validate() error {
	if len(p.Value) < 5 {
		return errors.New("password must be at least 5 characters long")
	}

	if !strings.Contains(p.Value, "@") || !strings.Contains(p.Value, ".") {
		return errors.New("invalid email address")
	}

	return nil
}
