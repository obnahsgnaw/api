package main

import (
	"github.com/obnahsgnaw/api"
	"github.com/obnahsgnaw/api/service/apidoc"
	"github.com/obnahsgnaw/application"
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/url"
)

func main() {
	app := application.New(application.NewCluster("dev", "Dev"), "apiDemo")
	defer app.Release()

	app.With(application.Debug(func() bool {
		return true
	}))

	//errhandler.SetDefaultDebugger()
	//errhandler.SetDefaultErrObjProvider()
	//jwt.SetKeyPrefix()

	s := api.New(app, "auth", "auth", endtype.Backend, "/auth", url.Host{Ip: "127.0.0.1", Port: 8001}, 1)

	s.WithRpcServer(8002, true)

	s.WithDocService(&apidoc.Config{
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
	})

	app.AddServer(s)

	app.Run(func(err error) {
		panic(err)
	})

	app.Wait()
}
