package nsqauth

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
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
	tls := r.FormValue("tls")
	s.ctx.nsqAuthd.Opts.Log.Output(2, fmt.Sprintf("auth request with secret: %v, remote_ip: %v, tls: %v", secret, ip, tls))

	entries := s.ctx.nsqAuthd.Db.Get(secret, ip, tls)
	auths := make([]authorizations, 0, 1)
	if entries != nil {
		auths = append(auths, authorizations{Topic: entries[3], Channels: entries[4:5], Permissions: entries[5:]})
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
