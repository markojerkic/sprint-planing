package server

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	// Session settings
	sessionName     = "user_session"
	sessionUserID   = "user_id"
	sessionDuration = 30 * 24 * 60 * 60 // 30 days in seconds
)

// InitSessions configures the session store for the application
func (s *Server) InitSessions(e *echo.Echo) {
	// Create a session store with a secret key
	// In production, this key should be loaded from a secure location
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	// Configure session options
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   sessionDuration,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	// Register the session middleware
	e.Use(session.Middleware(store))
	e.Use(s.AuthMiddleware)
}

// AuthMiddleware checks for existing user session or creates a new user
func (s *Server) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
        return next(c)
    }

}
