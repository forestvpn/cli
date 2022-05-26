package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/go-resty/resty/v2"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "forest",
		Usage:       "",
		Description: "ForestVPN client for Linux",
		Commands: []*cli.Command{
			{
				Name:    "signup",
				Aliases: []string{"s"},
				Usage:   "Sign up and use ForestVPN",
				Action: func(c *cli.Context) error {
					reader := bufio.NewReader(os.Stdin)
					var email string
					var password []byte
					var confirmation string

					for {
						fmt.Print("Enter email: ")
						input, _ := reader.ReadString('\n')

						if len(input) > 5 && strings.Index(input, "@") > 0 {
							email = input[:len(input)-1]
							break
						} else {
							fmt.Println("Please, enter a correct email")
						}
					}
					for {
						fmt.Print("Enter password: ")
						input, _ := term.ReadPassword(0)

						if len(input) > 6 {
							password = input
							break
						} else {
							fmt.Println("\nPassword should be at least 7 characters long")
						}

					}

					fmt.Println()

					for {
						fmt.Print("Confirm password: ")
						input, _ := term.ReadPassword(0)
						fmt.Println()

						if !bytes.Equal(password, input) {
							confirmation = string(input)
							break
						} else {
							return errors.New("password confirmation doesn't match")
						}
					}

					firebaseApiKey := os.Getenv("FB_API_KEY")
					client := resty.New()
					body, err := json.Marshal(map[string]any{"email": email, "password": confirmation, "returnSecureToken": true})

					if err != nil {
						return err
					}

					resp, err := client.R().
						SetHeader("Content-Type", "application/json").
						SetQueryParams(map[string]string{
							"key": firebaseApiKey,
						}).
						SetBody(body).
						Post("https://identitytoolkit.googleapis.com/v1/accounts:signUp")

					if err == nil {
						fmt.Print(resp)
					}
					return err
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
