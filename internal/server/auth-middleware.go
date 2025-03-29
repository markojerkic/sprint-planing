package server

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/server/auth"
	gormstore "github.com/wader/gormstore/v2"
)

const (
	// Session settings
	sessionName     = "user_session"
	sessionUserID   = "user_id"
	sessionDuration = 30 * 24 * 60 * 60 // 30 days in seconds
)

// InitSessions configures the session store for the application
func (s *Server) InitSessions(e *echo.Echo) {
	store := gormstore.New(s.db.DB, []byte(os.Getenv("SESSION_SECRET")))
	store.MaxLength(32 * 1024)
	store.SessionOpts = &sessions.Options{
		Path:     "/",
		MaxAge:   sessionDuration,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   os.Getenv("ENV") == "production",
	}

	go store.PeriodicCleanup(1*time.Hour, make(<-chan struct{}))

	// Register the session middleware
	e.Use(session.Middleware(store))
	e.Use(s.AuthMiddleware)

}

// AuthMiddleware checks for existing user session or creates a new user
func (s *Server) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := s.checkUser(c); err != nil {
			return err
		}

		s.checkJiraUser(c)

		return next(c)
	}
}

func (s *Server) checkJiraUser(c echo.Context) {
	session, err := session.Get(sessionName, c)
	if err != nil {
		return
	}

	acessToken, ok := session.Values[auth.JiraSessionAccessToken].(string)
	if !ok {
		slog.Warn("Access token not found in session")
		return
	}
	refreshToken, ok := session.Values[auth.JiraSessionRefreshToken].(string)
	if !ok {
		slog.Warn("Refresh token not found in session")
		return
	}
	resourceID, ok := session.Values[auth.JiraSessionResourceID].(string)
	if !ok {
		slog.Warn("Resource ID not found in session")
		return
	}
	expiry, ok := session.Values[auth.JiraSessionExpiry].(int64)
	if !ok {
		slog.Warn("Expiry not found in session")
		return
	}

	jiraClientInfo := auth.JiraClientInfo{
		AccessToken:  acessToken,
		RefreshToken: refreshToken,
		ResourceID:   resourceID,
		Expiry:       time.Unix(expiry, 0),
	}

	c.Set(auth.JiraClientInfoKey, &jiraClientInfo)
}

func (s *Server) checkUser(c echo.Context) error {
	session, err := session.Get(sessionName, c)
	if err != nil {
		return err
	}

	var user database.User
	userID := session.Values[sessionUserID]
	if userID == nil {
		// Create a new user
		s.db.DB.Create(&database.User{}).Scan(&user)
		session.Values[sessionUserID] = user.ID
		session.Save(c.Request(), c.Response())
	} else {
		// Fetch the user from the database
		if err := s.db.DB.First(&user, userID).Scan(&user).Error; err != nil {
			return err
		}
	}
	c.Set("user", user)

	return nil
}
