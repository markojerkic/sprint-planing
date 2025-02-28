package server

import (
	"log"
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
		ctx := c.Request().Context()
		sess, err := session.Get(sessionName, c)
		if err != nil {
			// Clear the potentially corrupted session
			sess, _ = session.Get(sessionName, c)
			sess.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   -1, // Delete the cookie
				HttpOnly: true,
			}
			sess.Save(c.Request(), c.Response())

			// Create a new session
			sess, _ = session.Get(sessionName, c)
		}

		// Check if there's a user ID in the session
		userIDInterface := sess.Values[sessionUserID]
		if userIDInterface != nil {
			// Try to get the user from the database
			if userID, ok := userIDInterface.(int64); ok {
				user, err := s.db.Queries.GetUser(ctx, userID)
				if err == nil {
					// User found, set in context and proceed
					c.Set("user", user)
					log.Printf("User found: %v", user)
					return next(c)
				}
				// Invalid user ID, will create new user below
			}
		}
		log.Printf("No user found in session")

		// No valid user found, create a new one
		userID, err := s.db.Queries.CreateUser(ctx)
		log.Printf("Created new user: %v", userID)
		if err != nil {
			c.Logger().Errorf("Error creating user: %v", err)
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}

		// Save the user ID to the session
		sess.Values[sessionUserID] = userID.ID
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			c.Logger().Errorf("Error saving session: %v", err)
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}

		// Get the newly created user and set in context
		user, err := s.db.Queries.GetUser(ctx, userID.ID)
		if err != nil {
			c.Logger().Errorf("Error retrieving new user: %v", err)
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}

		c.Set("user", user)
		return next(c)
	}
}
