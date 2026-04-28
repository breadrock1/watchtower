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
			//err := fmt.Errorf("missing x-organization-id header key")
			//span.SetStatus(codes.Error, err.Error())
			//span.RecordError(err)
			//return eCtx.Status(fiber.StatusUnauthorized).SendString(err.Error())

			authHeaderValue = DefaultOrganizationIDHeader
		}

		eCtx.Locals(OrganizationIDHeader, authHeaderValue)

		return eCtx.Next()
	}
}
