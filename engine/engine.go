package engine

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/server"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/http"
	"github.com/obnahsgnaw/http/engine"
)

type MuxHttp struct {
	e   *http.Http
	mux *runtime.ServeMux
}

func New(e *http.Http, mux *runtime.ServeMux) *MuxHttp {
	return &MuxHttp{
		e:   e,
		mux: mux,
	}
}

func Default(host url.Host, cnf *engine.Config) (*MuxHttp, error) {
	if e, err := http.Default(host.Ip, host.Port, cnf); err != nil {
		return nil, err
	} else {
		return New(e, server.NewMux()), nil
	}
}

func (s *MuxHttp) Http() *http.Http {
	return s.e
}

func (s *MuxHttp) Mux() *runtime.ServeMux {
	return s.mux
}
