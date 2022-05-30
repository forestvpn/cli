package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"forest/api"
	"forest/forms"
	"forest/utils"
	"os"

	"golang.org/x/term"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func main() {
	utils.Init()
	var email string
	var password string
	app := &cli.App{
		Name:        "forest",
		Usage:       "Connect to ForestVPN servers around the world!",
		Description: "ForestVPN client for Linux",
		Commands: []*cli.Command{
			{
				Name:    "signup",
				Aliases: []string{"s"},
				Usage:   "Sign up and use ForestVPN",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "email",
						Destination: &email,
						Usage:       "Email address to use to sign up",
						Value:       "",
						Aliases:     []string{"e"},
					},
					&cli.StringFlag{
						Name:        "password",
						Destination: &password,
						Usage:       "Password must be at least 8 characters long",
						Value:       "",
						Aliases:     []string{"p"},
					},
				},
				Action: func(c *cli.Context) error {
					reader := bufio.NewReader(os.Stdin)
					signupform := forms.SignUpForm{}
					signupform.Email = email
					signupform.Password = []byte(password)

					if len(email) == 0 && len(password) == 0 {
						fmt.Print("Enter email: ")
						email, err := reader.ReadString('\n')

						if err != nil {
							return err
						}

						signupform.Email = email[:len(email)-1]
						fmt.Print("Enter password: ")
						password, err := term.ReadPassword(0)

						if err != nil {
							return err

						}

						signupform.Password = password
						fmt.Println()

					}

					fmt.Println("Confirm password: ")
					PasswordConfirmation, err := term.ReadPassword(0)

					if err != nil {
						return err
					}

					fmt.Println()
					signupform.PasswordConfirmation = PasswordConfirmation

					valid, err := forms.IsValidForm(signupform)

					if !valid {
						return err
					}

					resp, err := api.SignUp(os.Getenv("FB_API_KEY"), signupform)

					if err == nil {
						var body map[string]map[string]string
						json.Unmarshal(resp.Body(), &body)

						if body["error"] != nil {
							respError := body["error"]
							return errors.New(respError["message"])
						}
						err := utils.JsonDump(resp.Body(), utils.FB_AUTH_DIR+"/firebase.json")

						if err != nil {
							return err
						}
						color.Green("Signed Up")
					}
					return err
				},
			},
			// {
			// 	Name:    "login",
			// 	Aliases: []string{"l"},
			// 	Usage:   "Login to your forestVPN account",
			// 	Action: func(c *cli.Context) error {
			// 		//
			// 	},
			// },
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		color.Red(err.Error())
	}

}
