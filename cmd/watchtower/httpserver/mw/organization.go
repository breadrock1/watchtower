package mw

import (
	"github.com/gofiber/fiber/v2"
)

const (
	DefaultOrganizationIDHeader = "default"
	OrganizationIDHeader        = "X-Organization-Id"
)

func OrganizationContext() fiber.Handler {
	return func(eCtx *fiber.Ctx) error {
		authHeaderValue := eCtx.Get(OrganizationIDHeader)

		if authHeaderValue == "" {
			authHeaderValue = DefaultOrganizationIDHeader
		}

		eCtx.Locals(OrganizationIDHeader, authHeaderValue)

		return eCtx.Next()
	}
}
