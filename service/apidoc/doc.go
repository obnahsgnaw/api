package apidoc

import (
	"github.com/obnahsgnaw/application/endtype"
	"github.com/obnahsgnaw/application/pkg/url"
	"github.com/obnahsgnaw/application/service/regCenter"
)

type Config struct {
	Protocol     url.Protocol
	origin       url.Origin
	Path         string
	Title        string
	Provider     func() ([]byte, error)
	DebugOrigin  url.Origin
	EndType      endtype.EndType
	RegTtl       int64
	RegKeyPreGen regCenter.RegKeyPrefixGenerator
}

func (c *Config) SetOrigin(origin url.Origin) {
	c.origin = origin
}
func (c *Config) Url() string {
	return c.origin.String() + c.Path
}
