package mw

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

var (
	excludedPaths = []string{
		"/health",
		"/metrics",
		"/favicon.ico",
		"/static/",
		"/api/v1/swagger",
	}
)

func TracerSkipper(eCtx echo.Context) bool {
	for _, excluded := range excludedPaths {
		if strings.HasPrefix(eCtx.Path(), excluded) {
			return true
		}
	}

	return eCtx.Request().Method == http.MethodOptions
}
