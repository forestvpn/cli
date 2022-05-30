package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"forest/utils"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/urfave/cli/v2"
)

func main() {
	utils.Init()
	app := &cli.App{
		Name:        "forest",
		Usage:       "",
		Description: "ForestVPN client for Linux",
		Commands: []*cli.Command{
			{
				Name:    "signup",
				Aliases: []string{"s"},
				Usage:   "Sign up and use ForestVPN",
				Action: func(c *cli.Context) error {
					reader := bufio.NewReader(os.Stdin)
					var email string
					var password []byte
					var confirmation string

					for {
						fmt.Print("Enter email: ")
						input, _ := reader.ReadString('\n')

						if len(input) > 5 && strings.Index(input, "@") > 0 {
							email = input[:len(input)-1]
							break
						} else {
							fmt.Println("Please, enter a correct email")
						}
					}
					for {
						fmt.Print("Enter password: ")
						input, _ := term.ReadPassword(0)

						if len(input) > 6 {
							password = input
							break
						} else {
							fmt.Println("\nPassword should be at least 7 characters long")
						}

					}

					fmt.Println()

					for {
						fmt.Print("Confirm password: ")
						input, _ := term.ReadPassword(0)
						fmt.Println()

						if bytes.Equal(password, input) {
							confirmation = string(input)
							break
						} else {
							return errors.New("password confirmation doesn't match")
						}
					}

					firebaseApiKey := os.Getenv("FB_API_KEY")
					client := resty.New()
					req, err := json.Marshal(map[string]any{"email": email, "password": confirmation, "returnSecureToken": true})

					if err != nil {
						return err
					}

					resp, err := client.R().
						SetHeader("Content-Type", "application/json").
						SetQueryParams(map[string]string{
							"key": firebaseApiKey,
						}).
						SetBody(req).
						Post("https://identitytoolkit.googleapis.com/v1/accounts:signUp")

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
