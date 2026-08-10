package mw

import (
	"github.com/gofiber/fiber/v2"
)

const (
	DefaultUserIDHeader = "unknown"
	UserIDHeaderKey     = "user_id"
	userIDHeader        = "X-User-Id"
)

func UserContext() fiber.Handler {
	return func(eCtx *fiber.Ctx) error {
		authHeaderValue := eCtx.Get(userIDHeader)

		if authHeaderValue == "" {
			authHeaderValue = DefaultUserIDHeader
		}

		eCtx.Locals(UserIDHeaderKey, authHeaderValue)

		return eCtx.Next()
	}
}
