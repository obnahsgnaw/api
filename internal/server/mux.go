package server

import (
	"context"
	"github.com/obnahsgnaw/api/internal/errhandler"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/application/pkg/debug"
	"github.com/obnahsgnaw/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"net/http"
)

func getRpcApiProxyMux(mdProviders service.MdProviders, middlewares []service.MuxRouteHandleFunc, p errobj.Provider, debugger debug.Debugger) *runtime.ServeMux {
	ops := []runtime.ServeMuxOption{
		runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
			return "", false
		}),
		runtime.WithOutgoingHeaderMatcher(errhandler.OutgoingHeaderMatcher),
		// trans header to metadata
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			var metaData []string
			for k, valProvider := range mdProviders {
				metaData = append(metaData, k)
				metaData = append(metaData, valProvider(ctx, request))
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
	}
	if len(middlewares) > 0 {
		for _, m := range middlewares {
			ops = append(ops, runtime.WithBeforeRoute(m))
		}
	}
	return runtime.NewServeMux(
		ops...,
	)
}
