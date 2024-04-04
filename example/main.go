package main

import (
	"github.com/obnahsgnaw/api"
	"github.com/obnahsgnaw/api/service/apidoc"
	"github.com/obnahsgnaw/application"
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/logging/logger"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/application/service/regCenter"
	engine2 "github.com/obnahsgnaw/http/engine"
	"time"
)

func main() {
	app := application.New("demo")
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
	r, _ := regCenter.NewEtcdRegister([]string{"127.0.0.1:2379"}, time.Second*5)
	app.With(application.Register(r, 5))

	//jwt.SetKeyPrefix()

	e, _ := api.NewEngine(app, "auth", endtype.Backend, url.Host{Ip: "127.0.0.1", Port: 8001}, &engine2.Config{})
	s := api.New(app, e, "auth", "auth", endtype.Backend, 1,
		api.RpcServer(),
		api.ErrCodePrefix(1),
		api.DocServer(&apidoc.Config{
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
		}),
	)

	app.AddServer(s)

	app.Run(func(err error) {
		panic(err)
	})

	app.Wait()
}
