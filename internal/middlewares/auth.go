package middlewares

import (
	"net/http"
	"strings"
	"weblog/internal/auth"

	"github.com/labstack/echo/v5"
)

func RequireAuth(jwtKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "no auth here")
			}
			slice := strings.Split(header, " ")
			if slice[0] != "Bearer" || len(slice) != 2 {
				return echo.NewHTTPError(http.StatusUnauthorized, "not auth correct form")

			}

			token := slice[1]
			claims, err := auth.ParseToken(token, jwtKey)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "expired token or something is wrong with this token")
			}
			c.Set("user_id", claims.UserID)
			return next(c)

		}
	}
}
