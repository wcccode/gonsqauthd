package nsqauth

import (
	"encoding/csv"
	"errors"
	"os"
)

var FileNotFound = errors.New("file not fount")

type Auth struct {
	Ttl            int            `json: "ttl"`
	Identity       string         `json: "identity"`
	IdentityUrl    string         `json: "identity_url"`
	Authorizations Authorizations `json: "authorizations"`
}

type Authorizations struct {
	Permissions []string `json: "permissions"`
	Topic       string   `json: "topic"`
	Channels    []string `json: "channels"`
}

//auth db
type AuthDb struct {
	records [][]string
}

func NewAuthDb(filePath string) *AuthDb {
	db := &AuthDb{}
	err := db.init(filePath)
	if err != nil {
		panic(err)
	}
	return db
}

func (db *AuthDb) init(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return FileNotFound
		}
		return err
	}

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	db.records = records
	return nil
}

func (db *AuthDb) Get(login, ip, tlsRequired string) *Auth {
	auth := &Auth{Ttl: 3600, Identity: "nsqauthd", IdentityUrl: ""}

	for _, elm := range db.records {
		if len(elm) < 3 {
			continue
		}
		if login == elm[0] && ip == elm[1] && tlsRequired == elm[2] {
			return elm
		}
	}
	return auth
}

func (db *AuthDb) List() [][]string {
	return db.records
}
