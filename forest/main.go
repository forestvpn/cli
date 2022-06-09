package main

import (
	"encoding/json"
	"fmt"
	"forest/api"
	"forest/auth"
	"forest/auth/forms"
	"forest/utils"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/mitchellh/mapstructure"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func main() {
	var email string
	var password string
	utils.Init()

	if utils.IsRefreshTokenExists() {
		response, err := auth.GetAccessToken()

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

							response, err := auth.SignUp(signupform)

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

								response, err := auth.SignIn(signinform)

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

								response, err = auth.GetAccessToken()

								if err != nil {
									return err
								}

								err = utils.JsonDump(response.Body(), utils.FirebaseAuthFile)

								if err != nil {
									return err
								}
							}

							if !utils.IsDeviceCreated() {
								accessToken, err := utils.LoadAccessToken()

								if err != nil {
									return err
								}

								response, err := api.CreateDevice(accessToken)

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
						Name:  "signout",
						Usage: "Sign out from your ForestVPN account on this device",
						Action: func(c *cli.Context) error {
							err := os.Remove(utils.FirebaseAuthFile)

							if err == nil {
								color.Green("Signed out")
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
			{
				Name:  "connect",
				Usage: "Connect to the ForestVPN",
				Action: func(c *cli.Context) error {
					response, err := api.GetLocations()

					if err != nil {
						return err
					}

					err = utils.HandleApiResponse(response)

					if err != nil {
						return err
					}

					var body []map[string]any

					err = json.Unmarshal(response.Body(), &body)

					if err != nil {
						return err
					}

					var location api.Location
					var locations []api.Location

					for _, loc := range body {
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
						locations = append(locations, location)
					}

					sort.Slice(locations, func(i, j int) bool {
						return locations[i].Name < locations[j].Name
					})

					template := promptui.SelectTemplates{
						Active:   "{{.Name | green }}, {{.Country.Name}}",
						Inactive: "{{.Name}}, {{.Country.Name | faint}}",
						Selected: fmt.Sprintf(`{{ "%s" | green }} {{ .Name | faint }}%s {{.Country.Name | faint}}`, promptui.IconGood, color.New(color.FgHiBlack).Sprint(",")),
					}

					prompt := promptui.Select{
						Label:     "Select location",
						Items:     locations,
						Size:      15,
						Templates: &template,
					}

					_, result, err := prompt.Run()
					var locationId string

					for _, location := range locations {
						if strings.Contains(result, location.Id) {
							locationId = location.Id
						}
					}

					if len(locationId) == 0 {
						return fmt.Errorf("no location found with ID: %s", locationId)
					}

					if err != nil {
						return err
					}

					deviceID, err := utils.LoadDeviceID()

					if err != nil {
						return err
					}

					accessToken, err := utils.LoadAccessToken()

					if err != nil {
						return err
					}

					response, err = api.UpdateDevice(accessToken, deviceID, locationId)

					if err != nil {
						return err
					}

					fmt.Print(response.String())
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
