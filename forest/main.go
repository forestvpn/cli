package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"forest/api"
	"forest/auth"
	"forest/utils"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	ini "gopkg.in/ini.v1"
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
							signinform, err := auth.GetSignInForm(email, []byte(password))

							if err != nil {
								return err
							}

							signupform := auth.SignUpForm{}
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
								signinform, err := auth.GetSignInForm(email, []byte(password))

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

								resp, err := api.CreateDevice(accessToken)

								if err != nil {
									return err
								}

								b, err := json.MarshalIndent(resp, "", "    ")

								if err != nil {
									return err
								}

								err = utils.JsonDump(b, utils.DeviceFile)

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
									emailfield, err := auth.GetEmailField(email)

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
					locations, err := api.GetLocations()

					if err != nil {
						return err
					}

					accessToken, err := utils.LoadAccessToken()

					if err != nil {
						return err
					}

					resp, err := api.GetBillingFeatures(accessToken)

					if err != nil {
						return err
					}

					billingFeature := resp[0]
					constraint := billingFeature.GetConstraints()[0]
					subject := constraint.GetSubject()

					var items []forestvpn_api.Location

					for _, loc := range locations {
						for _, locationID := range subject {
							if loc.GetId() == locationID {
								items = append(items, loc)
							}
						}
					}

					sort.Slice(items, func(i, j int) bool {
						return items[i].Name < items[j].Name
					})

					template := promptui.SelectTemplates{
						Active:   "{{.Name | green }}, {{.Country.Name}}",
						Inactive: "{{.Name}}, {{.Country.Name | faint}}",
						Selected: fmt.Sprintf(`{{ "%s" | green }} {{ .Name | faint }}%s {{.Country.Name | faint}}`, promptui.IconGood, color.New(color.FgHiBlack).Sprint(",")),
					}

					prompt := promptui.Select{
						Label:     "Select location",
						Items:     items,
						Size:      15,
						Templates: &template,
					}

					_, result, err := prompt.Run()

					if err != nil {
						return err
					}

					var locationID string

					for _, location := range locations {
						if strings.Contains(result, location.Id) {
							locationID = location.Id
						}
					}

					if len(locationID) == 0 {
						return fmt.Errorf("no location found with ID: %s", locationID)
					}

					deviceID, err := utils.LoadDeviceID()

					if err != nil {
						return err
					}

					device, err := api.UpdateDevice(accessToken, deviceID, locationID)

					if err != nil {
						return err
					}

					config := ini.Empty()
					interfaceSection, err := config.NewSection("Interface")

					if err != nil {
						return err
					}

					_, err = interfaceSection.NewKey("Address", "127.0.0.1/8")

					if err != nil {
						return err
					}

					_, err = interfaceSection.NewKey("PrivateKey", device.Wireguard.GetPrivKey())

					if err != nil {
						return err
					}

					_, err = interfaceSection.NewKey("DNS", strings.Join(device.GetDns()[:], ","))

					if err != nil {
						return err
					}

					for _, peer := range device.Wireguard.GetPeers() {
						peerSection, err := config.NewSection("Peer")

						if err != nil {
							return err
						}

						_, err = peerSection.NewKey("AllowedIPs", strings.Join(peer.GetAllowedIps()[:], ","))

						if err != nil {
							return err
						}

						_, err = peerSection.NewKey("Endpoint", peer.GetEndpoint())

						if err != nil {
							return err
						}

						_, err = peerSection.NewKey("PublicKey", peer.GetPubKey())

						if err != nil {
							return err
						}

					}

					err = config.SaveTo(utils.WireguardConfig)

					if err != nil {
						return err
					}

					conn, err := net.Dial("tcp", "localhost:2405")

					if err != nil {
						return err
					}

					_, err = bufio.NewWriter(conn).WriteString(fmt.Sprintf("connect %s", utils.WireguardConfig))

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
