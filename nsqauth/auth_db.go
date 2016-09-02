package nsqauth

import (
	"encoding/csv"
	"errors"
	"os"
)

var FileNotFound = errors.New("file not fount")

// AuthDb represents auth's info data source
type AuthDb struct {
	entries [][]string
}

// new AuthDb and initiation entries
func NewAuthDb(filePath string) *AuthDb {
	db := &AuthDb{}
	err := db.init(filePath)
	if err != nil {
		panic(err)
	}
	return db
}

// init file and read entries
// file content is
// login,ip,tls_required,topic,channel,subscribe,publish
func (db *AuthDb) init(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return FileNotFound
		}
		return err
	}

	r := csv.NewReader(f)
	entries, err := r.ReadAll()
	if err != nil {
		return err
	}

	db.entries = entries
	return nil
}

func (db *AuthDb) Get(login, ip, tlsRequired string) []string {
	for _, elm := range db.entries {
		if len(elm) < 3 {
			continue
		}
		if login == elm[0] && ip == elm[1] && tlsRequired == elm[2] {
			return elm
		}
	}
	return nil
}

func (db *AuthDb) List() [][]string {
	return db.entries
}
