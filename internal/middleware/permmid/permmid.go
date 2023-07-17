package permmid

import (
	"github.com/obnahsgnaw/api/internal/errhandler"
	"github.com/obnahsgnaw/api/internal/marshaler"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/service"
	"github.com/obnahsgnaw/api/service/perm"
	"net/http"
	"strings"
)

// NewMuxPermissionMid 需要拿pattern来确定权限，所以没放到gin中间件
func NewMuxPermissionMid(manager *perm.Manager) service.MuxRouteHandleFunc {
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string, pattern string) bool {
		if manager.Logger() != nil {
			manager.Logger().Debug("Middleware [perm ]: enabled")
		}
		appId := r.Header.Get("X-App-Id")
		userId := r.Header.Get("X-User-Id")
		if manager.Logger() != nil {
			manager.Logger().Debug("Middleware [perm ]: enabled")
		}
		method := strings.ToLower(r.Method)
		var err error
		// 验证权限
		if err = manager.Provider().Can(appId, userId, method, pattern); err != nil {
			if manager.Logger() != nil {
				manager.Logger().Debug("Middleware [perm ]: no perm, err=" + err.Error())
			}
			errhandler.HandlerErr(
				apierr.ToStatusError(apierr.NewForbiddenError(apierr.PermMidNoPerm, err)),
				marshaler.GetMarshaler(r.Header.Get("Accept")),
				w,
				nil,
				manager.ErrObjProvider(),
				manager,
			)
			return false
		} else {
			if manager.Logger() != nil {
				manager.Logger().Debug("Middleware [perm ]: accessed")
			}
		}

		return true
	}
}
