package userserver

import (
	"auth/internal/authservice"
	"auth/internal/models"
	"auth/internal/storagenode"
	"auth/internal/utils"
	"auth/pkg/logger"
	"auth/pkg/storage/tokenrepo"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"
)

var (
	refreshSecretKey     = []byte("Secret!!") //TODO брать из env или volt
	accessSecret         = []byte("Access!!") //TODO брать из env или volt
	accessTokenDuration  = 5 * time.Minute
	refreshTokenDuration = time.Hour * 24
)

type Server struct {
	ctx          context.Context
	port         string
	repo         UserRepo
	outerstorage storagenode.OuterStorage
	shutdown     context.CancelFunc
	tokens       TokenRepo
}

type UserRepo interface {
	AddUser(email, password string) (*models.User, string, error)
	Authenticate(username, password string) bool
}

type TokenRepo interface {
	Save(context.Context, string, string, string, time.Time, time.Time) error
	Rotate(context.Context, string, string, string, string, time.Time, time.Time) error
}

func New(ctx context.Context, port string, repo UserRepo, outerstorage storagenode.OuterStorage, tokens TokenRepo) *Server {
	return &Server{ctx: ctx, port: port, repo: repo, outerstorage: outerstorage, tokens: tokens}
}

func (s *Server) Run() error {
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

	e.GET("/api/v1/auth/refreshTokens", s.refreshTokens) // refresh refresh and access token
	logger.GetLogger(s.ctx).Info(s.ctx, "starting server")
	ctx, stop := context.WithCancel(context.Background())
	s.shutdown = stop
	sc := echo.StartConfig{
		Address:         fmt.Sprintf(":%s", s.port),
		GracefulTimeout: 5 * time.Second,
	}

	return sc.Start(ctx, e)
} //fmt.Sprintf(":%s", s.port)

func (s *Server) login(c *echo.Context) error {
	log := logger.GetLogger(s.ctx)
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
	if err := s.saveTokens(c.Request().Context(), req.Email, access, refresh); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not store tokens")
	}
	log.Info(s.ctx, "User login", zap.String("user email", req.Email))
	setRefreshCookie(c, refresh)
	return c.JSON(http.StatusOK, models.UserAuthResponse{AccessToken: access})
}

func (s *Server) signup(c *echo.Context) error {
	log := logger.GetLogger(s.ctx)
	req := models.UserAuthRequest{}
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if _, _, err := s.repo.AddUser(req.Email, req.Password); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return echo.NewHTTPError(http.StatusConflict, "user already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "could not create user")
	}
	access, refresh, err := generateTokens(&req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if err := s.saveTokens(c.Request().Context(), req.Email, access, refresh); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not store tokens")
	}
	log.Info(s.ctx, "User sign up", zap.String("user email", req.Email))
	setRefreshCookie(c, refresh)
	return c.JSON(http.StatusOK, models.UserAuthResponse{AccessToken: access})
}

func (s *Server) getUserKeys(c *echo.Context) error {
	log := logger.GetLogger(s.ctx)
	claims, _ := c.Get("userClaims").(*models.UserClaims)
	data, _ := s.outerstorage.GetUserData(claims.Email)
	//TODO error handler
	if data == nil {
		data = []models.KeyValue{}
	}
	log.Info(s.ctx, "got keys", zap.String("user email", claims.Email))
	return c.JSON(http.StatusOK, models.UserKeyValue{UserKeyValue: data})
}

func (s *Server) addKey(c *echo.Context) error {
	log := logger.GetLogger(s.ctx)

	claims, _ := c.Get("userClaims").(*models.UserClaims)

	var data models.KeyValue
	if c.Bind(&data) != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	log.Info(s.ctx, "User added value", zap.String("user email", claims.Email))
	_ = s.outerstorage.AddValue(claims.Email, data.Key, data.Value)
	return echo.NewHTTPError(http.StatusOK, "OK")
}

func (s *Server) refreshTokens(c *echo.Context) error {
	log := logger.GetLogger(s.ctx)

	token, err := c.Cookie("refresh_token")
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	claims, err := utils.VerifyToken(token.Value, refreshSecretKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	access, refresh, err := generateTokens(&models.UserAuthRequest{Email: claims.Email})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	accessClaims, err := utils.VerifyToken(access, accessSecret)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not verify generated token")
	}
	refreshClaims, err := utils.VerifyToken(refresh, refreshSecretKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not verify generated token")
	}
	if err := s.tokens.Rotate(c.Request().Context(), claims.Email, token.Value, access, refresh,
		accessClaims.ExpiresAt.Time, refreshClaims.ExpiresAt.Time); err != nil {
		if errors.Is(err, tokenrepo.ErrInvalidRefresh) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "could not store tokens")
	}
	log.Info(s.ctx, "User refreshed tokens", zap.String("user email", claims.Email))

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

func (s *Server) Stop() error {
	s.shutdown()
	return nil
}
func (s *Server) saveTokens(ctx context.Context, email, access, refresh string) error {
	a, err := utils.VerifyToken(access, accessSecret)
	if err != nil {
		return err
	}
	r, err := utils.VerifyToken(refresh, refreshSecretKey)
	if err != nil {
		return err
	}
	return s.tokens.Save(ctx, email, access, refresh, a.ExpiresAt.Time, r.ExpiresAt.Time)
}
