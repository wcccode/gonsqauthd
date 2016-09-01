package main

import (
	"flag"
	"github.com/wcccode/gonsqauthd/nsqauth"
	"log"
	"os"
)

var (
	port      = flag.Int("port", 8000, "http listen port")
	auth_file = flag.String("auth_file", "", "auth file path")
)

func main() {
	flag.Parse()

	logger := log.New(os.Stderr, "[nsqauthd] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	opts := &nsqauth.Options{Port: *port, Log: logger, AuthFilePath: *auth_file}

	nsqAuthd := nsqauth.NewNsqAuthd(opts)
	nsqAuthd.Main()
}
