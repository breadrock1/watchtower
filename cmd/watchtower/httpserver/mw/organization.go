package mw

import (
	"github.com/gofiber/fiber/v2"
)

const (
	DefaultOrganizationIDHeader = "default"
	OrganizationIDKey           = "organization_id"
	organizationIDHeader        = "X-Organization-Id"
)

func OrganizationContext() fiber.Handler {
	return func(eCtx *fiber.Ctx) error {
		authHeaderValue := eCtx.Get(organizationIDHeader)

		if authHeaderValue == "" {
			authHeaderValue = DefaultOrganizationIDHeader
		}

		eCtx.Locals(OrganizationIDKey, authHeaderValue)

		return eCtx.Next()
	}
}
