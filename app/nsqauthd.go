package main

import (
	"flag"
	"github.com/wcccode/gonsqauthd/nsqauth"
	"log"
	"os"
)

var (
	port      = flag.Int("port", 8000, "http listen port")
	auth_file = flag.String("auth_file", "auth_data.csv", "auth file path")
	ttl       = flag.Int("ttl", 3600, "time to leave")
)

func main() {
	flag.Parse()

	logger := log.New(os.Stderr, "[nsqauthd] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	opts := &nsqauth.Options{Port: *port, Log: logger, AuthFilePath: *auth_file, Ttl: *ttl}

	nsqAuthd := nsqauth.NewNsqAuthd(opts)
	nsqAuthd.Main()
}
