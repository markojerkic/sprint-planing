package auth

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/cmd/web/components"
)

func JiraAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, ok := c.Get(JiraClientInfoKey).(*JiraClientInfo); !ok {
			slog.Warn("User not logged in via Jira, showing login page")
			return components.LoginToJira().Render(c.Request().Context(), c.Response().Writer)
		}

		return next(c)
	}
}
