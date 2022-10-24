// actions is a package containing a high-level structure that implements the functions to use as CLI Actions.
//
// See https://cli.urfave.org/v2/ for more information.

package actions

import (
	"encoding/json"
	"errors"
	"os"
	"sort"

	"github.com/go-resty/resty/v2"

	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/forestvpn/cli/api"
	"github.com/forestvpn/cli/auth"
)

var (
	// firebaseApiKey is stored in an environment variable and assigned during the build with ldflags.
	firebaseApiKey = os.Getenv("STAGING_FIREBASE_API_KEY")
	// ApiHost is a hostname of Forest VPN back-end API that is stored in an environment variable and assigned during the build with ldflags.
	apiHost = os.Getenv("STAGING_API_URL")
)

// AuthClientWrapper is a structure that is used as a high-level wrapper for both AuthClient and ApiClient.
// It is used as main wgrest and Firebase REST API client as both of wrapped structures share the same AccessToken for authentication purposes.
type AuthClientWrapper struct {
	AuthClient  auth.AuthClient
	ApiClient   api.ApiClientWrapper
	AccountsMap auth.AccountsMap
}

func (w AuthClientWrapper) SetUpProfile(response *resty.Response) (string, error) {
	var data map[string]any
	err := json.Unmarshal(response.Body(), &data)

	if err != nil {
		return "", err
	}

	var x interface{} = data["refreshToken"]
	switch refreshToken := x.(type) {
	case string:
		if len(refreshToken) > 0 {
			response, err = w.AuthClient.GetAccessToken(refreshToken)

			if err != nil {
				return "", err
			}

			err = auth.HandleFirebaseSignInResponse(response)

			if err != nil {
				return "", err
			}

			data = make(map[string]any)
			err = json.Unmarshal(response.Body(), &data)

			if err != nil {
				return "", err
			}
		}
	}

	var i interface{} = data["refresh_token"]
	switch refreshToken := i.(type) {
	case string:
		if len(refreshToken) > 0 {
			var y interface{} = data["user_id"]
			switch user_id := y.(type) {
			case string:
				if len(user_id) > 0 {
					email := w.AccountsMap.GetEmail(user_id)

					if len(email) == 0 {
						var y interface{} = data["access_token"]
						switch accessToken := y.(type) {
						case string:
							if len(accessToken) > 0 {
								w.ApiClient.AccessToken = accessToken
							} else {
								return user_id, errors.New("unexpected error: invalid access token")
							}
						}

						device, _ := auth.LoadDevice(user_id)

						if len(device.GetId()) == 0 {
							device, err = w.ApiClient.CreateDevice()

							if err != nil {
								return user_id, err
							}

							path := auth.ProfilesDir + user_id

							if _, err := os.Stat(path); os.IsNotExist(err) {
								err = os.Mkdir(path, 0755)

								if err != nil {
									return user_id, err
								}
							}

							data, err := json.MarshalIndent(device, "", "    ")

							if err != nil {
								return user_id, err
							}

							path = auth.ProfilesDir + user_id + auth.DeviceFile
							err = auth.JsonDump(data, path)

							if err != nil {
								return user_id, err
							}
						}

						err = w.SetLocation(device, user_id)

						if err != nil {
							return user_id, err
						}
					}

					path := auth.ProfilesDir + user_id + auth.FirebaseAuthFile
					err = auth.JsonDump(response.Body(), path)

					if err != nil {
						return user_id, err
					}

					var z interface{} = data["expires_in"]
					switch exp := z.(type) {
					case string:
						err = auth.DumpAccessTokenExpireDate(user_id, exp)
						return user_id, err
					}

				} else {
					return user_id, errors.New("error parsing firebase sign in response: invalid user_id")
				}
			}
		} else {
			return "", errors.New("unknown response type")
		}
	}

	return "", nil

}

func (w AuthClientWrapper) GetUnexpiredOrMostRecentBillingFeature(user_id string) (forestvpn_api.BillingFeature, error) {
	var billingFeatures []forestvpn_api.BillingFeature
	var err error
	foundUnexpiredBillingFeature := false
	var b forestvpn_api.BillingFeature

	for i := 0; i < 2; i++ {
		if auth.BillingFeautureExists(user_id) && !foundUnexpiredBillingFeature {
			billingFeatures, err = auth.LoadBillingFeatures(user_id)

			if err != nil {
				return b, err
			}

			for _, b := range billingFeatures {
				if !auth.BillingFeatureExpired(b) {
					foundUnexpiredBillingFeature = true
				}
			}
		}

		if !foundUnexpiredBillingFeature && i == 0 {
			resp, err := w.ApiClient.GetBillingFeatures()

			if err != nil {
				return b, err
			}

			data, err := json.MarshalIndent(resp, "", "    ")

			if err != nil {
				return b, err
			}

			path := auth.ProfilesDir + user_id + auth.BillingFeatureFile
			err = auth.JsonDump(data, path)

			if err != nil {
				return b, err
			}
		}
	}

	if !foundUnexpiredBillingFeature {
		sort.Slice(billingFeatures, func(i, j int) bool {
			return billingFeatures[i].GetExpiryDate().After(billingFeatures[j].GetExpiryDate())
		})
	}

	return billingFeatures[0], nil
}

func GetAuthClientWrapper() (AuthClientWrapper, error) {
	accountsmap := auth.GetAccountsMap(auth.AccountsMapFile)
	authClientWrapper := AuthClientWrapper{AccountsMap: accountsmap}
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
