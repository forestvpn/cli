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
	"github.com/forestvpn/cli/utils"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

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
								fmt.Println("Try 'forest account login'")
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

							b, err := authClientWrapper.GetUnexpiredOrMostRecentBillingFeature(user_id)

							if err != nil {
								return err
							}

							expiryDate := b.GetExpiryDate()
							now := time.Now()
							left := expiryDate.Sub(now)
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

							caser := cases.Title(language.English)
							plan := caser.String(strings.Split(b.GetBundleId(), ".")[2])
							fmt.Printf("Logged-in as %s\n", email)
							fmt.Printf("Plan: %s\n", plan)
							dt := fmt.Sprintf("%d-%d-%d %d:%d:%d",
								expiryDate.Year(),
								expiryDate.Month(),
								expiryDate.Day(),
								expiryDate.Hour(),
								expiryDate.Minute(),
								expiryDate.Second())
							tz, err := utils.GetLocalTimezone()

							if err != nil {
								n, _ := now.Zone()
								tz = n
							}

							if now.After(expiryDate) {
								t := now.Sub(expiryDate)
								fmt.Printf("Status: expired %s ago at %s %s\n", utils.HumanizeDuration(t), dt, tz)
							} else {
								fmt.Printf("Status: expires in %s at %s %s\n", utils.HumanizeDuration(left), dt, tz)

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
								fmt.Println("Try 'fvpn account logout'")
								return nil
							}

							authClientWrapper, err := getAuthClientWrapper()

							if err != nil {
								return err
							}

							err = authClientWrapper.Register(email, password)

							if err == nil {
								fmt.Println("Registered")
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
								fmt.Println("Logged in")
							}

							return err
						},
					},
					{
						Name:  "logout",
						Usage: "Log out from your ForestVPN account on this device",
						Action: func(c *cli.Context) error {
							if !auth.IsAuthenticated() {
								fmt.Println("Are you logged in?")
								fmt.Println("Try 'forest account login'")
								return nil
							}

							state := actions.State{}
							status := state.GetStatus()

							if status {
								fmt.Println("Please, set down the connection before attempting to log out.")
								fmt.Println("Try 'forest state down'")
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

									m := auth.GetAccountMap(accountsMapFile)
									err = m.RemoveAccount(user_id)

									if err != nil {
										return err
									}
								}
							}

							fmt.Println("Logged out")
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
								fmt.Println("Try 'forest account login'")
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

							state := actions.State{}

							if state.GetStatus() {
								fmt.Println("State is already up and running")
								os.Exit(1)
							}

							b, err := client.GetUnexpiredOrMostRecentBillingFeature(user_id)

							if err != nil {
								return err
							}

							device, err := auth.LoadDevice(user_id)

							if err != nil {
								return err
							}

							bid := b.GetBundleId()
							location := device.GetLocation()
							now := time.Now()
							exp := b.GetExpiryDate()
							left := exp.Sub(now)

							if now.After(exp) {
								if actions.IsPremiumLocation(b, location) && bid == "com.forestvpn.premium" {
									fmt.Println("The location you were using is now unavailable, as your paid subscription has ended.")
									fmt.Printf("You can keep using ForestVPN once you watch an ad in our mobile app, or simply go Premium at %s.\n", url)
									os.Exit(1)
								} else {
									fmt.Println("Your 30-minute session is over.")
									fmt.Printf("You can keep using ForestVPN once you watch an ad in our mobile app, or simply go Premium at %s.\n", url)
									os.Exit(1)
								}
							} else if bid == "com.forestvpn.freemium" && int(left.Minutes()) < 5 {
								fmt.Println("You currently have 5 more minutes of freemium left.")
							} else if int(left.Hours()/24) < 3 {
								if bid == "com.forestvpn.premium" {
									fmt.Println("Your premium subscription will end in less than 3 days.")
								} else {
									fmt.Println("Your free trial will end in less than 3 days.")
								}
							}

							err = state.SetUp(user_id)

							if err != nil {
								return err
							}

							if state.GetStatus() {
								country := location.GetCountry()
								fmt.Printf("Connected to %s, %s\n", location.GetName(), country.GetName())
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
								fmt.Println("Try 'forest account login'")
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

								fmt.Println("Disconnected")
							} else {
								fmt.Println("State is already down")
								os.Exit(1)
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
								fmt.Println("Are you logged in?")
								fmt.Println("Try 'forest account login'")
								return nil
							}

							state := actions.State{}

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

								fmt.Printf("Connected to %s, %s\n", location.GetName(), country.GetName())
							} else {
								fmt.Println("Disconnected")
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
							if !auth.IsAuthenticated() {
								fmt.Println("Are you logged in?")
								fmt.Println("Try 'fvpn account login'")
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
							fmt.Printf("Default location is set to %s, %s\n", location.GetName(), country.GetName())
							return nil
						},
					},
					{
						Name:  "set",
						Usage: "Set the default location by specifying `UUID` or `Name`",
						Action: func(cCtx *cli.Context) error {
							if !auth.IsAuthenticated() {
								fmt.Println("Are you logged in?")
								fmt.Println("Try 'fvpn account login'")
								return nil
							}

							state := actions.State{}

							if state.GetStatus() {
								fmt.Println("Please, set down the connection before setting a new location.")
								fmt.Println("Try 'fvpn state down'")
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

							b, err := authClientWrapper.GetUnexpiredOrMostRecentBillingFeature(user_id)

							if err != nil {
								return err
							}

							locations, err := authClientWrapper.ApiClient.GetLocations()

							if err != nil {
								return err
							}

							wrappedLocations := actions.GetLocationWrappers(b, locations)
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

							if time.Now().After(b.GetExpiryDate()) && location.Premium {
								fmt.Println("The location you want to use is now unavailable, as it requires a paid subscription.")
								fmt.Printf("You can keep using ForestVPN once you watch an ad in our mobile app, or simply go Premium at %s.\n", url)
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
							fmt.Printf("Default location is set to %s, %s\n", location.Location.GetName(), country.GetName())
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
							if !auth.IsAuthenticated() {
								fmt.Println("Are you logged in?")
								fmt.Println("Try 'fvpn account login'")
								return nil
							}

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
		fmt.Println(strings.Join(msg, " "))
	}
}
