package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func JiraAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, ok := c.Get(JiraClientInfoKey).(*JiraClientInfo); !ok {
			return c.JSON(http.StatusForbidden, "Jira client info not found")
		}

		return next(c)
	}
}
