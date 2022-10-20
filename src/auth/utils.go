// Authentication related utilities around firebase REST authentication workflow.
// See https://firebase.google.com/docs/reference/rest for more information.
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/getsentry/sentry-go"
	"github.com/go-resty/resty/v2"
	"github.com/olekukonko/tablewriter"
)

var home, _ = os.UserHomeDir()

// AppDir is Forest CLI application directory.
var AppDir = home + "/.forestvpn/"

// The DeviceFile represents the device created for the user.
//
// Read more: https://github.com/forestvpn/api-client-go.
const DeviceFile string = "/device.json"

// WireguardConfig is a Wireguard configuration file.
//
// It's being rewrittten per location change.
const WireguardConfig = "/fvpn0.conf"

// Deprecated: The SessionFile is a file for storing the last session information.
//
// It's used to track down the status of connection.
var SessionFile = AppDir + "session.json"

var ProfilesDir = AppDir + "profiles/"

const ActiveUserLockFile string = "/.active.lock"

// FirebaseAuthFile is a file to dump Firebase responses.
const FirebaseAuthFile = "/firebase.json"

// FirebaseExtensionFile contains access token expiration date recieved from firebase login response's expire time value.
const FirebaseExtensionFile = "/firebase-ext.json"

// BillingFeatureFile is a file to store user's billing features locally.
const BillingFeatureFile = "/billing.json"

// AccountsMapFile is a file to dump the AccountsMap structure that holds the mappings from user logged in emails to uuids.
const AccountsMapFile = ".accounts.json"

// AccountsMap is a structure that helps to track down user's logins and logouts. It also helps to show local accounts in account ls command.
type AccountsMap struct {
	path string
}

// GetAccountMap is a faactory function to get the AccountsMap structure
func GetAccountsMap(accountsMapFile string) AccountsMap {
	path := AppDir + accountsMapFile
	accountmap := AccountsMap{path: path}
	return accountmap
}

// ListLocalAccounts is a method to print local user accounts in a table.
func (a AccountsMap) ListLocalAccounts() error {
	var data [][]string
	m, err := a.loadMap()

	if err != nil {
		return err
	}

	for k, v := range m {
		data = append(data, []string{k, v})
	}

	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader([]string{"Email", "UUID"})
	t.SetBorder(false)
	t.AppendBulk(data)
	t.Render()
	return nil
}

// RemoveAccount is a method to remove knonw local account from AccountMap.
// It loads AccountMap structure from the AccountMapFile, removes the key with value of user_id and dumps updated data using JsonDump.
func (a AccountsMap) RemoveAccount(user_id string) error {
	m, err := a.loadMap()

	if err != nil {
		return err
	}

	for k, v := range m {
		if v == user_id {
			delete(m, k)
		}
	}
	b, err := json.MarshalIndent(m, "", "    ")

	if err != nil {
		return err
	}

	return JsonDump(b, a.path)
}

// loadMap is a function that reads the AccountMapFile and returns unmurshalled map.
func (a AccountsMap) loadMap() (map[string]string, error) {
	m := make(map[string]string)

	if _, err := os.Stat(a.path); os.IsNotExist(err) {
		return m, err
	}

	b, err := os.ReadFile(a.path)

	if err != nil {
		return m, err
	}

	err = json.Unmarshal(b, &m)
	return m, err
}

// AddAccount is a method to add a new logged in account to the AccountMap.
func (a AccountsMap) AddAccount(email string, user_id string) error {
	m, _ := a.loadMap()
	m[email] = user_id
	b, err := json.MarshalIndent(m, "", "    ")

	if err != nil {
		return err
	}

	return JsonDump(b, a.path)
}

// GetUserID is a method to get uuid of a user by related email.
func (a AccountsMap) GetUserID(email string) string {
	m, _ := a.loadMap()
	return m[email]
}

func (a AccountsMap) GetEmail(user_id string) string {
	m, _ := a.loadMap()

	for k, v := range m {
		if v == user_id {
			return k
		}
	}

	return ""
}

// loadFirebaseAuthFile is a function to get local firebase authentication response per user by the user ID.
func loadFirebaseAuthFile(user_id string) (map[string]any, error) {
	var firebaseAuthFile map[string]any

	path := ProfilesDir + user_id + FirebaseAuthFile
	data, err := readFile(path)

	if err != nil {
		return firebaseAuthFile, err
	}

	err = json.Unmarshal(data, &firebaseAuthFile)
	return firebaseAuthFile, err
}

