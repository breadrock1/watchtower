package mw

import (
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

func TracerSkipper(c echo.Context) bool {
	for _, excluded := range excludedPaths {
		if strings.HasPrefix(c.Path(), excluded) {
			return true
		}
	}

	if c.Request().Method == "OPTIONS" {
		return true
	}

	return false
}
