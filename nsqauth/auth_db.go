package nsqauth

import (
	"encoding/csv"
	"errors"
	"os"
)

var ErrFileNotFound = errors.New("file not fount")

// AuthDb represents auth's info data source
type AuthDb struct {
	entries []Entry
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
			return ErrFileNotFound
		}
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	if len(records) > 0 {
		records = records[1:]
	}

	entries := make([]Entry, 0, 10)
	for _, record := range records {
		tlsRequired := false
		if record[2] == "true" {
			tlsRequired = true
		}
		entry := Entry{Login: record[0], Ip: record[1], TlsRequired: tlsRequired, Topic: record[3], Channel: record[4], Subscribe: record[5], Publish: record[6]}
		entries = append(entries, entry)
	}
	db.entries = entries
	return nil
}

func (db *AuthDb) Get(login, ip string, tlsRequired bool) []Entry {
	entries := make([]Entry, 0, 1)
	for _, entry := range db.entries {
		if entry.Login != "" && entry.Login != login {
			continue
		}
		if entry.Ip != "" && entry.Ip != ip {
			continue
		}
		if entry.TlsRequired && !tlsRequired {
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}

func (db *AuthDb) List() []Entry {
	return db.entries
}
