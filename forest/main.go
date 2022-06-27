package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/forestvpn/cli/api"
	"github.com/forestvpn/cli/auth"
	"github.com/forestvpn/cli/sockets"

	"github.com/fatih/color"
	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/getsentry/sentry-go"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	ini "gopkg.in/ini.v1"
)

func main() {
	var email string
	var password string

	err := auth.Init()

	if err != nil {
		panic(err)
	}

	if auth.IsRefreshTokenExists() {
		response, err := auth.GetAccessToken()

		if err == nil {
			auth.JsonDump(response.Body(), auth.FirebaseAuthFile)
		}
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		TracesSampleRate: 1.0,
	})

	if err != nil {
		sentry.Logger.Panicf("sentry.Init: %s", err)
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
								sentry.CaptureException(err)
								return err
							}

							signupform := auth.SignUpForm{}
							fmt.Print("Confirm password: ")
							password, err := term.ReadPassword(0)
							fmt.Println()

							if err != nil {
								sentry.CaptureException(err)
								return err
							}

							signupform.PasswordConfirmationField.Value = password
							signupform.SignInForm = signinform
							err = signupform.ValidatePasswordConfirmation()

							if err != nil {
								sentry.CaptureException(err)
								return err
							}

							response, err := auth.SignUp(signupform)

							if err != nil {
								sentry.CaptureException(err)
								return err
							}

							err = auth.HandleFirebaseAuthResponse(response)

							if err == nil {
								color.Green("Signed up")
							} else {
								sentry.CaptureException(err)
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
							if !auth.IsRefreshTokenExists() {
								signinform, err := auth.GetSignInForm(email, []byte(password))

								if err != nil {
									sentry.CaptureException(err)
									return err
								}

								response, err := auth.SignIn(signinform)

								if err != nil {
									sentry.CaptureException(err)
									return err
								}

								err = auth.HandleFirebaseSignInResponse(response)

								if err != nil {
									sentry.CaptureException(err)
									return err
								}

								err = auth.JsonDump(response.Body(), auth.FirebaseAuthFile)

								if err != nil {
									sentry.CaptureException(err)
									return err
								}

								response, err = auth.GetAccessToken()

								if err != nil {
									sentry.CaptureException(err)
									return err
								}

								err = auth.JsonDump(response.Body(), auth.FirebaseAuthFile)

								if err != nil {
									sentry.CaptureException(err)
									return err
								}
							}

							if !auth.IsDeviceCreated() {
								accessToken, err := auth.LoadAccessToken()

								if err != nil {
									sentry.CaptureException(err)
									return err
								}

								resp, err := api.CreateDevice(accessToken)

								if err != nil {
									sentry.CaptureException(err)
									return err
								}

								b, err := json.MarshalIndent(resp, "", "    ")

								if err != nil {
									sentry.CaptureException(err)
									return err
								}

								err = auth.JsonDump(b, auth.DeviceFile)

								if err != nil {
									sentry.CaptureException(err)
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
							err := os.Remove(auth.FirebaseAuthFile)

							if err == nil {
								color.Green("Signed out")
							} else {
								sentry.CaptureException(err)
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
									} else {
										sentry.CaptureException(err)
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
						sentry.CaptureException(err)
						return err
					}

					var wrappedLocations []*auth.LocationWrapper

					for _, loc := range locations {
						location := auth.LocationWrapper{Location: loc}
						wrappedLocations = append(wrappedLocations, &location)
					}

					accessToken, err := auth.LoadAccessToken()

					if err != nil {
						sentry.CaptureException(err)
						return err
					}

					resp, err := api.GetBillingFeatures(accessToken)

					if err != nil {
						sentry.CaptureException(err)
						return err
					}

					billingFeature := resp[0]
					constraint := billingFeature.GetConstraints()[0]
					subject := constraint.GetSubject()
					expireDate := billingFeature.GetExpiryDate()
					now := time.Now()

					if !expireDate.After(now) {
						return fmt.Errorf("expired: %s < %s; want >", expireDate, now)
					}

					for _, loc := range wrappedLocations {
						if strings.Contains(strings.Join(subject[:], " "), loc.Location.GetId()) {
							loc.IsAvailableOnSubscritionOnly = false
						} else {
							loc.IsAvailableOnSubscritionOnly = true
							loc.Message = "This location is available only for our subscribers. See our pricing details: https://forestvpn.com/pricing/"
						}
					}

					sort.Slice(wrappedLocations, func(i, j int) bool {
						return wrappedLocations[i].Location.Name < wrappedLocations[j].Location.Name
					})

					template := promptui.SelectTemplates{
						Active:   "{{.Location.Name | green }}, {{.Location.Country.Name}}",
						Inactive: "{{if .IsAvailableOnSubscritionOnly}} {{.Location.Name | faint }} {{else}} {{.Location.Name}} {{end}}" + ", {{.Location.Country.Name | faint}}",
						Selected: fmt.Sprintf(`{{ "%s" | green }} {{ .Location.Name | faint }}%s {{.Location.Country.Name | faint}}`, promptui.IconGood, color.New(color.FgHiBlack).Sprint(",")),
						Details:  "{{if .IsAvailableOnSubscritionOnly}} {{ .Message | cyan}} {{end}}",
					}

					prompt := promptui.Select{
						Label:     "Select location",
						Items:     wrappedLocations,
						Size:      15,
						Templates: &template,
					}

					_, result, err := prompt.Run()

					if err != nil {
						sentry.CaptureException(err)
						return err
					}

					var choice *auth.LocationWrapper

					for _, loc := range wrappedLocations {
						if strings.Contains(result, loc.Location.GetId()) {
							choice = loc
						}
					}

					if !strings.Contains(strings.Join(subject[:], " "), choice.Location.GetId()) {
						prompt := promptui.Prompt{
							Label: "See pricing? ([Y]es/[N]o)",
						}

						result, _ := prompt.Run()

						if strings.Contains("YESyesYesYEsyEsyeSyES", result) {
							err := exec.Command("xdg-open", "https://forestvpn.com/pricing/").Run()

							if err != nil {
								sentry.CaptureException(err)
								return err
							}
						}
						return nil

					} else {
						deviceID, err := auth.LoadDeviceID()

						if err != nil {
							sentry.CaptureException(err)
							return err
						}

						device, err := api.UpdateDevice(accessToken, deviceID, choice.Location.GetId())

						if err != nil {
							sentry.CaptureException(err)
							return err
						}

						config := ini.Empty()
						interfaceSection, err := config.NewSection("Interface")

						if err != nil {
							sentry.CaptureException(err)
							return err
						}

						_, err = interfaceSection.NewKey("Address", "127.0.0.1/8")

						if err != nil {
							sentry.CaptureException(err)
							return err
						}

						_, err = interfaceSection.NewKey("PrivateKey", device.Wireguard.GetPrivKey())

						if err != nil {
							sentry.CaptureException(err)
							return err
						}

						_, err = interfaceSection.NewKey("DNS", strings.Join(device.GetDns()[:], ","))

						if err != nil {
							sentry.CaptureException(err)
							return err
						}

						for _, peer := range device.Wireguard.GetPeers() {
							peerSection, err := config.NewSection("Peer")

							if err != nil {
								sentry.CaptureException(err)
								return err
							}

							_, err = peerSection.NewKey("AllowedIPs", strings.Join(peer.GetAllowedIps()[:], ","))

							if err != nil {
								sentry.CaptureException(err)
								return err
							}

							_, err = peerSection.NewKey("Endpoint", peer.GetEndpoint())

							if err != nil {
								sentry.CaptureException(err)
								return err
							}

							_, err = peerSection.NewKey("PublicKey", peer.GetPubKey())

							if err != nil {
								sentry.CaptureException(err)
								return err
							}

						}

						err = config.SaveTo(auth.WireguardConfig)

						if err != nil {
							sentry.CaptureException(err)
							return err
						}

						err = sockets.Disconnect()

						if err != nil {
							sentry.CaptureException(err)
							return err
						}

						request := fmt.Sprintf("connect %s%c", auth.WireguardConfig, sockets.DELIMITER)
						status, err := sockets.Communicate(request)

						if err != nil {
							sentry.CaptureException(err)
							return err
						}

						if status != 0 {
							return fmt.Errorf(`forestd could not perform action "connect" (exit status: %d)`, status)
						}
						color.Green(fmt.Sprintf("Connected to %s, %s", choice.Location.Name, choice.Location.Country.Name))
						session := map[string]string{"locationID": choice.Location.Id, "status": strconv.Itoa(status)}
						data, err := json.Marshal(session)

						if err != nil {
							sentry.CaptureException(err)
							return err
						}

						auth.JsonDump(data, auth.SessionFile)
					}
					return nil
				},
			},
			{
				Name:        "disconnect",
				Description: "Disconnect from ForestVPN",
				Action: func(ctx *cli.Context) error {
					err := sockets.Disconnect()

					if err == nil {
						color.Red("Disconnected")
					} else {
						sentry.CaptureException(err)
					}

					return err
				},
			},
			{
				Name:        "status",
				Description: "See wether connection is active",
				Action: func(ctx *cli.Context) error {
					isActive, err := sockets.IsActiveConnection()

					if err != nil {
						sentry.CaptureException(err)
						return err
					}

					session, err := auth.JsonLoad(auth.SessionFile)

					if err != nil {
						sentry.CaptureException(err)
						return err
					}

					if isActive && session["status"] == "0" {
						locationID := session["locationID"]
						locations, err := api.GetLocations()
						var location *forestvpn_api.Location

						if err != nil {
							sentry.CaptureException(err)
							return err
						}

						for _, loc := range locations {
							if loc.Id == locationID {
								location = &loc
							}
						}

						color.Green(fmt.Sprintf("Connected to %s, %s", location.Name, location.Country.Name))
					} else {
						color.Red("Disconnected")
					}
					return nil
				}},
		},
	}
	err = app.Run(os.Args)
	if err != nil {
		color.Red(err.Error())
	}

}
