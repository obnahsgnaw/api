package errhandler

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/obnahsgnaw/api/pkg/apierr"
	"github.com/obnahsgnaw/api/pkg/errobj"
	"github.com/obnahsgnaw/application/pkg/debug"
	"github.com/obnahsgnaw/application/pkg/dynamic"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
	"io"
	"net/http"
	"net/textproto"
	"strings"
)

var defErrObjProvider errobj.Provider
var defDebugger debug.Debugger

func SetDefaultErrObjProvider(provider errobj.Provider) {
	defErrObjProvider = provider
}

func SetDefaultDebugger(debugger2 debug.Debugger) {
	defDebugger = debugger2
}

func OutgoingHeaderMatcher(key string) (string, bool) {
	if key == "StatusCode" {
		return key, true
	}
	return fmt.Sprintf("%s%s", runtime.MetadataHeaderPrefix, key), true
}

func DefaultErrorHandler(err error, marshaler runtime.Marshaler, w http.ResponseWriter) {
	if defErrObjProvider == nil {
		panic("default error object provider not set")
	}
	if defDebugger == nil {
		defDebugger = debug.New(dynamic.NewBool(func() bool {
			return true
		}))
	}

	HandlerErr(err, marshaler, w, nil, defErrObjProvider, defDebugger)
}

func HandlerErr(err error, marshaler runtime.Marshaler, w http.ResponseWriter, ext func(), p errobj.Provider, debugger debug.Debugger) {
	// return Internal when Marshal failed
	const fallback = `{"code": -1, "message": "failed to marshal error message"}`
	var customStatus *runtime.HTTPStatusError
	if errors.As(err, &customStatus) {
		err = customStatus.Err
	}
	var apiErr *apierr.ApiError
	var s *status.Status
	if !errors.As(err, &apiErr) {
		apiErr = apierr.NewCommonInternalError(err)
	}
	apiErrs := ParseErrors(apiErr, debugger.Debug())

	s = status.New(codes.Code(apiErr.ErrCode.Code()), apiErr.Error())

	pb := s.Proto()
	contentType := marshaler.ContentType(pb)
	w.Header().Set("Content-Type", contentType)
	resp := CommonErrorResponse(pb)
	resp.Errors = apiErrs
	var resp1 interface{}
	if p != nil {
		resp1 = p(resp)
	} else {
		resp1 = resp
	}
	buf, merr := marshaler.Marshal(resp1)
	if apiErr.StatusCode == apierr.StatusCreated {
		buf, merr = marshaler.Marshal(apiErr.Data)
	}
	if merr != nil {
		grpclog.Infof("Failed to marshal error message %q: %v", s, merr)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			grpclog.Infof("Failed to write response: %v", err)
		}
		return
	}
	if ext != nil {
		ext()
	}
	st := runtime.HTTPStatusFromCode(s.Code())
	if customStatus != nil {
		st = customStatus.HTTPStatus
	}

	w.WriteHeader(st)
	if apiErr.StatusCode != apierr.StatusDeleted {
		if _, err := w.Write(buf); err != nil {
			grpclog.Infof("Failed to write response: %v", err)
		}
	}

	return
}

// HTTPErrorHandler is the default error handler.
// If "err" is a gRPC Status, the function replies with the status code mapped by HTTPStatusFromCode.
// If "err" is a HTTPStatusError, the function replies with the status code provide by that struct. This is
// intended to allow passing through of specific statuses via the function set via WithRoutingErrorHandler
// for the ServeMux constructor to handle edge cases which the standard mappings in HTTPStatusFromCode
// are insufficient for.
// If otherwise, it replies with http.StatusInternalServerError.
//
// The response body written by this function is a Status message marshaled by the Marshaler.
func HTTPErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error, p errobj.Provider, debugger debug.Debugger) {
	w.Header().Del("Trailer")
	w.Header().Del("Transfer-Encoding")
	var doForwardTrailers bool
	var md runtime.ServerMetadata
	HandlerErr(err, marshaler, w, func() {
		var ok bool
		md, ok = runtime.ServerMetadataFromContext(ctx)
		if !ok {
			grpclog.Infof("Failed to extract ServerMetadata from context")
		}

		handleForwardResponseServerMetadata(w, mux, md)

		// RFC 7230 https://tools.ietf.org/html/rfc7230#section-4.1.2
		// Unless the request includes a TE header field indicating "trailers"
		// is acceptable, as described in Section 4.3, a server SHOULD NOT
		// generate trailer fields that it believes are necessary for the user
		// agent to receive.
		doForwardTrailers = requestAcceptsTrailers(r)

		if doForwardTrailers {
			handleForwardResponseTrailerHeader(w, md)
			w.Header().Set("Transfer-Encoding", "chunked")
		}
	}, p, debugger)

	if doForwardTrailers {
		handleForwardResponseTrailer(w, md)
	}
}

func CommonErrorResponse(pb *spb.Status) errobj.Param {
	return errobj.Param{
		Code:    uint32(pb.GetCode()),
		Message: pb.GetMessage(),
	}
}

func ParseErrors(err error, debug bool) (errs []errobj.Param) {
	for {
		if err = errors.Unwrap(err); err == nil {
			break
		}
		var ae *apierr.ApiError
		if errors.As(err, &ae) {
			errs = append(errs, errobj.Param{Code: ae.ErrCode.Code(), Message: ae.Error()})
		} else {
			msg := apierr.InternalError.Message(nil, "")
			if debug {
				msg += ":" + err.Error()
			}
			errs = append(errs, errobj.Param{Code: apierr.InternalError.Code(), Message: msg})
		}
	}

	return
}

func handleForwardResponseServerMetadata(w http.ResponseWriter, _ *runtime.ServeMux, md runtime.ServerMetadata) {
	for k, vs := range md.HeaderMD {
		if h, ok := OutgoingHeaderMatcher(k); ok {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
	}
}

func requestAcceptsTrailers(req *http.Request) bool {
	te := req.Header.Get("TE")
	return strings.Contains(strings.ToLower(te), "trailers")
}

func handleForwardResponseTrailerHeader(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k := range md.TrailerMD {
		tKey := textproto.CanonicalMIMEHeaderKey(fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k))
		w.Header().Add("Trailer", tKey)
	}
}

func handleForwardResponseTrailer(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k, vs := range md.TrailerMD {
		tKey := fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k)
		for _, v := range vs {
			w.Header().Add(tKey, v)
		}
	}
}
