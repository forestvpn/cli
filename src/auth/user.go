package auth

import (
	"context"
	"encoding/json"
	"fmt"
	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/forestvpn/cli/api"
	"github.com/forestvpn/goauthlib/pkg/svc"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// AccountsMapFile is a file to dump the AccountsMap structure that holds the mappings from user logged in emails to uuids.
const AccountsMapFile = ".accounts.json"

type ProfilePK string
type ProfileID string
type ProfileEmail string

type Profile struct {
	Email    ProfileEmail
	ID       ProfileID
	LastSeen int64
	Active   bool
	Pk       ProfilePK
	db       *UserDB
}

func (p *Profile) Touch() {
	p.LastSeen = time.Now().Unix()
	p.Save()
}

func (p *Profile) MarkAsActive() {
	p.Active = true
	p.Save()
}

func (p *Profile) MarkAsInactive() {
	p.Active = false
	p.Save()
}

func (p *Profile) Save() {
	path := filepath.Join(AppDir + "/profiles")
	_ = os.Mkdir(path, 0755)
	path = filepath.Join(path, string(p.ID))
	_ = os.Mkdir(path, 0755)

	p.db.Users[p.Pk] = p
	p.db.persist()
}

func (p *Profile) Token() (svc.Token, error) {
	return AuthService(string(p.Pk)).GetToken(context.Background())
}

func (p *Profile) ApiClient(apiHost string) *api.ApiClientWrapper {
	token, err := p.Token()
	if err != nil {
		log.Fatalf("failed to get token for user %s: %v", p.Pk, err)
	}
	return api.GetApiClient(token.Raw(), apiHost)
}

func (p *Profile) SignIn(apiHost string) error {
	token, err := p.Token()
	if err != nil {
		return err
	}

	if p.Email == "" {
		// Create a new context with the token as the access token
		authCtx := context.WithValue(context.Background(), forestvpn_api.ContextAccessToken, token.Raw())
		// Make a request to the WhoAmI endpoint
		userInfo, _, loginErr := api.GetApiClient(token.Raw(), apiHost).APIClient.AuthApi.WhoAmI(authCtx).Execute()
		// If there is an error, log it and return it
		if loginErr != nil {
			fmt.Println(token.Raw())
			return loginErr
		}
		p.ID, p.Email = ProfileID(userInfo.GetId()), ProfileEmail(userInfo.GetEmail())
		p.Touch()
		p.MarkAsActive()
	}

	return nil
}

func (p *Profile) DB() *UserDB {
	return p.db
}

type UserDB struct {
	path    string
	current ProfilePK
	Users   map[ProfilePK]*Profile
}

func (db *UserDB) CurrentUser() *Profile {
	db.Sync()
	p := db.Users[db.current]
	if p == nil {
		p = db.CreateUser()
	}
	p.db = db
	return p
}

func (db *UserDB) ListUsers() []*Profile {
	var users []*Profile
	for _, profile := range db.Users {
		if profile.Active {
			profile.db = db
			users = append(users, profile)
		}
	}
	return users
}

func (db *UserDB) CreateUser() *Profile {
	profile := &Profile{Pk: ProfilePK(uuid.New().String()), db: db}
	profile.db = db
	db.Users[profile.Pk] = profile // TODO: make this assignment not a pointer
	db.current = profile.Pk
	profile.Touch()
	db.persist()
	return profile
}

func (db *UserDB) persist() {
	data, err := json.Marshal(db)
	if err != nil {
		log.Fatalf("failed to marshal Users to json: %v", err)
	}

	err = os.WriteFile(db.path, data, 0644)
	if err != nil {
		log.Fatalf("failed to write to accounts map file %s: %v", db.path, err)
	}
}

func (db *UserDB) Sync() *UserDB {
	data, err := os.ReadFile(db.path)
	if err != nil {
		log.Fatalf("failed to read accounts map file %s: %v", db.path, err)
	}

	err = json.Unmarshal(data, &db)
	if err != nil {
		log.Fatalf("failed to unmarshal Users from json: %v", err)

	}

	// if there is no user with last seen time in 30 days or more, create one
	for _, user := range db.Users {
		if !user.Active {
			continue
		}
		if time.Now().Unix()-user.LastSeen > 30*24*60*60 {
			user.MarkAsInactive()
			continue
		}
		if db.current == "" {
			db.current = user.Pk
			continue
		}
		if user.LastSeen > db.Users[db.current].LastSeen {
			db.current = user.Pk
		}
	}

	return db
}

func OpenUserDB() *UserDB {
	path := filepath.Join(AppDir, AccountsMapFile)
	_ = os.Mkdir(AppDir, 0755)
	if fileInfo, err := os.Stat(path); err == nil {
		if fileInfo.IsDir() {
			log.Fatalf("accounts map file %s is a directory", path)
		}
	} else {
		// file does not exist, create it
		f, fErr := os.Create(path)
		if fErr != nil {
			log.Fatalf("failed to create accounts map file %s: %v", path, fErr)
		}

		_, _ = f.Write([]byte("{\"Users\":{}}"))
		err = f.Close()
		if err != nil {
			log.Fatalf("failed to stat accounts map file %s: %v", path, err)
		}
	}

	db := &UserDB{path: path, Users: map[ProfilePK]*Profile{}}
	db.Sync()
	return db
}
