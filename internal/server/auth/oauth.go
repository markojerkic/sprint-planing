package auth

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

const sessionName = "user_session"

type OAuthRouter struct {
	BaseURL string
	Config  oauth2.Config
}

func (o *OAuthRouter) Login(c echo.Context) error {
	state := uuid.New().String()
	authUrl := o.Config.AuthCodeURL(state)

	cookie := new(http.Cookie)
	cookie.Name = "oauth_state"
	cookie.Value = state
	cookie.Expires = time.Now().Add(5 * time.Minute)
	c.SetCookie(cookie)

	return c.Redirect(302, authUrl)
}

func (o *OAuthRouter) Callback(c echo.Context) error {
	// Verify state parameter
	state := c.QueryParam("state")
	cookie, err := c.Cookie("oauth_state")
	if err != nil || cookie.Value != state {
		slog.Error("Invalid state parameter", slog.Any("error", err), slog.String("cookie", cookie.Value), slog.String("state", state))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid state parameter"})
	}

	// Exchange authorization code for token
	code := c.QueryParam("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing authorization code"})
	}

	token, err := o.ExchangeCodeForToken(c.Request().Context(), code)
	if err != nil {
		slog.Error("Failed to exchange code for token", slog.Any("error", err), slog.String("code", code))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to exchange code for token"})
	}

	// Save token to session
	session, err := session.Get(sessionName, c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get session"})
	}
	session.Values["token"] = token
	if err := session.Save(c.Request(), c.Response()); err != nil {
		slog.Error("Failed to save session", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save session"})
	}

	return c.Redirect(http.StatusTemporaryRedirect, "/auth/jira/issues")
}

func (o *OAuthRouter) GetMyJiraIssues(c echo.Context) error {
	session, err := session.Get(sessionName, c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to get session"})
	}
	token := session.Values["token"].(*oauth2.Token)
	if token == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No token in session"})
	}
	if token.Expiry.Before(time.Now()) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Token expired"})
	}

	client := o.Config.Client(c.Request().Context(), token)
	encodedJql := url.QueryEscape("assignee = currentUser()")
	url := fmt.Sprintf("%s/rest/api/3/search/jql?jql=%s", o.BaseURL, encodedJql)
	resp, err := client.Get(url)
	if err != nil {
		slog.Error("Failed to get issues", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get issues"})
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read response"})
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to get issues", slog.Any("status", resp.StatusCode), slog.String("body", string(bodyBytes)))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get issues"})
	}

	var searchResult map[string]any
	if err := json.Unmarshal(bodyBytes, &searchResult); err != nil {
		slog.Error("Failed to decode issues", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode issues"})
	}

	return c.JSON(http.StatusOK, searchResult)
}

// ExchangeCodeForToken exchanges an authorization code for an access token
func (c *OAuthRouter) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %v", err)
	}
	return token, nil
}

func NewOAuthRouter(group *echo.Group) *OAuthRouter {
	router := OAuthRouter{
		BaseURL: os.Getenv("OAUTH_BASE_URL"),
		Config: oauth2.Config{
			ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
			ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://auth.atlassian.com/authorize",
				TokenURL: "https://auth.atlassian.com/oauth/token",
			},
			Scopes: []string{"read:jira-work", "write:jira-work"},
		},
	}
	gob.Register(&oauth2.Token{})

	slog.Info("NewOAuthRouter",
		slog.String("BaseURL", router.BaseURL),
		slog.String("ClientID", router.Config.ClientID),
		slog.String("ClientSecret", router.Config.ClientSecret),
		slog.String("RedirectURL", router.Config.RedirectURL),
	)

	e := group
	e.GET("/login", router.Login)
	e.GET("/response", router.Callback)
	e.GET("/issues", router.GetMyJiraIssues)

	return &router
}
