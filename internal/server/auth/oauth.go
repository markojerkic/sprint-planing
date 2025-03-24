package auth

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

const (
	sessionName         = "user_session"
	sessionAccessToken  = "jira_access_token"
	sessionRefreshToken = "jira_refresh_token"
	sessionJiraDomain   = "jira_domain"
	JiraClientInfoKey   = "jira_client_info"
)

var oauthConf = oauth2.Config{
	ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://auth.atlassian.com/authorize",
		TokenURL: "https://auth.atlassian.com/oauth/token",
	},
	Scopes: []string{"read:jira-work", "write:jira-work"},
}

type JiraClientInfo struct {
	ResourceID   string
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

func (j *JiraClientInfo) HttpClient(ctx context.Context) *http.Client {
	return oauthConf.Client(ctx, &oauth2.Token{
		AccessToken:  j.AccessToken,
		TokenType:    "Bearer",
		RefreshToken: j.RefreshToken,
		Expiry:       j.Expiry,
	})
}

type JiraResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

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

	// Accessable resources
	client := o.Config.Client(c.Request().Context(), token)
	resp, err := client.Get("https://api.atlassian.com/oauth/token/accessible-resources")
	if err != nil {
		slog.Error("Failed to get accessible resources", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get accessible resources"})
	}

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read response"})
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to get accessible resources", slog.Any("status", resp.StatusCode), slog.String("body", string(bodyBytes)))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get accessible resources"})
	}

	var accessibleResources []JiraResource
	if err := json.Unmarshal(bodyBytes, &accessibleResources); err != nil {
		slog.Error("Failed to decode accessible resources", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode accessible resources"})
	}
	slog.Info("Accessible resources", slog.Any("resources", accessibleResources))

	session.Values[JiraClientInfoKey] = JiraClientInfo{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		ResourceID:   accessibleResources[0].ID,
	}

	if err := session.Save(c.Request(), c.Response()); err != nil {
		slog.Error("Failed to save session", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save session"})
	}

	slog.Info("Accessible resources", slog.Any("resources", accessibleResources))

	return c.Redirect(http.StatusTemporaryRedirect, "/jira/issues")
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
		Config:  oauthConf,
	}
	gob.Register(oauth2.Token{})
	gob.Register(JiraClientInfo{})

	slog.Info("NewOAuthRouter",
		slog.String("BaseURL", router.BaseURL),
		slog.String("ClientID", router.Config.ClientID),
		slog.String("ClientSecret", router.Config.ClientSecret),
		slog.String("RedirectURL", router.Config.RedirectURL),
	)

	e := group
	e.GET("/login", router.Login)
	e.GET("/response", router.Callback)
	// e.GET("/issues", router.GetMyJiraIssues)

	return &router
}
