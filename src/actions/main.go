// actions is a package containing a high-level structure that implements the functions to use as CLI Actions.
//
// See https://cli.urfave.org/v2/ for more information.

package actions

import (
	"encoding/json"
	"sort"

	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/forestvpn/cli/api"
	"github.com/forestvpn/cli/auth"
)

// AuthClientWrapper is a structure that is used as a high-level wrapper for both AuthClient and ApiClient.
// It is used as main wgrest and Firebase REST API client as both of wrapped structures share the same AccessToken for authentication purposes.
type AuthClientWrapper struct {
	ApiClient   *api.ApiClientWrapper
	AccountsMap *auth.UserDB
}

func GetAuthClientWrapper(profile *auth.Profile, apiHost string) (AuthClientWrapper, error) {
	accessToken, err := profile.Token()
	if err != nil {
		return AuthClientWrapper{}, err
	}
	return AuthClientWrapper{ApiClient: api.GetApiClient(accessToken.Raw(), apiHost)}, nil
}

func (w AuthClientWrapper) GetUnexpiredOrMostRecentBillingFeature(userID auth.ProfileID) (forestvpn_api.BillingFeature, error) {
	var billingFeatures []forestvpn_api.BillingFeature
	var err error
	var b forestvpn_api.BillingFeature

	if auth.BillingFeautureExists(userID) {
		billingFeatures, err = auth.LoadBillingFeatures(userID)
		if err != nil {
			return b, err
		}
		for _, b := range billingFeatures {
			if !auth.BillingFeatureExpired(b) {
				return b, nil
			}
		}
	}

	resp, err := w.ApiClient.GetBillingFeatures()
	if err != nil {
		return b, err
	}
	data, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		return b, err
	}
	path := auth.ProfilesDir + string(userID) + auth.BillingFeatureFile
	err = auth.JsonDump(data, path)
	if err != nil {
		return b, err
	}
	billingFeatures, err = auth.LoadBillingFeatures(userID)
	if err != nil {
		return b, err
	}
	sort.Slice(billingFeatures, func(i, j int) bool {
		return billingFeatures[i].GetExpiryDate().After(billingFeatures[j].GetExpiryDate())
	})

	return billingFeatures[0], nil
}
