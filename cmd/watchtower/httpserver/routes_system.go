package httpserver

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) CreateSystemGroup(group fiber.Router) {
	group.Get("/", s.Home)
	group.Get("/health", s.Health)
}

func (s *Server) Home(eCtx *fiber.Ctx) error {
	fileData, err := os.ReadFile("./static/index.html")
	if err != nil {
		return eCtx.SendStatus(fiber.StatusInternalServerError)
	}

	eCtx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return eCtx.SendString(string(fileData))
}

func (s *Server) Health(eCtx *fiber.Ctx) error {
	ctx := eCtx.UserContext()

	if err := s.state.Health(ctx); err != nil {
		return eCtx.Status(fiber.StatusServiceUnavailable).
			SendString(fmt.Sprintf("service unavailable: %s", err.Error()))
	}

	return eCtx.SendStatus(http.StatusOK)
}
