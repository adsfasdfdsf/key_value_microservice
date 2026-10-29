package userserver

import (
	"auth/internal/authservice"
	"auth/internal/models"
	nodestorage "auth/internal/storage/node_storage"
	"auth/internal/storage/storagenode"
	"auth/internal/utils"
	"auth/pkg/logger"
	"auth/pkg/storage/tokenrepo"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"
)

var (
	refreshSecretKey     = []byte(os.Getenv("REFRESH_TOKEN_SECRET"))
	accessSecret         = []byte(os.Getenv("ACCESS_TOKEN_SECRET"))
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
	ctx, cancel := context.WithCancel(ctx)
	return &Server{ctx: ctx, port: port, repo: repo, outerstorage: outerstorage, tokens: tokens, shutdown: cancel}
}

func (s *Server) Run() error {
	if len(accessSecret) < 32 || len(refreshSecretKey) < 32 || string(accessSecret) == string(refreshSecretKey) {
		return fmt.Errorf("distinct ACCESS_TOKEN_SECRET and REFRESH_TOKEN_SECRET of at least 32 bytes required")
	}
	e := echo.New()
	origins := []string{"http://localhost:5173"}
	if configured := strings.TrimSpace(os.Getenv("CORS_ORIGINS")); configured != "" {
		origins = strings.Split(configured, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
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

	e.PUT("/api/v1/addKey", authservice.CheckJwt(s.addKey, accessSecret))
	e.POST("/api/v1/addKey", authservice.CheckJwt(s.addKey, accessSecret)) // добавить значение формат json {key, value}

	e.GET("/api/v1/auth/refreshTokens", s.refreshTokens) // refresh refresh and access token
	logger.GetLogger(s.ctx).Info(s.ctx, "starting server")
	ctx := s.ctx
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
	req.Email = strings.TrimSpace(req.Email)
	if address, err := mail.ParseAddress(req.Email); err != nil || address.Address != req.Email || len(req.Password) < 8 || len(req.Password) > 72 {
		return echo.NewHTTPError(http.StatusBadRequest, "valid email and password of 8 to 72 bytes required")
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
	data, err := s.outerstorage.GetUserData(claims.Email)
	if err != nil {
		log.Error(s.ctx, "Could not get user values", zap.Error(err))
		return storageError(err)
	}
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

	if err := s.outerstorage.AddValue(claims.Email, data.Key, data.Value); err != nil {
		log.Error(s.ctx, "Could not add user value", zap.Error(err))
		return storageError(err)
	}
	log.Info(s.ctx, "User added value", zap.String("user email", claims.Email))
	return c.JSON(http.StatusOK, map[string]bool{"success": true})
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
	cookie.Secure = os.Getenv("COOKIE_SECURE") == "true"
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

func storageError(err error) error {
	var upstream *nodestorage.HTTPError
	if errors.As(err, &upstream) {
		switch upstream.Status {
		case http.StatusBadRequest:
			return echo.NewHTTPError(http.StatusBadRequest, "invalid key or value")
		case http.StatusServiceUnavailable:
			return echo.NewHTTPError(http.StatusServiceUnavailable, "storage electing leader or recovering; retry later")
		case http.StatusGatewayTimeout:
			return echo.NewHTTPError(http.StatusGatewayTimeout, "storage request timed out")
		}
	}
	return echo.NewHTTPError(http.StatusBadGateway, "storage unavailable")
}
