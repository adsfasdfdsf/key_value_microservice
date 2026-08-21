package userserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
)

type Server struct {
	port string
}

func New(port string) *Server {
	return &Server{port: port}
}

func (s *Server) Run(ctx context.Context) error {
	e := echo.New()

	e.POST("api/v1/login", login)

	e.POST("api/v1/signup", signup)

	e.GET("/api/v1/getUserKeys", getUserKeys)

	e.POST("/api/v1/refreshAccessToken", refreshAccessToken)

	e.GET("/api/v1/getRefreshToken", getRefreshToken)

	return e.Start(fmt.Sprintf(":%s", s.port))
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

func refreshAccessToken(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func getRefreshToken(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}