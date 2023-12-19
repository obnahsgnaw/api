package main

import (
	"github.com/obnahsgnaw/api"
	"github.com/obnahsgnaw/api/service/apidoc"
	"github.com/obnahsgnaw/application"
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/logging/logger"
	"github.com/obnahsgnaw/application/pkg/url"
	"time"
)

func main() {
	app := application.New(application.NewCluster("dev", "Dev"), "demo")
	defer app.Release()

	app.With(application.Debug(func() bool {
		return true
	}))

	app.With(application.Logger(&logger.Config{
		Dir:        "/Users/wangshanbo/Documents/Data/projects/api/out",
		MaxSize:    5,
		MaxBackup:  1,
		MaxAge:     1,
		Level:      "debug",
		TraceLevel: "error",
	}))
	app.With(application.EtcdRegister([]string{"127.0.0.1:2379"}, time.Second*5))

	//errhandler.SetDefaultDebugger()
	//errhandler.SetDefaultErrObjProvider()
	//jwt.SetKeyPrefix()

	s := api.New(app, "auth", "auth", endtype.Backend, "/auth", 1)
	s.With(api.NewEngine(url.Host{Ip: "127.0.0.1", Port: 8001}))
	s.With(api.RpcInsOrNew(nil, url.Host{Ip: "127.0.0.1", Port: 8002}, true))

	s.With(api.Doc(&apidoc.Config{
		Protocol: url.HTTP,
		Path:     "/doc",
		Title:    "认证",
		Provider: func() ([]byte, error) {
			return []byte("ok"), nil
		},
		DebugOrigin: url.Origin{
			Protocol: url.HTTP,
			Host:     url.Host{Ip: "127.0.0.1", Port: 8001},
		},
		EndType: endtype.Backend,
	}))

	app.AddServer(s)

	app.Run(func(err error) {
		panic(err)
	})

	app.Wait()
}
