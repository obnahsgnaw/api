package authroute

import (
	"errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"strings"
)

// Manager need auth route manager
type Manager struct {
	route map[string]struct{}
}

// New return new manager
func New() *Manager {
	return &Manager{route: make(map[string]struct{})}
}

// AddAuthIgnoredRoute add ignore auth route
func (m *Manager) AddAuthIgnoredRoute(method, uriPattern string) {
	m.route[strings.ToLower(method)+uriPattern] = struct{}{}
}

// AuthMust check the route auth must or not
func (m *Manager) AuthMust(method, uriPattern string) bool {
	_, ok := m.route[strings.ToLower(method)+uriPattern]
	return !ok
}

// ScanNoAuth scan ignore auth proto route
func ScanNoAuth(pathPrefix []string, handler func(prefix, method, urlPattern string)) error {
	var err error
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		var prefix string
		for _, pre := range pathPrefix {
			if strings.HasPrefix(fd.Path(), pre) {
				break
			}
		}
		if prefix == "" {
			return true
		}
		// loop service
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			service := services.Get(i)
			methods := service.Methods()
			// loop method
			for j := 0; j < methods.Len(); j++ {
				method := methods.Get(j)
				noAuth := true
				methodName := "GET"
				uri := ""
				if ope, _ := proto.GetExtension(method.Options(), options.E_Openapiv2Operation).(*options.Operation); ope != nil {
					securities := ope.Security
					for _, security := range securities {
						if _, secExist := security.SecurityRequirement["BearerToken"]; secExist {
							noAuth = false
							break
						}
					}
				} else {
					continue
				}
				if api, _ := proto.GetExtension(method.Options(), annotations.E_Http).(*annotations.HttpRule); api != nil {
					if uri = api.GetGet(); uri != "" {
						methodName = "GET"
					} else if uri = api.GetPost(); uri != "" {
						methodName = "POST"
					} else if uri = api.GetPut(); uri != "" {
						methodName = "PUT"
					} else if uri = api.GetPatch(); uri != "" {
						methodName = "PATCH"
					} else if uri = api.GetDelete(); uri != "" {
						methodName = "DELETE"
					} else {
						err = errors.New("resource scan: Unknown method")
						return false
					}
				}
				if uri == "" || methodName == "" {
					err = errors.New("resource scan: no method uri or method type[" + string(method.FullName()) + "]")
					return false
				}
				if noAuth {
					handler(prefix, methodName, uri)
				}
			}
		}
		return true
	})
	return err
}
