package nsqauth

import (
	"fmt"
	"testing"
)

var filePath = "auth_testdata.csv"

var getTable = []struct {
	Login       string
	Ip          string
	TlsRequired string
}{
	{"", "127.0.0.3", "false"},
	{"test_local", "127.0.0.2", "true"},
}

func TestAuthDbList(t *testing.T) {
	db := NewAuthDb(filePath)
	fmt.Println(db.List())
}

func TestAuthDbGet(t *testing.T) {
	db := NewAuthDb(filePath)
	for _, elm := range getTable {
		r := db.Get(elm.Login, elm.Ip, elm.TlsRequired)
		fmt.Println(r)
	}
}
