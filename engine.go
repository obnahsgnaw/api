package api

import (
	"github.com/obnahsgnaw/api/engine"
	"github.com/obnahsgnaw/application"
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/logging/logger"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/application/pkg/utils"
	engine2 "github.com/obnahsgnaw/http/engine"
)

func NewEngine(app *application.Application, id string, et endtype.EndType, host url.Host, cnf *engine2.Config) (e *engine.MuxHttp, err error) {
	cnf.Name = utils.ToStr(et.String(), "-", id)
	if cnf.AccessWriter == nil {
		cnf.AccessWriter, err = logger.NewDefAccessWriter(app.LogConfig(), func() bool {
			return app.Debugger().Debug()
		})
		if err != nil {
			return nil, err
		}
	}
	if cnf.ErrWriter == nil {
		cnf.ErrWriter, err = logger.NewDefErrorWriter(app.LogConfig(), func() bool {
			return app.Debugger().Debug()
		})
		if err != nil {
			return nil, err
		}
	}
	return engine.Default(host, cnf)
}
