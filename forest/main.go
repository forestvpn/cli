package main

import (
	"encoding/json"
	"errors"
	"forest/api"
	"forest/forms"
	"forest/utils"
	"os"

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
				Name:  "signup",
				Usage: "Sign up to use ForestVPN",
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
					signinform := forms.SignInForm{email, []byte(password)}

					for {
						if !signinform.IsFilled() {
							err := signinform.PromptEmail()

							if err != nil {
								return err
							}

							err = signinform.PromptPassword()

							if err != nil {
								return err
							}
						} else {
							break
						}
					}

					err := signinform.ValidateEmail()

					if err != nil {
						return err
					}

					err = signinform.ValidatePassword()

					if err != nil {
						return err
					}

					signupform := forms.SignUpForm{}
					err = signupform.PromptPasswordConfirmation()

					if err != nil {
						return err
					}

					signupform.SignInForm = signinform
					err = signupform.ValidatePasswordConfirmation()

					if err != nil {
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

						err := utils.JsonDump(resp.Body(), utils.FirebaseAuthFile)

						if err == nil {
							color.Green("Signed Up")
						}
					}
					return err
				},
			},
			{
				Name:  "signin",
				Usage: "Sign into your forestVPN account",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "email",
						Destination: &email,
						Usage:       "Your email address",
						Value:       "",
						Aliases:     []string{"e"},
					},
					&cli.StringFlag{
						Name:        "password",
						Destination: &password,
						Usage:       "Your password",
						Value:       "",
						Aliases:     []string{"p"},
					},
				},
				Action: func(c *cli.Context) error {
					auth, _ := utils.IsAuthenticated()

					if auth {
						color.Green("Signed in")
					} else {
						//
					}

				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		color.Red(err.Error())
	}

}
