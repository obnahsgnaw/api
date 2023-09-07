package main

import (
	"github.com/obnahsgnaw/api"
	"github.com/obnahsgnaw/api/service/apidoc"
	"github.com/obnahsgnaw/application"
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/url"
	"log"
)

func main() {
	app := application.New("demo", "Demo")

	defer app.Release()

	s := api.New(app, "auth", "auth", endtype.Backend, "/auth", url.Host{Ip: "127.0.0.1", Port: 8001}, 1)

	s.WithRpcServer(8002)

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

	log.Printf("Exited")
}
