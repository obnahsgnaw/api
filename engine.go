package api

import (
	"github.com/obnahsgnaw/api/engine"
	"github.com/obnahsgnaw/application"
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/application/pkg/utils"
	engine2 "github.com/obnahsgnaw/http/engine"
)

func NewEngine(app *application.Application, id string, et endtype.EndType, host url.Host, cnf *engine2.Config) (*engine.MuxHttp, error) {
	cnf.Name = utils.ToStr(et.String(), "-", id)
	cnf.LogCnf = logCnf(app, et, id)
	return engine.Default(host, cnf)
}
