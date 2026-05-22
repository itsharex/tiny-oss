package middleware

import (
	"myoss/config"

	"github.com/labstack/echo/v5"
)

func ApiAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		token := c.Request().Header.Get("X-API-Token")
		if token == "" {
			token = c.QueryParam("api_token")
		}

		if token == "" || token != config.Config.Security.APIToken {
			return c.JSON(403, map[string]string{"msg": "无访问权限"})
		}
		return next(c)
	}
}