// LoadAccessToken is a helper function to quickly read and return access token of specific user by user ID from the firebase login response avalable locally.
func LoadAccessToken(user_id string) (string, error) {
	var accessToken string
	firebaseAuthFile, err := loadFirebaseAuthFile(user_id)

	if err != nil {
		return accessToken, err
	}

	var y interface{} = firebaseAuthFile["access_token"]
	switch v := y.(type) {
	case string:
		accessToken = v
	}

	return accessToken, err
}

// LoadUserID is a helper function to quickly read and return user ID of currently active user from the firebase login response avalable locally.
func LoadUserID() (string, error) {
	var id string
	user_id, err := loadActiveUserId()

	if err != nil {
		return id, err
	}

	auth, err := loadFirebaseAuthFile(user_id)

	if err != nil {
		return id, err
	}

	var y interface{} = auth["user_id"]
	switch v := y.(type) {
	case string:
		id = v
	}

	return id, err
}

// Init is a function that creates directories structure for Forest CLI.
func Init() error {
	for _, dir := range []string{AppDir, ProfilesDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.Mkdir(dir, 0755)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// JsonDump is a function that dumps the json data into the file at filepath.
func JsonDump(data []byte, filepath string) error {
	file, err := os.Create(filepath)

	if err != nil {
		return err
	}

	defer file.Close()
	n, err := file.WriteString(string(data))

	if err != nil {
		return err
	}

	if n != len(string(data)) {
		return fmt.Errorf("error dumping %s to %s", string(data), filepath)
	}
	return err
}

// readFile is a function that reads the content of a file at filepath
func readFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)

	if err != nil {
		return []byte(""), err
	}

	defer file.Close()
	return ioutil.ReadAll(file)

}

// handleFirebaseAuthResponse is a function to extracts the error message from the Firebase REST API response when the status is ok.
func handleFirebaseAuthResponse(response *resty.Response) (string, error) {
	var body map[string]any
	var message string

	err := json.Unmarshal(response.Body(), &body)

	if err != nil {
		return message, err
	}

	var x interface{} = body["error"]
	switch v := x.(type) {
	case map[string]interface{}:
		var x interface{} = v["message"]
		switch v := x.(type) {
		case string:
			message = v
		}
	}

	return message, nil
}

// HandleFirebaseSignInResponse is a function that dumps the Firebase REST API response into FirebaseAuthFile.
func HandleFirebaseSignInResponse(response *resty.Response) error {
	message, err := handleFirebaseAuthResponse(response)

	if err != nil {
		return err
	}

	if len(message) > 0 {
		return errors.New("invalid email or password")
	}

	return nil
}

// HandleFirebaseSignUpResponse is a function to check whether there's an error in firebase response from it's sign up endpoint.
//
// See https://firebase.google.com/docs/reference/rest/auth#section-create-email-password for more information.
func HandleFirebaseSignUpResponse(response *resty.Response) error {
	message, err := handleFirebaseAuthResponse(response)

	if err != nil {
		return err
	}

	switch message {
	case "EMAIL_EXISTS":
		return errors.New("the email address is already in use by another account")
	case "OPERATION_NOT_ALLOWED":
		return errors.New("password sign-in is disabled")
	case "TOO_MANY_ATTEMPTS_TRY_LATER":
		return errors.New("try again later")
	}

	return nil
}

// LoadRefreshToken is a function to quickly get the local refresh token from FirebaseAuthFile.
func LoadRefreshToken() (string, error) {
	var refreshToken string
	user_id, err := loadActiveUserId()

	if err != nil {
		return refreshToken, err
	}

	firebaseAuthFile, err := loadFirebaseAuthFile(user_id)

	if err != nil {
		return "", err
	}

	var y interface{} = firebaseAuthFile["refresh_token"]
	switch v := y.(type) {
	case string:
		refreshToken = v
	}

	if len(refreshToken) == 0 {
		var y interface{} = firebaseAuthFile["refreshToken"]
		switch v := y.(type) {
		case string:
			refreshToken = v
		}
	}

	return refreshToken, err
}

