package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

// MdValParser the income meta data value parser
type MdValParser func(ctx context.Context, r *http.Request) string

// MdProviders meta data providers
type MdProviders map[string]MdValParser

// RouteProvider route provider
type RouteProvider func(engine *gin.Engine)

type MuxRouteHandleFunc func(w http.ResponseWriter, r *http.Request, pathParams map[string]string, pattern string) bool
