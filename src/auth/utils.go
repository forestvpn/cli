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

	"github.com/fatih/color"
	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/getsentry/sentry-go"
	"github.com/go-resty/resty/v2"
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

const FirebaseExtensionFile = "/firebase-ext.json"
const BillingFeatureFile = "/billing.json"

type AccountsMap struct {
	path string
}

func GetAccountMap(accountsMapFile string) AccountsMap {
	path := AppDir + accountsMapFile
	accountmap := AccountsMap{path: path}
	return accountmap
}

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

func (a AccountsMap) AddAccount(email string, user_id string) error {
	m, _ := a.loadMap()
	m[email] = user_id
	b, err := json.MarshalIndent(m, "", "    ")

	if err != nil {
		return err
	}

	return JsonDump(b, a.path)
}

func (a AccountsMap) SetEmail(email string) error {
	m, err := a.loadMap()

	if err != nil {
		return err
	}

	m[email] = ""
	b, err := json.Marshal(m)

	if err != nil {
		return err
	}

	return JsonDump(b, a.path)
}

func (a AccountsMap) SetUserID(email string, user_id string) error {
	m, err := a.loadMap()

	if err != nil {
		return err
	}

	m[email] = user_id

	b, err := json.Marshal(m)

	if err != nil {
		return err
	}

	return JsonDump(b, a.path)
}

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

// Deprecated
func AddProfile(response *resty.Response, device *forestvpn_api.Device, activate bool) error {
	jsonresponse := make(map[string]string)
	err := json.Unmarshal(response.Body(), &jsonresponse)

	if err != nil {
		return err
	}

	user_id := jsonresponse["user_id"]
	path := ProfilesDir + user_id

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)

		if err != nil {
			return err
		}
	}

	err = JsonDump(response.Body(), path+FirebaseAuthFile)

	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(device, "", "    ")

	if err != nil {
		return err
	}

	err = JsonDump(data, path+DeviceFile)

	if err != nil {
		return err
	}

	if activate {
		err = SetActiveProfile(user_id)
	}
	return err
}

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

func LoadDevice(user_id string) (*forestvpn_api.Device, error) {
	var device *forestvpn_api.Device
	data, err := readFile(ProfilesDir + user_id + DeviceFile)

	if err != nil {
		return device, err
	}

	err = json.Unmarshal(data, &device)
	return device, err

}

// BuyPremiumDialog is a function that prompts the user to by premium subscrition on Forest VPN.
// If the user prompts 'yes', then it opens https://forestvpn.com/pricing/ page in the default browser.
func BuyPremiumDialog(city string) error {
	// var answer string
	// var openCommand string
	const url = "https://forestvpn.com/checkout/"
	// os := runtime.GOOS
	// switch os {
	// case "windows":
	// 	openCommand = "start"
	// case "darwin":
	// 	openCommand = "open"
	// case "linux":
	// 	openCommand = "xdg-open"
	// }

	faint := color.New(color.Faint)
	fmt.Println("How to use FREE version")
	fmt.Println()
	fmt.Println("1. Install ForestVPN mobile app")
	faint.Println("You can install it through out App Store or Google Play")
	fmt.Println("2. Watch ad in mobile app")
	faint.Println("Account in mobile app and here should be same as:- account@example.com")
	fmt.Println("3. Connect to VPN")
	faint.Println("30 minutes connection in exchange for 30 seconds Ad")
	fmt.Println()
	fmt.Println("OR")
	fmt.Println()
	// faint.Println(fmt.Sprintf("%s availble for premium plan subscribers.", city))
	fmt.Printf("Go Premium: %s", url)
	fmt.Println()
	// fmt.Scanln(&answer)

	// if strings.Contains("YESyesYesYEsyEsyeSyES", answer) {
	// 	err := exec.Command(openCommand, url).Run()

	// 	if err != nil {
	// 		fmt.Println(url)
	// 	}
	// }
	return nil
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

// Deprecated: IsLocationSet is a function to check wether Wireguard configuration file created after location selection.
func IsLocationSet() bool {
	_, err := os.Stat(WireguardConfig)
	return !os.IsNotExist(err)
}

func ProfileExists(user_id string) bool {
	if _, err := os.Stat(ProfilesDir + user_id); os.IsNotExist(err) {
		return false
	}

	return true
}

func IsActiveProfile(user_id string) bool {
	path := ProfilesDir + user_id + ActiveUserLockFile
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true

}

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

func GetAccessTokenExpireDate(expireTime string) (time.Time, error) {
	now := time.Now()
	h, err := time.ParseDuration(expireTime + "s")

	if err != nil {
		return now, err
	}

	return now.Add(time.Second + h), nil
}

func Date2Json(date time.Time) ([]byte, error) {
	expireDate := make(map[string]string)
	expireDate["expireDate"] = date.Format(time.RFC3339)
	return json.MarshalIndent(expireDate, "", "    ")
}

func loadAccessTokenExpireDate(user_id string) (time.Time, error) {
	var expireTime time.Time
	data, err := loadFirebaseExtensionFile(user_id)

	if err != nil {
		return expireTime, err
	}

	expireTime, err = time.Parse(time.RFC3339, data["expireDate"])
	return expireTime, err
}

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

func DumpAccessTokenExpireDate(user_id string, expires_in string) error {
	expireTime, err := GetAccessTokenExpireDate(expires_in)

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

func BillingFeautureExists(user_id string) bool {
	path := ProfilesDir + user_id + BillingFeatureFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func BillingFeatureExpired(billingFeature forestvpn_api.BillingFeature) bool {
	return time.Now().After(billingFeature.GetExpiryDate())
}

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
