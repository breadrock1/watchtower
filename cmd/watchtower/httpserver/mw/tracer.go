package mw

import (
	"strings"

	"github.com/gofiber/fiber/v2"
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

func TracerSkipper(eCtx *fiber.Ctx) bool {
	for _, excluded := range excludedPaths {
		if strings.HasPrefix(eCtx.Path(), excluded) {
			return true
		}
	}

	return eCtx.Request().Header.IsOptions()
}
