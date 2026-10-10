package userserver

import (
	"auth/pkg/logger"
	"context"
	"fmt"

	"github.com/labstack/echo/v5"
)

func LogInterceptor(ctx context.Context) echo.MiddlewareFunc {
	return func (next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			log := logger.GetLogger(ctx)
			log.Info(ctx, fmt.Sprintf("New Request: %s ", c.Request().RequestURI))
			return next(c)
		}
	}
}