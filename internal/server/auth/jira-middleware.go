package auth

import (
	"log/slog"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (o *OAuthRouter) JiraContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := session.Get(sessionName, c)
		if err != nil {
			return err
		}

		acessToken, ok := session.Values[sessionAccessToken].(string)
		if !ok {
			slog.Warn("Access token not found in session")
			return next(c)
		}
		refreshToken, ok := session.Values[sessionRefreshToken].(string)
		if !ok {
			slog.Warn("Refresh token not found in session")
			return next(c)
		}
		resourceID, ok := session.Values[sessionResourceID].(string)
		if !ok {
			slog.Warn("Resource ID not found in session")
			return next(c)
		}
		expiry, ok := session.Values[sessionExpiry].(int64)
		if !ok {
			slog.Warn("Expiry not found in session")
			return next(c)
		}

		jiraClientInfo := JiraClientInfo{
			AccessToken:  acessToken,
			RefreshToken: refreshToken,
			ResourceID:   resourceID,
			Expiry:       time.Unix(expiry, 0),
		}

		c.Set(JiraClientInfoKey, &jiraClientInfo)
		return next(c)

	}
}
