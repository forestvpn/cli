package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/forestvpn/cli/actions"
	"github.com/forestvpn/cli/api"
	"github.com/forestvpn/cli/auth"
	"github.com/forestvpn/cli/utils"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/fatih/color"
	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/getsentry/sentry-go"
	"github.com/urfave/cli/v2"
)

var (
	// DSN is a Data Source Name for Sentry. It is stored in an environment variable and assigned during the build with ldflags.
	//
	// See https://docs.sentry.io/product/sentry-basics/dsn-explainer/ for more information.
	DSN = os.Getenv("SENTRY_DSN")
	// appVersion value is stored in an environment variable and assigned during the build with ldflags.
	appVersion string
	// firebaseApiKey is stored in an environment variable and assigned during the build with ldflags.
	firebaseApiKey = os.Getenv("STAGING_FIREBASE_API_KEY")
	// ApiHost is a hostname of Forest VPN back-end API that is stored in an environment variable and assigned during the build with ldflags.
	ApiHost = os.Getenv("STAGING_API_URL")
)

func main() {
	// email is user's email address used to sign in or sign up on the Firebase.
	var email string
	// password is user's password used during sign in or sign up on the Firebase.
	var password string
	// country is stores prompted country name to filter locations by country.
	var country string
	// includeRoutes is a flag that indicates wether to route networks from system routing table into Wireguard tunnel interface.
	var includeRoutes bool

	err := auth.Init()

	if err != nil {
		panic(err)
	}

	authClient := auth.AuthClient{ApiKey: firebaseApiKey}

	if auth.IsRefreshTokenExists() {
		response, err := authClient.GetAccessToken()

		if err == nil {
			auth.JsonDump(response.Body(), auth.FirebaseAuthFile)
		}
	}

	accessToken, _ := auth.LoadAccessToken()
	wrapper := api.GetApiClient(accessToken, ApiHost)
	apiClient := actions.AuthClientWrapper{AuthClient: authClient, ApiClient: wrapper}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:              DSN,
		TracesSampleRate: 1.0,
	})

	if err != nil {
		sentry.Logger.Panicf("%s: %s", err, DSN)
	}

	app := &cli.App{
		EnableBashCompletion: true,
		Suggest:              true,
		Name:                 "fvpn",
		Usage:                "fast, secure, and modern VPN",
		Commands: []*cli.Command{
			{
				Name:  "account",
				Usage: "Manage your account",
				Subcommands: []*cli.Command{
					{
						Name:  "register",
						Usage: "Sign up to use ForestVPN",
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
								Usage:       "Password must be at least 8 characters long",
								Value:       "",
								Aliases:     []string{"p"},
							},
						},
						Action: func(c *cli.Context) error {
							return apiClient.Register(email, password)
						},
					},
					{
						Name:  "login",
						Usage: "Log into your ForestVPN account",
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
							localDevice, _ := auth.JsonLoad(auth.DeviceFile)

							// if err != nil {
							// 	return err
							// }

							deviceID := localDevice["id"]
							return apiClient.Login(email, password, deviceID)
						},
					},
					{
						Name:  "logout",
						Usage: "Log out from your ForestVPN account on this device",
						Action: func(c *cli.Context) error {
							return apiClient.Logout()
						},
					},
					// {
					// 	Name:  "account",
					// 	Usage: "Manage multiple accounts",
					// 	Subcommands: []*cli.Command{
					// 		{
					// 			Name:  "show",
					// 			Usage: "Show all user accounts logged in",
					// 		},
					// 		{
					// 			Name:  "default",
					// 			Usage: "Set a default account",
					// 			Flags: []cli.Flag{
					// 				&cli.StringFlag{
					// 					Name:        "email",
					// 					Destination: &email,
					// 					Usage:       "Email address of your account",
					// 					Value:       "",
					// 					Aliases:     []string{"e"},
					// 				},
					// 			},
					// 			Action: func(c *cli.Context) error {
					// 				if !auth.IsAuthenticated() {
					// 					fmt.Println("Are you signed in?")
					// 					color := color.New(color.Faint)
					// 					color.Println("Try 'forest auth signin'")
					// 					return nil
					// 				}
					// 				emailfield, err := auth.GetEmailField(email)

					// 				if err == nil {
					// 					fmt.Println(emailfield.Value)
					// 				} else {
					// 					sentry.CaptureException(err)
					// 				}
					// 				return err
					// 			},
					// 		},
					// 	},
					// },
				},
			},
			{
				Name:  "state",
				Usage: "Control the state of connection",
				Subcommands: []*cli.Command{
					{

						Name:  "up",
						Usage: "Connect to the ForestVPN",
						Action: func(c *cli.Context) error {
							if !auth.IsAuthenticated() {
								fmt.Println("Are you logged in?")
								color := color.New(color.Faint)
								color.Println("Try 'forest account login'")
							} else if !auth.IsLocationSet() {
								fmt.Println("Please, choose the location to connect.")
								color := color.New(color.Faint)
								color.Println("Use 'fvpn location ls' to see available locations.")
							} else {
								state := actions.State{}
								err := state.SetUp(auth.WireguardConfig)

								if err != nil {
									return err
								}

								locations, err := wrapper.GetLocations()

								if err != nil {
									return err
								}

								session, err := auth.JsonLoad(auth.SessionFile)

								if err != nil {
									return err
								}

								endpoint := session["endpoint"]

								if len(endpoint) == 0 {
									err := errors.New("no endpoint found in session file")
									sentry.CaptureException(err)
									return err
								}

								endpoints, err := net.LookupIP(endpoint)

								if err != nil {
									return err
								}

								hostip, err := utils.GetHostIP()

								if err != nil {
									return err
								}

								var matchedIP string

								for _, endpoint := range endpoints {
									if endpoint.String() == hostip.String() {
										matchedIP = endpoint.String()
									}
								}

								if len(matchedIP) == 0 {
									sentry.CaptureException(fmt.Errorf("no host ip %s found in endpoints %s", hostip.String(), endpoints))
									matchedIP = "unknown"
								}

								var id, city, country string

								for _, loc := range locations {
									id = loc.GetId()

									if id == session["location"] {
										city = loc.GetName()
										country = loc.Country.GetName()
									}
								}

								if len(city)+len(country) < 0 {
									err := fmt.Errorf("no such location: %s", id)
									sentry.CaptureException(err)
									return err
								}

								color.New(color.FgGreen).Printf("Your IP address is %s", matchedIP)
								color.New(color.FgGreen).Printf("Connected to %s, %s\n", city, country)
								session["status"] = "up"
								data, err := json.MarshalIndent(session, "", "    ")

								if err != nil {
									return err
								}

								return auth.JsonDump(data, auth.SessionFile)

							}
							return nil
						},
					},
					{
						Name:        "down",
						Description: "Disconnect from ForestVPN",
						Action: func(ctx *cli.Context) error {
							if !auth.IsAuthenticated() {
								fmt.Println("Are you signed in?")
								color := color.New(color.Faint)
								color.Println("Try 'forest auth signin'")
								return nil
							}

							state := actions.State{}

							if state.GetStatus() {
								err := state.SetDown(auth.WireguardConfig)

								if err != nil {
									return err
								}
							}

							color.Red("Disconnected")
							session, err := auth.JsonLoad(auth.SessionFile)

							if err != nil {
								return err
							}

							session["status"] = "down"
							data, err := json.MarshalIndent(session, "", "    ")

							if err != nil {
								return err
							}

							return auth.JsonDump(data, auth.SessionFile)

						},
					},
					{
						Name:        "status",
						Description: "See wether connection is active",
						Action: func(ctx *cli.Context) error {
							if !auth.IsAuthenticated() {
								fmt.Println("Are you signed in?")
								color := color.New(color.Faint)
								color.Println("Try 'forest auth signin'")
								return nil
							}

							state := actions.State{}
							session, err := auth.JsonLoad(auth.SessionFile)

							if err != nil {
								return err
							}

							if state.GetStatus() && session["status"] == "up" {
								id := session["location"]
								locations, err := wrapper.GetLocations()

								if err != nil {
									return err
								}

								var location forestvpn_api.Location

								for _, loc := range locations {
									if loc.GetId() == id {
										location = loc
									}
								}

								color.Green(fmt.Sprintf("Connected to %s, %s", location.Name, location.Country.Name))

							} else {
								color.Red("Disconnected")
							}

							return nil

						},
					},
				},
			},
			{
				Name:  "location",
				Usage: "Manage locations",
				Subcommands: []*cli.Command{
					{
						Name:        "set",
						Description: "Set the default location by specifying `UUID` or `Name`",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:        "include-routes",
								Destination: &includeRoutes,
								Usage:       "Include existing routes",
								Value:       false,
								Aliases:     []string{"i"},
							},
						},
						Action: func(cCtx *cli.Context) error {

							if !auth.IsAuthenticated() {
								fmt.Println("Are you logged in?")
								color := color.New(color.Faint)
								color.Println("Try 'forest account login'")
								return nil
							}

							arg := cCtx.Args().Get(0)

							if len(arg) < 1 {
								return errors.New("UUID or name required")
							}

							resp, err := wrapper.GetBillingFeatures()

							if err != nil {
								return err
							}

							billingFeature := resp[0]
							locations, err := wrapper.GetLocations()

							if err != nil {
								return err
							}

							wrappedLocations := actions.GetWrappedLocations(billingFeature, locations)
							var location actions.LocationWrapper
							id, err := uuid.Parse(arg)
							found := false

							if err != nil {
								for _, loc := range wrappedLocations {
									if strings.EqualFold(loc.Location.GetName(), arg) {
										location = loc
										found = true
										break
									}
								}
							} else {
								for _, loc := range wrappedLocations {
									if strings.EqualFold(location.Location.GetId(), id.String()) {
										location = loc
										found = false
										break
									}
								}
							}

							if !found {
								return fmt.Errorf("no such location: %s", arg)
							}

							err = apiClient.SetLocation(billingFeature, location, includeRoutes)

							if err != nil {
								return err
							}

							session, _ := auth.JsonLoad(auth.SessionFile)
							endpoint := session["endpoint"]
							session["location"] = location.Location.GetId()
							session["status"] = "down"
							session["endpoint"] = endpoint
							data, err := json.MarshalIndent(session, "", "    ")

							if err != nil {
								return err
							}

							err = auth.JsonDump(data, auth.SessionFile)

							return err

						},
					},
					{
						Name:        "ls",
						Description: "List locations",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "country",
								Destination: &country,
								Usage:       "Show locations by country",
								Value:       "",
								Aliases:     []string{"c"},
								Required:    false,
							},
						},
						Action: func(c *cli.Context) error {
							return apiClient.ListLocations(country)
						},
					},
				},
			},
			{
				Name:  "version",
				Usage: "Show the version of ForestVPN CLI",
				Action: func(ctx *cli.Context) error {
					fmt.Printf("ForestVPN CLI %s\n", appVersion)
					return nil
				},
			},
		},
	}

	err = app.Run(os.Args)

	if err != nil {
		caser := cases.Title(language.AmericanEnglish)
		sentry.CaptureException(err)
		msg := strings.Split(err.Error(), " ")
		msg[0] = caser.String(msg[0])
		color.Red(strings.Join(msg, " "))
	}
}
