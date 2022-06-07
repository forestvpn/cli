package main

import (
	"encoding/json"
	"fmt"
	"forest/api"
	"forest/auth"
	"forest/auth/forms"
	"forest/utils"
	"os"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/mitchellh/mapstructure"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func main() {
	var firebaseApiKey = os.Getenv("FIREBASE_API_KEY")
	var email string
	var password string
	utils.Init()

	if utils.IsRefreshTokenExists() {
		response, err := auth.GetAccessToken(firebaseApiKey)

		if err != nil {
			color.Red(err.Error())
		}

		err = utils.JsonDump(response.Body(), utils.FirebaseAuthFile)

		if err != nil {
			color.Red(err.Error())
		}
	}
	app := &cli.App{
		Name:        "forest",
		Usage:       "ForestVPN client for Linux",
		Description: "Fast, secure, and modern VPN",
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

							response, err := auth.SignUp(firebaseApiKey, signupform)

							if err != nil {
								return err
							}

							err = utils.HandleFirebaseAuthResponse(response)

							if err == nil {
								color.Green("Signed up")
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
							if !utils.IsRefreshTokenExists() {
								signinform, err := forms.GetSignInForm(email, []byte(password))

								if err != nil {
									return err
								}

								response, err := auth.SignIn(firebaseApiKey, signinform)

								if err != nil {
									return err
								}

								err = utils.HandleFirebaseSignInResponse(response)

								if err != nil {
									return err
								}

								err = utils.JsonDump(response.Body(), utils.FirebaseAuthFile)

								if err != nil {
									return err
								}

								response, err = auth.GetAccessToken(firebaseApiKey)

								if err != nil {
									return err
								}

								err = utils.JsonDump(response.Body(), utils.FirebaseAuthFile)

								if err != nil {
									return err
								}
							}

							if !utils.IsDeviceRegistered() {
								accessToken, err := utils.LoadAccessToken()

								if err != nil {
									return err
								}

								response, err := api.RegisterDevice(accessToken)

								if err != nil {
									return err
								}

								err = utils.HandleApiResponse(response)

								if err != nil {
									return err
								}

								err = utils.JsonDump(response.Body(), utils.DeviceFile)

								if err != nil {
									return err
								}

							}

							color.Green("Signed in")

							return nil
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
			{
				Name:  "locations",
				Usage: "Show available locations",
				Action: func(c *cli.Context) error {
					response, err := api.GetLocations()

					if err != nil {
						return err
					}

					err = utils.HandleApiResponse(response)

					if err != nil {
						return err
					}

					var locations []map[string]any

					err = json.Unmarshal(response.Body(), &locations)

					if err != nil {
						return err
					}

					var location api.Location
					var locationStructs []api.Location

					for _, loc := range locations {
						var country api.Country
						err := mapstructure.Decode(loc["country"], &country)

						if err != nil {
							return err
						}

						err = mapstructure.Decode(loc, &location)

						if err != nil {
							return err
						}

						location.Country = country
						locationStructs = append(locationStructs, location)
					}

					var items []string

					for _, loc := range locationStructs {
						items = append(items, fmt.Sprintf("%s, %s", loc.Name, loc.Country.Name))
					}

					prompt := promptui.Select{
						Label: "Select location",
						Items: items,
					}

					_, result, err := prompt.Run()

					if err != nil {
						return err
					}

					fmt.Printf("You choose %q\n", result)
					return err
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		color.Red(err.Error())
	}

}
