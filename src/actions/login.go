package actions

import (
	"context"
	"errors"
	"fmt"

	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/forestvpn/cli/auth"
)

func (w AuthClientWrapper) Login() error {
	// Create the user profile
	profile := w.AccountsMap.CreateUser()
	token, loginErr := profile.Token()
	if loginErr != nil {
		return loginErr
	}
	// Create a new context with the token as the access token
	authCtx := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, token.Raw())
	// Make a request to the WhoAmI endpoint
	userInfo, _, loginErr := w.ApiClient.APIClient.AuthApi.WhoAmI(authCtx).Execute()
	// If there is an error, log it and return it
	if loginErr != nil {
		fmt.Println(token.Raw())
		return loginErr
	}
	profile.ID, profile.Email = auth.ProfileID(userInfo.GetId()), auth.ProfileEmail(userInfo.GetEmail())
	profile.Touch()
	profile.MarkAsActive()

	device, err := auth.LoadDevice(profile.ID)
	if err != nil {
		return errors.Join(err, errors.New("(w AuthClientWrapper) Login()"))
	}
	if err = w.SetLocation(device, profile.ID); err != nil {
		return errors.Join(err, errors.New("(w AuthClientWrapper) Login()"))
	}
	// Create a new device for the user
	if device, err := w.ApiClient.CreateDevice(); err != nil {
		return err
	} else if err = auth.UpdateProfileDevice(device, profile.ID); err != nil {
		return err
	}
	return nil
}
