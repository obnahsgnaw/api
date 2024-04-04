package permmid

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/api/service/perm"
	"net/http"
	"strings"
)

// NewMuxPermissionMid 需要拿pattern来确定权限，所以没放到gin中间件
func NewMuxPermissionMid(manager *perm.Manager, debugCb func(msg string), errHandle func(err error, marshaler runtime.Marshaler, w http.ResponseWriter)) service.MuxRouteHandleFunc {
	if debugCb == nil {
		debugCb = func(msg string) {}
	}
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string, pattern string) bool {
		appId := r.Header.Get(manager.AppIdHeaderKey())
		userId := r.Header.Get(manager.UserIdHeaderKey())
		method := strings.ToLower(r.Method)
		var err error
		// 验证权限
		if !manager.Ignored(method, pattern) {
			if err = manager.Provider().Can(appId, userId, method, pattern); err != nil {
				debugCb("perm-middleware: no perm, desc=" + err.Error())
				errHandle(
					apierr.ToStatusError(apierr.NewForbiddenError(apierr.PermMidNoPerm, err)),
					marshaler.GetMarshaler(r.Header.Get("Accept")),
					w,
				)
				return false
			} else {
				debugCb("perm-middleware: accessed")
			}
		} else {
			debugCb("perm-middleware: validate ignored by ignorer")
		}

		return true
	}
}
