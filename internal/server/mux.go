package server

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/errhandler"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/application/pkg/debug"
	"google.golang.org/grpc/metadata"
	"net/http"
	"strings"
)

func NewMux() *runtime.ServeMux {
	return runtime.NewServeMux()
}
func InitMux(mux *runtime.ServeMux, mdProviders *service.MethodMdProvider, middlewares []service.MuxRouteHandleFunc, p errobj.Provider, debugger debug.Debugger) {
	ops := []runtime.ServeMuxOption{
		runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
			return "", false
		}),
		runtime.WithOutgoingHeaderMatcher(errhandler.OutgoingHeaderMatcher),
		// trans header to metadata
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			var metaData []string
			if mdProviders.All() || mdProviders.MethodAll(ctx) {
				for k, v := range request.Header {
					metaData = append(metaData, k, strings.Join(v, " "))
				}
			} else {
				metaData = append(metaData, "rq_id", request.Header.Get("X-Request-Id"))
				metaData = append(metaData, "rq_type", request.Header.Get("X-Request-Type"))
				metaData = append(metaData, "rq_from", request.Header.Get("X-Request-From"))
				mdProviders.Range(ctx, request, func(key, val string) {
					metaData = append(metaData, key, val)
				})
			}
			md := metadata.Pairs(metaData...)
			return md
		}),
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, m runtime.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
			errhandler.HTTPErrorHandler(ctx, mux, m, writer, request, err, p, debugger)
		}),
		runtime.WithMarshalerOption("*", marshaler.JsonMarshaler()),
		runtime.WithMarshalerOption("application/json", marshaler.JsonMarshaler()),
		runtime.WithMarshalerOption("application/octet-stream", marshaler.ProtoMarshaler()),
		runtime.WithMarshalerOption("application/x-protobuf", marshaler.ProtoMarshaler()),
	}
	if len(middlewares) > 0 {
		for _, m := range middlewares {
			ops = append(ops, runtime.WithBeforeRoute(m))
		}
	}
	for _, o := range ops {
		o(mux)
	}
}
