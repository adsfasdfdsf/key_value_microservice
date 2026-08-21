package userserver

import (
	"auth/internal/authservice"
	"auth/internal/models"
	"auth/internal/utils"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

var (
	refreshSecretKey     = []byte("Secret!!")
	accessSecret         = []byte("Access!!")
	accessTokenDuration  = time.Minute
	refreshTokenDuration = time.Hour * 24
)

type Server struct {
	port string
	repo UserRepo
}

type UserRepo interface {
	AddUser(username, password string)
	Authenticate(username, password string) bool
}

func New(port string, repo UserRepo) *Server {
	return &Server{port: port, repo: repo}
}

func (s *Server) Run(ctx context.Context) error {
	e := echo.New()

	e.POST("/api/v1/login", s.login) // передаем json {email, password} получаем в Secure http-only cookie refresh
	//в json access + верификация пользователя

	e.POST("/api/v1/signup", s.signup) // передаем json {email, password} получаем в Secure http-only cookie refresh
	//в json access

	e.GET("/api/v1/getUserKeys", authservice.CheckJwt(s.getUserKeys, accessSecret)) //получить все ключ значения проверка валидности jwt
	//TODO в check jwt выбрасывать не internal server error а unauthorized и прекидывать на login

	e.GET("/api/v1/auth/refreshTokens", s.refreshTokens) // refresh access and access token

	return e.Start(fmt.Sprintf(":%s", s.port))
}

func (s *Server) login(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func (s *Server) signup(c *echo.Context) error {
	req := models.UserAuthRequest{}
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	//TODO выделить в отдельный метод
	s.repo.AddUser(req.Email, req.Password)
	access, err := utils.GenerateToken(models.UserInfo{Username: req.Email}, accessSecret, accessTokenDuration)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	refresh, err := utils.GenerateToken(models.UserInfo{Username: req.Email}, refreshSecretKey, refreshTokenDuration)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	cookie := new(http.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = refresh
	cookie.Expires = time.Now().Add(refreshTokenDuration)
	cookie.HttpOnly = true
	cookie.Secure = true

	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, access)
}

func (s *Server) getUserKeys(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func (s *Server) refreshTokens(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
