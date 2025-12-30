package httpserver

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func (s *Server) CreateSystemGroup() error {
	s.server.GET("/", s.Home)
	return nil
}

func (s *Server) Home(eCtx echo.Context) error {
	fileData, err := os.ReadFile("./static/index.html")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return eCtx.HTMLBlob(http.StatusOK, fileData)
}
