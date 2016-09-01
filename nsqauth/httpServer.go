package nsqauth

import (
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

	router.Handle("GET", "/", Decorate(s.index, log(ctx.nsqAuthd.Opts.Log)))
	router.Handle("GET", "/auth", Decorate(s.auth, log(ctx.nsqAuthd.Opts.Log)))

	return s
}

func (s *httpServer) index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprint(w, "Welcome to nsqauthd!")
}

func (s *httpServer) auth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	secret := r.FormValue("secret")
	ip := r.FormValue("remote_ip")
	tls := r.FormValue("tls")
	s.ctx.nsqAuthd.Opts.Log.Output(2, fmt.Sprintf("%v, %v, %v", secret, ip, tls))
	auths := s.ctx.nsqAuthd.Db.Get(secret, ip, tls)
	fmt.Fprint(w, auths)
}

//api handler
type ApiHandler func(w http.ResponseWriter, r *http.Request, ps httprouter.Params)

//api handler decorater
type Decorater func(ApiHandler) ApiHandler

func Decorate(apiHandler ApiHandler, decoraters ...Decorater) httprouter.Handle {
	handler := apiHandler
	for _, decorater := range decoraters {
		handler = decorater(handler)
	}
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		handler(w, r, ps)
	}
}

func log(log Logger) Decorater {
	return func(handler ApiHandler) ApiHandler {
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			log.Output(2, "start handler")
			handler(w, r, ps)
			log.Output(2, "end handler")
		}
	}
}
