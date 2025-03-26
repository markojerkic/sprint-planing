package server

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database"
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
	store.MaxLength(32 * 1024) // 32KB should be plenty
	store.SessionOpts = &sessions.Options{
		Path:     "/",
		MaxAge:   sessionDuration,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   os.Getenv("ENV") == "production",
	}

	go store.PeriodicCleanup(1*time.Hour, make(<-chan struct{}))
	// Configure session options

	// Register the session middleware
	e.Use(session.Middleware(store))
	e.Use(s.AuthMiddleware)

}

// AuthMiddleware checks for existing user session or creates a new user
func (s *Server) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		return next(c)
	}

}
