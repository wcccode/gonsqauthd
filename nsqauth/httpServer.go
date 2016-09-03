package nsqauth

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"strings"
)

type httpServer struct {
	ctx    *Context
	router http.Handler
}

func newHttpServer(ctx *Context) *httpServer {
	router := httprouter.New()
	s := &httpServer{ctx, router}

	router.Handle("GET", "/", Decorate(s.index, log(ctx.nsqAuthd.Opts.Log), response()))
	router.Handle("GET", "/auth", Decorate(s.auth, log(ctx.nsqAuthd.Opts.Log), response()))

	return s
}

// Error represent httpServer server's error
type Error struct {
	Code int
	Text string
}

func (e Error) Error() string {
	return e.Text
}

func (s *httpServer) index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (interface{}, error) {
	return "Welcome to nsqauthd!", nil
}

// authInfo represent a consumer's authorizations
type authInfo struct {
	Ttl            int              `json:"ttl"`
	Identity       string           `json:"identity"`
	IdentityUrl    string           `json:"identity_url"`
	Authorizations []authorizations `json:"authorizations"`
}

type authorizations struct {
	Permissions []string `json:"permissions"`
	Topic       string   `json:"topic"`
	Channels    []string `json:"channels"`
}

func (s *httpServer) auth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (interface{}, error) {
	secret := r.FormValue("secret")
	ip := r.FormValue("remote_ip")
	tls, err := strconv.ParseBool(r.FormValue("tls"))

	if err != nil {
		return nil, Error{Code: 500, Text: err.Error()}
	}

	s.ctx.nsqAuthd.Opts.Log.Output(2, fmt.Sprintf("auth request with secret: %v, remote_ip: %v, tls: %v", secret, ip, tls))

	entries := s.ctx.nsqAuthd.Db.Get(secret, ip, tls)
	auths := make([]authorizations, 0, 1)
	for _, entry := range entries {
		permiss := make([]string, 0, 2)
		if entry.Subscribe != "" {
			permiss = append(permiss, entry.Subscribe)
		}
		if entry.Publish != "" {
			permiss = append(permiss, entry.Publish)
		}
		auths = append(auths, authorizations{Topic: entry.Topic, Channels: strings.Split(entry.Channel, ","), Permissions: permiss})
	}
	return &authInfo{Ttl: s.ctx.nsqAuthd.Opts.Ttl, Identity: "nsqauthd", IdentityUrl: "", Authorizations: auths}, nil
}

// api handler function type
type ApiHandler func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (interface{}, error)

// api handler decorater
type Decorater func(ApiHandler) ApiHandler

func Decorate(f ApiHandler, decoraters ...Decorater) httprouter.Handle {
	decorated := f
	for _, decorater := range decoraters {
		decorated = decorater(decorated)
	}
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		decorated(w, r, ps)
	}
}

func log(log Logger) Decorater {
	return func(f ApiHandler) ApiHandler {
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (interface{}, error) {
			result, err := f(w, r, ps)
			code := 200
			if err != nil {
				code = err.(Error).Code
			}
			log.Output(2, fmt.Sprintf("%d %v %v %v", code, r.Method, r.RequestURI, r.RemoteAddr))
			return result, err
		}
	}
}

func response() Decorater {
	return func(f ApiHandler) ApiHandler {
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (interface{}, error) {
			result, err := f(w, r, ps)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.(Error).Error()))
				return nil, nil
			}

			var response []byte
			code := 200
			isJson := false

			switch data := result.(type) {
			case string:
				response = []byte(data)
			case []byte:
				response = data
			case nil:
				response = []byte{}
			default:
				resJson, merr := json.Marshal(data)
				response = resJson
				isJson = true
				if merr != nil {
					code = 500
					isJson = false
					response = []byte(merr.Error())
				}
			}

			if isJson {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
			}
			w.WriteHeader(code)
			w.Write(response)
			return nil, nil
		}
	}
}
