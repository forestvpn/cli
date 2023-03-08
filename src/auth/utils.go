// Authentication related utilities around firebase REST authentication workflow.
// See https://firebase.google.com/docs/reference/rest for more information.
package auth

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	forestvpn_api "github.com/forestvpn/api-client-go"
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

var ProfilesDir = AppDir + "profiles/"

// BillingFeatureFile is a file to store user's billing features locally.
const BillingFeatureFile = "/billing.json"

func LoadUserID() (string, error) {
	userId, _ := AuthStore.Load("last_id")
	return userId, nil
}

// Init is a function that creates directories structure for Forest CLI.
func Init() error {
	dirs := []string{AppDir, ProfilesDir}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
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

	n, err := file.Write(data)
	if err != nil {
		return err
	}

	if n != len(data) {
		return fmt.Errorf("error dumping %s to %s", string(data), filepath)
	}

	return nil
}

// readFile is a function that reads the content of a file at filepath
func readFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// LoadDevice is a function that reads local device file depending on the user ID provided and returns it as a forestvpn_api.Device.
func LoadDevice(userID ProfileID) (*forestvpn_api.Device, error) {
	var device *forestvpn_api.Device
	data, err := readFile(ProfilesDir + string(userID) + DeviceFile)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &device); err != nil {
		return nil, err
	}

	return device, nil
}

// UpdateProfileDevice is a helper function to quickly update the local device file of the logged in (active) user.
func UpdateProfileDevice(device *forestvpn_api.Device, userID ProfileID) error {
	data, err := json.MarshalIndent(device, "", "    ")
	if err != nil {
		return err
	}

	return JsonDump(data, ProfilesDir+string(userID)+DeviceFile)
}

// LoadBillingFeatures is a function to read local billing features from file for the user with id value of given user id.
func LoadBillingFeatures(userID ProfileID) ([]forestvpn_api.BillingFeature, error) {
	var billingFeatures []forestvpn_api.BillingFeature
	path := ProfilesDir + string(userID) + BillingFeatureFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	data, err := readFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &billingFeatures); err != nil {
		return nil, err
	}

	return billingFeatures, nil
}

// BillingFeautureExists is a helper function to quickly check whether the local billing features exists for the user with id value of given user id.
func BillingFeautureExists(userID ProfileID) bool {
	path := filepath.Join(ProfilesDir, string(userID), BillingFeatureFile)
	fStat, _ := os.Stat(path)
	return fStat != nil && fStat.Size() > 0
}

// BillingFeatureExpired is a function to check whether the expire date of given billing feature is before current local time.
func BillingFeatureExpired(billingFeature forestvpn_api.BillingFeature) bool {
	return time.Now().After(*billingFeature.ExpiryDate)
}

// PrintLocalAccounts is a method to print local user accounts in a table.
func PrintLocalAccounts() error {
	db := OpenUserDB()
	data := [][]string{}
	current := db.CurrentUser()
	for _, v := range db.ListUsers() {
		var mark = ""
		if v.Pk == current.Pk {
			mark = "*"
		}
		data = append(data, []string{mark, string(v.Email), string(v.ID)})
	}

	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader([]string{"IsActive", "Email", "UUID"})
	t.SetBorder(false)
	t.AppendBulk(data)
	t.Render()

	return nil
}
