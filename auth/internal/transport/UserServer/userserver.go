package userserver

import (
	"auth/internal/authservice"
	"auth/internal/models"
	"auth/internal/storagenode"
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
	accessTokenDuration  = 5 * time.Minute
	refreshTokenDuration = time.Hour * 24
)

type Server struct {
	port         string
	repo         UserRepo
	outerstorage storagenode.OuterStorage
}

type UserRepo interface {
	AddUser(username, password string)
	Authenticate(username, password string) bool
}

func New(port string, repo UserRepo, outerstorage storagenode.OuterStorage) *Server {
	return &Server{port: port, repo: repo, outerstorage: outerstorage}
}

func (s *Server) Run(ctx context.Context) error {
	e := echo.New()

	e.POST("/api/v1/login", s.login) // передаем json {email, password} получаем в Secure http-only cookie refresh
	//в json access + верификация пользователя

	e.POST("/api/v1/signup", s.signup) // передаем json {email, password} получаем в Secure http-only cookie refresh
	//в json access

	e.GET("/api/v1/getUserKeys", authservice.CheckJwt(s.getUserKeys, accessSecret)) //получить все ключ значения проверка валидности jwt

	e.POST("/api/v1/addKey", authservice.CheckJwt(s.addKey, accessSecret)) // добавить значение формат json {key, value}

	e.GET("/api/v1/auth/refreshTokens", s.refreshTokens) // refresh access and access token

	return e.Start(fmt.Sprintf(":%s", s.port))
}

func (s *Server) login(c *echo.Context) error {
	req := models.UserAuthRequest{}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ok := s.repo.Authenticate(req.Email, req.Password)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	access, refresh, err := generateTokens(&req)
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

func (s *Server) signup(c *echo.Context) error {
	req := models.UserAuthRequest{}
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	s.repo.AddUser(req.Email, req.Password)
	access, refresh, err := generateTokens(&req)

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
	claims, _ := c.Get("userClaims").(*models.UserClaims)
	data, err := s.outerstorage.GetUserData(claims.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No such user")
	}

	return c.JSON(http.StatusOK, models.UserKeyValue{UserKeyValue: data})
}

func (s *Server) addKey(c *echo.Context) error {
	claims, _ := c.Get("userClaims").(*models.UserClaims)

	var data models.KeyValue
	if c.Bind(&data) != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	_ = s.outerstorage.AddValue(claims.Email, data.Key, data.Value)
	return echo.NewHTTPError(http.StatusOK, "OK")
}

func (s *Server) refreshTokens(c *echo.Context) error {
	token, err := c.Cookie("refresh_token")

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	//TODO deactivate other sessions with this token

	claims, err := utils.VerifyToken(token.Value, refreshSecretKey)

	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	access, refresh, err := generateTokens(&models.UserAuthRequest{Email: claims.Email})
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

func generateTokens(user *models.UserAuthRequest) (string, string, error) {
	access, err := utils.GenerateToken(models.UserInfo{Email: user.Email}, accessSecret, accessTokenDuration)

	if err != nil {
		return "", "", echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	refresh, err := utils.GenerateToken(models.UserInfo{Email: user.Email}, refreshSecretKey, refreshTokenDuration)

	if err != nil {
		return "", "", echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return access, refresh, nil
}
