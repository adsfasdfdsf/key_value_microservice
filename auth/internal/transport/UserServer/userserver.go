package userserver

import (
	"auth/internal/authservice"
	"auth/internal/models"
	"auth/internal/storagenode"
	"auth/internal/utils"
	"auth/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

var (
	refreshSecretKey     = []byte("Secret!!")
	accessSecret         = []byte("Access!!")
	accessTokenDuration  = 5 * time.Minute
	refreshTokenDuration = time.Hour * 24
)

type Server struct {
	ctx          context.Context
	port         string
	repo         UserRepo
	outerstorage storagenode.OuterStorage
}

type UserRepo interface {
	AddUser(username, password string)
	Authenticate(username, password string) bool
}

func New(ctx context.Context, port string, repo UserRepo, outerstorage storagenode.OuterStorage) *Server {
	return &Server{ctx: ctx, port: port, repo: repo, outerstorage: outerstorage}
}

func (s *Server) Run(ctx context.Context) error {
	e := echo.New()
	
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"http://localhost:5173"},
        AllowMethods: []string{
            http.MethodGet,
            http.MethodHead,
            http.MethodPut,
            http.MethodPatch,
            http.MethodPost,
            http.MethodDelete,
            http.MethodOptions,
        },
        AllowHeaders: []string{
            echo.HeaderOrigin,
            echo.HeaderContentType,
            echo.HeaderAccept,
            echo.HeaderAuthorization,
        },
		AllowCredentials: true, //TODO убрать в проде
    }))

	e.Use(LogInterceptor(s.ctx))

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
	//log := logger.GetLogger(s.ctx)
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

	setRefreshCookie(c, refresh)
	return c.JSON(http.StatusOK, models.UserAuthResponse{AccessToken: access})
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

	setRefreshCookie(c, refresh)
	return c.JSON(http.StatusOK, models.UserAuthResponse{AccessToken: access})
}

func (s *Server) getUserKeys(c *echo.Context) error {
	claims, _ := c.Get("userClaims").(*models.UserClaims)
	data, _ := s.outerstorage.GetUserData(claims.Email)
	//TODO error handler
	if data == nil {
		data = []models.KeyValue{}
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
	setRefreshCookie(c, refresh)
	return c.JSON(http.StatusOK, models.UserAuthResponse{AccessToken: access})
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


func setRefreshCookie(c *echo.Context, refresh string) {
    cookie := new(http.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = refresh
	cookie.Expires = time.Now().Add(refreshTokenDuration)
	cookie.HttpOnly = true
	cookie.Secure = false //TODO убрать в проде
	cookie.Path = "/"
    cookie.SameSite = http.SameSiteLaxMode //TODO убрать в проде

    c.SetCookie(cookie)
}