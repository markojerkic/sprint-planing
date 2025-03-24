package auth

import (
	"log/slog"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (o *OAuthRouter) JiraContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := session.Get(sessionName, c)
		if err != nil {
			return err
		}

		jiraClientInfo, ok := session.Values[JiraClientInfoKey].(JiraClientInfo)
		if !ok {
			slog.Warn("Jira client info not found in session")
			return next(c)
		}

		c.Set(JiraClientInfoKey, jiraClientInfo)
		return next(c)

	}
}
