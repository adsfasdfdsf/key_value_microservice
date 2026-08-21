package authservice

import (
	"auth/internal/utils"
	"fmt"
	"github.com/labstack/echo/v5"
	"net/http"
	"strings"
)

func CheckJwt(next echo.HandlerFunc, secretKey []byte) echo.HandlerFunc {
	return func(c *echo.Context) error {
		fmt.Printf("hi")
		token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "no token provided")
		}
		claims, err := utils.VerifyToken(token, secretKey)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}
		c.Set("userClaims", claims)
		return next(c)
	}
}
