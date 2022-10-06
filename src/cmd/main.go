package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/forestvpn/cli/actions"
	"github.com/forestvpn/cli/api"
	"github.com/forestvpn/cli/auth"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/fatih/color"
	"github.com/getsentry/sentry-go"
	"github.com/urfave/cli/v2"
)

var (
	// DSN is a Data Source Name for Sentry. It is stored in an environment variable and assigned during the build with ldflags.
	//
	// See https://docs.sentry.io/product/sentry-basics/dsn-explainer/ for more information.
	Dsn string
	// appVersion value is stored in an environment variable and assigned during the build with ldflags.
	appVersion string
	// firebaseApiKey is stored in an environment variable and assigned during the build with ldflags.
	firebaseApiKey string
	// ApiHost is a hostname of Forest VPN back-end API that is stored in an environment variable and assigned during the build with ldflags.
	apiHost string
)

const accountsMapFile = ".accounts.json"
const url = "https://forestvpn.com/checkout/"

func getAuthClientWrapper() (actions.AuthClientWrapper, error) {
	accountmap := auth.GetAccountMap(accountsMapFile)
	authClientWrapper := actions.AuthClientWrapper{AccountsMap: accountmap}
	authClient := auth.AuthClient{ApiKey: firebaseApiKey}

	user_id, _ := auth.LoadUserID()
	exists, _ := auth.IsRefreshTokenExists()

	if exists {
		expired, _ := auth.IsAccessTokenExpired(user_id)

		if expired {
			refreshToken, _ := auth.LoadRefreshToken()
			response, err := authClient.GetAccessToken(refreshToken)

			if err != nil {
				return authClientWrapper, err
			}

			user_id, err = authClientWrapper.SetUpProfile(response)

			if err != nil {
				return authClientWrapper, err
			}
		}
	}

	accessToken, _ := auth.LoadAccessToken(user_id)
	authClientWrapper.AuthClient = authClient
	authClientWrapper.ApiClient = api.GetApiClient(accessToken, apiHost)
	return authClientWrapper, nil
}