// IsRefreshTokenExists is a function to quickly check refresh token exists locally.
func IsRefreshTokenExists() (bool, error) {
	refreshToken, err := LoadRefreshToken()

	if err != nil {
		return false, err
	}

	return len(refreshToken) > 0, err
}

// IsDeviceCreated is a function that checks if the DeviceFile exist, i.e device is created.
func IsDeviceCreated(user_id string) bool {
	_, err := os.Stat(ProfilesDir + user_id + DeviceFile)
	return !os.IsNotExist(err)
}

// LoadDevice is a function that reads local device file depending on the user ID provided and returns it as a forestvpn_api.Device.
func LoadDevice(user_id string) (*forestvpn_api.Device, error) {
	var device *forestvpn_api.Device
	data, err := readFile(ProfilesDir + user_id + DeviceFile)

	if err != nil {
		return device, err
	}

	err = json.Unmarshal(data, &device)
	return device, err

}

// IsAuthenticated is a helper function to quickly check wether user is authenticated by checking existance of an access token.
func IsAuthenticated() bool {
	authenticated := false
	user_id, err := loadActiveUserId()

	if err != nil {
		return authenticated
	}

	accessToken, _ := LoadAccessToken(user_id)

	if len(accessToken) > 0 {
		authenticated = true
	}
	return authenticated

}

// IsActiveProfile is a function to check if the given user ID is an ID of currently active (logged in) user.
func IsActiveProfile(user_id string) bool {
	path := ProfilesDir + user_id + ActiveUserLockFile
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true

}

// SetActiveProfile is a function to set the user profile with given user ID active.
// To set the profile active:
// 1. Find another active profile if exists
// 2. Diactivate currently active profile if found on step 1.
// 3. Activate profile with id of given user_id
//
// How the profiles are activated or diactivated.
//
// Each user profile isolated in the own profile directory named by user's uuid.
// When profile X is activated:
// 1. The lock file is removed from other profiles directories if found.
// 2. The lock file is created in the directory of X profile.
//
// ActiveUserLockFile - is a lock file. It's an empty file that only serves to indicate whether profile is active.
// When profile X is diactivated, the lock file is removed from the X profile directory.
func SetActiveProfile(user_id string) error {
	files, err := ioutil.ReadDir(ProfilesDir)

	if err != nil {
		return err
	}

	removed := false
	created := false

	for _, f := range files {
		if f.IsDir() {
			path := ProfilesDir + f.Name() + ActiveUserLockFile

			if _, err := os.Stat(path); !removed && !os.IsNotExist(err) {
				err := os.Remove(path)

				if err != nil {
					sentry.CaptureException(err)
				}
				removed = true
			} else if !created && f.Name() == user_id {
				_, err := os.Create(ProfilesDir + f.Name() + ActiveUserLockFile)

				if err != nil {
					return err
				}

				created = true
			} else if created && removed {
				break
			}
		}
	}
	return nil
}

