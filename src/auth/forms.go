// forms is package contaning structures to work with user input during Firebase authentication process.
package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

// GetEmailField is a method that prompts a user an email and then validates it.
func GetEmailField(email string) (EmailField, error) {
	reader := bufio.NewReader(os.Stdin)
	emailfield := EmailField{Value: email}

	for len(emailfield.Value) == 0 {
		fmt.Print("Enter email: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return emailfield, err
		}

		input = strings.TrimSuffix(input, "\n")
		input = strings.TrimSuffix(input, "\r")
		emailfield.Value = input
	}

	err := emailfield.Validate()
	return emailfield, err
}
