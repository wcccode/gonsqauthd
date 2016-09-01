package nsqauth

import (
	"fmt"

	"net/http"
)

type NsqAuthd struct {
	Opts *Options
	Db   *AuthDb
}

func NewNsqAuthd(opts *Options) *NsqAuthd {
	return &NsqAuthd{
		Opts: opts,
		Db:   NewAuthDb(opts.AuthFilePath),
	}
}

func (n *NsqAuthd) Main() {
	ctx := &Context{n}
	server := newHttpServer(ctx)

	n.Opts.Log.Output(2, fmt.Sprintf("http server listen to port: %v", n.Opts.Port))
	http.ListenAndServe(fmt.Sprintf(":%v", n.Opts.Port), server.router)
}
