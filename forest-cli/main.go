package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/go-resty/resty/v2"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "forest",
		Usage: "",
		Commands: []*cli.Command{
			{
				Name:    "signup",
				Aliases: []string{"s"},
				Usage:   "Sign up and use ForestVPN",
				Action: func(c *cli.Context) error {
					reader := bufio.NewReader(os.Stdin)
					var email string
					var password []byte

					for {
						fmt.Print("Enter email: ")
						input, _ := reader.ReadString('\n')

						if len(input) > 5 && strings.Index(input, "@") > 0 {
							email = input
							break
						} else {
							fmt.Println("Please, enter a correct email")
						}
					}
					for {
						fmt.Print("Enter password: ")
						input, _ := terminal.ReadPassword(0)

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
						input, _ := terminal.ReadPassword(0)
						fmt.Println()

						if bytes.Compare(password, input) != 0 {
							return errors.New("Passwords doesn't match")
						}
						break
					}

					firebaseAuthDomain := os.Getenv("FB_AUTH_DOMAIN")
					firebaseApiKey := os.Getenv("FB_API_KEY")
					client := resty.New()
					resp, err := client.R().
						SetHeader("Content-Type", "application/json").
						SetQueryParams(map[string]string{
							"key": firebaseApiKey,
						}).
						SetBody(fmt.Sprintf(`{"email":"%d", "password":"%d", returnSecureToken:true}`, email, password)).
						Post(firebaseAuthDomain)

					if err != nil {
						fmt.Print(resp)
					}

					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
