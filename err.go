package api

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/errhandler"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/application/pkg/debug"
	"net/http"
)

func SetDefaultErrObjProvider(provider errobj.Provider) {
	errhandler.SetDefaultErrObjProvider(provider)
}

func SetDefaultDebugger(debugger2 debug.Debugger) {
	errhandler.SetDefaultDebugger(debugger2)
}

func HandleError(err error, marshaler runtime.Marshaler, w http.ResponseWriter) {
	errhandler.DefaultErrorHandler(err, marshaler, w)
}
