package userserver

import (
	"context"
	"github.com/labstack/echo/v5"
	"net/http"
)

type Server struct {
	port string
}

func New(port string) *Server {
	return &Server{port: port}
}

func (s *Server) Run(ctx context.Context) error {
	e := echo.New()

	e.POST("/login", login)

	e.POST("/signup", signup)

	e.GET("/api/v1/getUserKeys", getUserKeys)
	return nil
}

func login(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func signup(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func getUserKeys(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