func main() {
	// email is user's email address used to sign in or sign up on the Firebase.
	var email string
	// password is user's password used during sign in or sign up on the Firebase.
	var password string
	// country is stores prompted country name to filter locations by country.
	var country string

	err := auth.Init()

	if err != nil {
		sentry.CaptureException(err)
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn: Dsn,
	})

	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	defer sentry.Flush(2 * time.Second)

	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Println(cCtx.App.Version)
	}

	app := &cli.App{
		Version:              appVersion,
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
						Name:  "ls",
						Usage: "See local accounts ever logged in",
						Action: func(c *cli.Context) error {
							return nil
						},
					},
					{
						Name:  "status",
						Usage: "See logged-in account info",
						Action: func(c *cli.Context) error {
							if !auth.IsAuthenticated() {
								fmt.Println("Are you logged in?")
								color := color.New(color.Faint)
								color.Println("Try 'forest account login'")
								return nil
							}

							user_id, err := auth.LoadUserID()

							if err != nil {
								return err
							}

							authClientWrapper, err := getAuthClientWrapper()

							if err != nil {
								return err
							}

							billingFeature, err := authClientWrapper.LoadOrGetBillingFeature(user_id)

							if err != nil {
								return err
							}

							expiryDate := billingFeature.GetExpiryDate()
							now := time.Now()
							left := expiryDate.Sub(now)
							days := left.Hours() / 24
							idToken, err := auth.LoadIdToken(user_id)

							if err != nil {
								return err
							}

							response, err := authClientWrapper.AuthClient.GetUserData(idToken)

							if err != nil {
								return err
							}

							data := make(map[string]any)
							err = json.Unmarshal(response.Body(), &data)

							if err != nil {
								return err
							}

							var email string

							var x interface{} = data["users"]
							switch users := x.(type) {
							case []interface{}:
								var y interface{} = users[0]
								switch data := y.(type) {
								case map[string]any:
									var z interface{} = data["email"]
									switch v := z.(type) {
									case string:
										email = v
									}
								}
							}

							plan := strings.Split(billingFeature.GetBundleId(), ".")[2]

							color.Green("Logged-in as %s", email)
							color.Green("Plan: %s", plan)

							if plan == "premium" || days > 0 {
								color.Green("%s days left", fmt.Sprint(int(days)))
							}

							return nil
						},
					},
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
							if auth.IsAuthenticated() {
								fmt.Println("Please, logout before attempting to register a new account.")
								color := color.New(color.Faint)
								color.Println("Try 'fvpn account logout'")
								return nil
							}

							authClientWrapper, err := getAuthClientWrapper()

							if err != nil {
								return err
							}

							err = authClientWrapper.Register(email, password)

							if err == nil {
								color.Green("Registered")
							}

							return err
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
							authClientWrapper, err := getAuthClientWrapper()

							if err != nil {
								return err
							}

							err = authClientWrapper.Login(email, password)

							if err == nil {
								color.Green("Logged in")
							}

							return err
						},
					},
					{
						Name:  "logout",
						Usage: "Log out from your ForestVPN account on this device",
						Action: func(c *cli.Context) error {
							state := actions.State{}
							status := state.GetStatus()

							if status {
								fmt.Println("Please, set down the connection before attempting to log out.")
								color := color.New(color.Faint)
								color.Println("Try 'forest state down'")
								return nil
							}

							exists, err := auth.IsRefreshTokenExists()

							if err != nil {
								return err
							}

							if exists {
								user_id, err := auth.LoadUserID()

								if err != nil {
									return err
								}

								if len(user_id) > 0 {
									err = auth.RemoveFirebaseAuthFile(user_id)

									if err != nil {
										return err
									}

									err = auth.RemoveActiveUserLockFile()

									if err != nil {
										return err
									}
								}
							}

							color.Red("Logged out")
							return nil
						},
					},
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
								return nil
							}

							client, err := getAuthClientWrapper()

							if err != nil {
								return err
							}

							user_id, err := auth.LoadUserID()

							if err != nil {
								return err
							}

							b, err := client.LoadOrGetBillingFeature(user_id)

							if err != nil {
								return err
							}

							device, err := auth.LoadDevice(user_id)

							if err != nil {
								return err
							}

							location := device.GetLocation()
							locations, err := client.ApiClient.GetLocations()

							if err != nil {
								return err
							}

							w := actions.GetWrappedLocations(b, locations)
							now := time.Now()
							exp := b.GetExpiryDate()
							left := exp.Sub(now)
							bid := b.GetBundleId()
							faint := color.New(color.Faint)

							for _, l := range w {
								if location.GetId() == l.Location.GetId() {
									if now.After(exp) {
										if l.Premium {
											color.Yellow("The location you were using is now unavailable, as your paid subscription has ended.")
											color.Yellow("You can keep using our VPN once you watch an ad in our iOS/Android apps.")
											color.Yellow("Or you can go Premium at %s.", url)
											return nil
										} else {
											color.Yellow("Your premium subscription is over, but you can keep using our VPN once you watch an ad in our iOS/Android apps.")
											color.Yellow("Or you can go Premium at %s.", url)
											return nil
										}
									} else if bid == "com.forestvpn.freemium" && int(left.Minutes()) < 5 {
										faint.Println("You currently have 5 more minutes of freemium left.")
									} else if int(left.Hours()/24) < 3 {
										if bid == "com.forestvpn.premium" {
											faint.Println("Your premium subscription will end in less than 3 days.")
										} else {
											faint.Println("Your free trial will end in less than 3 days.")
										}
									}
								}
							}

							state := actions.State{}

							if state.GetStatus() {
								return errors.New("state is already up and running")
							}

							err = state.SetUp(user_id)

							if err != nil {
								return err
							}

							if state.GetStatus() {
								user_id, err := auth.LoadUserID()

								if err != nil {
									return err
								}

								device, err := auth.LoadDevice(user_id)

								if err != nil {
									return err
								}

								location := device.GetLocation()
								country := location.GetCountry()

								color.Green("Connected to %s, %s", location.GetName(), country.GetName())
							} else {
								return errors.New("unexpected error: state.status is false after state is up")
							}

							return nil
						},
					},
					{
						Name:        "down",
						Description: "Disconnect from ForestVPN",
						Usage:       "Shut down the connection",
						Action: func(ctx *cli.Context) error {
							if !auth.IsAuthenticated() {
								fmt.Println("Are you logged in?")
								color := color.New(color.Faint)
								color.Println("Try 'forest account login'")
								return nil
							}

							state := actions.State{}

							if state.GetStatus() {
								user_id, err := auth.LoadUserID()

								if err != nil {
									return err
								}

								err = state.SetDown(user_id)

								if err != nil {
									return err
								}

								if state.GetStatus() {
									return errors.New("unexpected error: state.status is true after state is down")
								}

								color.Red("Disconnected")
							} else {
								return errors.New("state is already down")
							}

							return nil
						},
					},
					{
						Name:        "status",
						Description: "See wether connection is active",
						Usage:       "Check the status of the connection",
						Action: func(ctx *cli.Context) error {
							if !auth.IsAuthenticated() {
								fmt.Println("Are you signed in?")
								color := color.New(color.Faint)
								color.Println("Try 'forest account login'")
								return nil
							}

							client, err := getAuthClientWrapper()

							if err != nil {
								return err
							}

							user_id, err := auth.LoadUserID()

							if err != nil {
								return err
							}

							_, err = client.LoadOrGetBillingFeature(user_id)

							if err != nil {
								return err
							}

							state := actions.State{}
							status := state.GetStatus()

							if status {
								user_id, err := auth.LoadUserID()

								if err != nil {
									return err
								}

								device, err := auth.LoadDevice(user_id)

								if err != nil {
									return err
								}

								location := device.GetLocation()
								country := location.GetCountry()

								color.Green("Connected to %s, %s", location.GetName(), country.GetName())
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
						Name:  "status",
						Usage: "See the location is set as default VPN connection",
						Action: func(cCtx *cli.Context) error {
							faint := color.New(color.Faint)

							if !auth.IsAuthenticated() {
								fmt.Println("Are you logged in?")
								faint.Println("Try 'fvpn account login'")
								return nil
							}

							user_id, err := auth.LoadUserID()

							if err != nil {
								return err
							}

							device, err := auth.LoadDevice(user_id)

							if err != nil {
								return err
							}

							location := device.GetLocation()
							country := location.GetCountry()
							color.New(color.FgGreen).Println(fmt.Sprintf("Default location is set to %s, %s", location.GetName(), country.GetName()))
							return nil
						},
					},
					{
						Name:  "set",
						Usage: "Set the default location by specifying `UUID` or `Name`",
						Action: func(cCtx *cli.Context) error {
							faint := color.New(color.Faint)

							if !auth.IsAuthenticated() {
								fmt.Println("Are you logged in?")
								faint.Println("Try 'fvpn account login'")
								return nil
							}

							state := actions.State{}

							if state.GetStatus() {
								fmt.Println("Please, set down the connection before setting a new location.")
								faint.Println("Try 'fvpn state down'")
								return nil
							}

							arg := cCtx.Args().Get(0)

							if len(arg) < 1 {
								return errors.New("UUID or name required")
							}

							authClientWrapper, err := getAuthClientWrapper()

							if err != nil {
								return err
							}

							user_id, err := auth.LoadUserID()

							if err != nil {
								return err
							}

							billingFeature, err := authClientWrapper.LoadOrGetBillingFeature(user_id)

							if err != nil {
								return err
							}

							locations, err := authClientWrapper.ApiClient.GetLocations()

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
								err := fmt.Errorf("no such location: %s", arg)
								return err
							}

							expireDate := billingFeature.GetExpiryDate()
							now := time.Now()

							if now.After(expireDate) && location.Premium {
								color.Yellow("The location you want to use is now unavailable, as it requires a paid subscription.")
								color.Yellow("You can keep using our VPN once you watch an ad in our iOS/Android apps.")
								color.Yellow("Or you can go Premium at %s.", url)
								return nil
							}

							device, err := auth.LoadDevice(user_id)

							if err != nil {
								return err
							}

							device, err = authClientWrapper.ApiClient.UpdateDevice(device.GetId(), location.Location.GetId())

							if err != nil {
								return err
							}

							err = auth.UpdateProfileDevice(device)

							if err != nil {
								return err
							}

							err = authClientWrapper.SetLocation(device, user_id)

							if err != nil {
								return err
							}

							country := location.Location.GetCountry()
							color.New(color.FgGreen).Println(fmt.Sprintf("Default location is set to %s, %s", location.Location.GetName(), country.GetName()))
							return nil
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
							authClientWrapper, err := getAuthClientWrapper()

							if err != nil {
								return err
							}

							return authClientWrapper.ListLocations(country)
						},
					},
				},
			},
		},
	}

	err = app.Run(os.Args)

	if err != nil {
		sentry.CaptureException(err)
		caser := cases.Title(language.AmericanEnglish)
		msg := strings.Split(err.Error(), " ")
		msg[0] = caser.String(msg[0])
		color.Red(strings.Join(msg, " "))
	}
}
