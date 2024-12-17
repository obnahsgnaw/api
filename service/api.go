package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"net/http"
)

// MdValParser the income meta data value parser
type MdValParser func(ctx context.Context, r *http.Request) string

// MdProviders meta data providers
type MdProviders map[string]MdValParser

// RouteProvider route provider
type RouteProvider func(engine *gin.Engine)

type MuxRouteHandleFunc func(w http.ResponseWriter, r *http.Request, pathParams map[string]string, pattern string) bool

type MethodMdProvider struct {
	defProvider    MdProviders
	methodProvider map[string]MdProviders
	all            bool
}

func NewMdProvider() *MethodMdProvider {
	return &MethodMdProvider{
		defProvider:    make(MdProviders),
		methodProvider: make(map[string]MdProviders),
	}
}

func (p *MethodMdProvider) AddDefault(key string, parser MdValParser) {
	if key != "" && parser != nil {
		p.defProvider[key] = parser
	}
}

// Add method like:/backendapi.index.v1.OptionsService/OptionList
func (p *MethodMdProvider) Add(method, key string, parser MdValParser) {
	if method == "" || key == "" || parser == nil {
		return
	}
	if p.methodAll(method) {
		return
	}
	if _, ok := p.methodProvider[method]; !ok {
		p.methodProvider[method] = make(MdProviders)
	}
	p.methodProvider[method][key] = parser
}

func (p *MethodMdProvider) Range(ctx context.Context, q *http.Request, handler func(key, val string)) {
	if len(p.defProvider) > 0 {
		for k, pp := range p.defProvider {
			handler(k, pp(ctx, q))
		}
	}
	method, ok := runtime.RPCMethod(ctx)
	if !ok {
		return
	}
	if pps, ok1 := p.methodProvider[method]; ok1 {
		for k, pp := range pps {
			handler(k, pp(ctx, q))
		}
	}
}

func (p *MethodMdProvider) AddAll() {
	p.all = true
}

func (p *MethodMdProvider) AddMethodAll(method string) {
	if method == "" {
		return
	}
	p.methodProvider[method] = make(MdProviders)
	p.methodProvider[method]["all"] = nil
}

func (p *MethodMdProvider) All() bool {
	return p.all
}

func (p *MethodMdProvider) MethodAll(ctx context.Context) bool {
	method, ok := runtime.RPCMethod(ctx)
	if !ok {
		return false
	}
	return p.methodAll(method)
}

func (p *MethodMdProvider) methodAll(method string) bool {
	if v, ok := p.methodProvider[method]; ok && len(v) == 1 {
		_, ok = v["all"]
		return ok
	}
	return false
}
