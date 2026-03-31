package httpserver

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) CreateSystemGroup(group fiber.Router) {
	group.Get("/", s.Home)
}

func (s *Server) Home(eCtx *fiber.Ctx) error {
	fileData, err := os.ReadFile("./static/index.html")
	if err != nil {
		return eCtx.SendStatus(fiber.StatusInternalServerError)
	}

	eCtx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return eCtx.SendString(string(fileData))
}
