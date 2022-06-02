package main

import (
	"fmt"
	"forest/api"
	"forest/auth/forms"
	"forest/utils"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func main() {
	utils.Init()
	var email string
	var password string
	var firebaseApiKey = os.Getenv("FB_API_KEY")
	app := &cli.App{
		Name:        "forest",
		Usage:       "Connect to ForestVPN servers around the world!",
		Description: "ForestVPN client for Linux",
		Commands: []*cli.Command{
			{
				Name:  "auth",
				Usage: "Authentication services",
				Subcommands: []*cli.Command{
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
							signinform, err := forms.GetSignInForm(email, []byte(password))

							if err != nil {
								return err
							}

							signupform := forms.SignUpForm{}
							fmt.Print("Confirm password: ")
							password, err := term.ReadPassword(0)
							fmt.Println()

							if err != nil {
								return err
							}

							signupform.PasswordConfirmationField.Value = password
							signupform.SignInForm = signinform
							err = signupform.ValidatePasswordConfirmation()

							if err != nil {
								return err
							}

							resp, err := api.SignUp(firebaseApiKey, signupform)

							if err == nil {
								return utils.HandleFirebaseSignInUpResponse(resp, "Signed up")
							}
							return err
						},
					},
					{
						Name:  "signin",
						Usage: "Sign into your ForestVPN account",
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
							signinform, err := forms.GetSignInForm(email, []byte(password))

							if err != nil {
								return err
							}
							resp, err := api.SignIn(firebaseApiKey, signinform)

							if err == nil {
								return utils.HandleFirebaseSignInUpResponse(resp, "Signed in")
							}
							return err

						},
					},
					{
						Name:  "account",
						Usage: "Manage multiple accounts",
						Subcommands: []*cli.Command{
							{
								Name:  "show",
								Usage: "Show all user accounts logged in",
							},
							{
								Name:  "default",
								Usage: "Set a default account",
								Flags: []cli.Flag{
									&cli.StringFlag{
										Name:        "email",
										Destination: &email,
										Usage:       "Email address of your account",
										Value:       "",
										Aliases:     []string{"e"},
									},
								},
								Action: func(c *cli.Context) error {
									emailfield, err := forms.GetEmailField(email)

									if err == nil {
										fmt.Println(emailfield.Value)
									}
									return err
								},
							},
						},
					},
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		color.Red(err.Error())
	}

}