// RemoveFirebaseAuthFile is a function to remove FirebaseAuthFile of the user with given user ID.
// Used to log out the user.
func RemoveFirebaseAuthFile(user_id string) error {
	path := ProfilesDir + user_id + FirebaseAuthFile
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		err = os.Remove(path)

		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveActiveUserLockFile is a helper function to quickly diactivate currently active (logged in) user.
func RemoveActiveUserLockFile() error {
	user_id, err := loadActiveUserId()

	if err != nil {
		return err
	}

	path := ProfilesDir + user_id + ActiveUserLockFile

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		err = os.Remove(path)

		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateProfileDevice is a helper function to quickly update the local device file of the logged in (active) user.
func UpdateProfileDevice(device *forestvpn_api.Device) error {
	data, err := json.MarshalIndent(device, "", "    ")

	if err != nil {
		return err
	}

	user_id, err := LoadUserID()

	if err != nil {
		return err
	}

	return JsonDump(data, ProfilesDir+user_id+DeviceFile)
}

// LoadIdToken is a function to load id token of the user with id value of given user id from the local firebase login response.
func LoadIdToken(user_id string) (string, error) {
	var idToken string
	firebaseAuthFile, err := loadFirebaseAuthFile(user_id)

	if err != nil {
		return idToken, err
	}

	var y interface{} = firebaseAuthFile["id_token"]
	switch v := y.(type) {
	case string:
		idToken = v
	}

	return idToken, err
}

// loadActiveUserId is a helper function to quickly read and return an id of a currently active (logged in) user.
func loadActiveUserId() (string, error) {
	var user_id string
	files, err := ioutil.ReadDir(ProfilesDir)

	if err != nil {
		return user_id, err
	}

	for _, profiledir := range files {
		if profiledir.IsDir() {
			path := ProfilesDir + profiledir.Name()
			files, err := ioutil.ReadDir(path)

			if err != nil {
				return user_id, err
			}

			for _, file := range files {
				if file.Name() == ActiveUserLockFile[1:] {
					user_id = profiledir.Name()
				}
			}
		}
	}

	return user_id, nil
}

// getAccessTokenExpireDate is a function to create a time.Time object from numeric string value.
// The numeric string values is a value that comes from the firebase login response and indicates the duration of an access token exparation period in seconds.
// The numeric string value is added to the current local time right after the firebase login response.
// Thus, getAccessTokenExpireDate returns the time object with exparation date of an access token.
func getAccessTokenExpireDate(expireTime string) (time.Time, error) {
	now := time.Now()
	h, err := time.ParseDuration(expireTime + "s")

	if err != nil {
		return now, err
	}

	return now.Add(time.Second + h), nil
}

// Date2Json is a function that marshals time.Time type into json string.
func Date2Json(date time.Time) ([]byte, error) {
	expireDate := make(map[string]string)
	expireDate["expireDate"] = date.Format(time.RFC3339)
	return json.MarshalIndent(expireDate, "", "    ")
}

// loadAccessTokenExpireDate is a function to load the exparation date of an access token for user with id value of given user id.
func loadAccessTokenExpireDate(user_id string) (time.Time, error) {
	var expireTime time.Time
	data, err := loadFirebaseExtensionFile(user_id)

	if err != nil {
		return expireTime, err
	}

	expireTime, err = time.Parse(time.RFC3339, data["expireDate"])
	return expireTime, err
}

// IsAccessTokenExpired is a function that reads an access token exparation date of the user with id value of given user id from the local firebase sign in response.
func IsAccessTokenExpired(user_id string) (bool, error) {
	expired := false
	now := time.Now()
	expireDate, err := loadAccessTokenExpireDate(user_id)

	if err != nil {
		return expired, err
	}

	if now.After(expireDate) {
		expired = true
	}
	return expired, nil
}

// DumpAccessTokenExpireDate is a function to write access token exparation date into the FirebaseExtensionFile for the user with id value of given user id.
func DumpAccessTokenExpireDate(user_id string, expires_in string) error {
	expireTime, err := getAccessTokenExpireDate(expires_in)

	if err != nil {
		return err
	}

	data, err := Date2Json(expireTime)

	if err != nil {
		return err
	}

	path := ProfilesDir + user_id + FirebaseExtensionFile
	return JsonDump(data, path)
}

// LoadBillingFeatures is a function to read local billing features from file for the user with id value of given user id.
func LoadBillingFeatures(user_id string) ([]forestvpn_api.BillingFeature, error) {
	var billingFeatures []forestvpn_api.BillingFeature
	path := ProfilesDir + user_id + BillingFeatureFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return billingFeatures, err
	}

	data, err := readFile(path)

	if err != nil {
		return billingFeatures, err
	}

	err = json.Unmarshal(data, &billingFeatures)

	if err != nil {
		return billingFeatures, err
	}

	return billingFeatures, nil
}

// BillingFeautureExists is a helper function to quickly check whether the local billing features exists for the user with id value of given user id.
func BillingFeautureExists(user_id string) bool {
	path := ProfilesDir + user_id + BillingFeatureFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// BillingFeatureExpired is a function to check whether the expire date of given billing feature is before current local time.
func BillingFeatureExpired(billingFeature forestvpn_api.BillingFeature) bool {
	return time.Now().After(billingFeature.GetExpiryDate())
}

// loadFirebaseExtensionFile is a function to read FirebaseExtensionFile and return an access token expire date it contains for the user with id value of given user id.
func loadFirebaseExtensionFile(user_id string) (map[string]string, error) {
	var data map[string]string
	path := ProfilesDir + user_id + FirebaseExtensionFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return data, err
	}

	d, err := readFile(path)

	if err != nil {
		return data, err
	}

	err = json.Unmarshal(d, &data)
	return data, err
}
