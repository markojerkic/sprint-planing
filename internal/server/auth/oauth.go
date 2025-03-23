package auth

import (
	"log/slog"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

type OAuthRouter struct {
	BaseURL string
	Config  oauth2.Config
}

func (o *OAuthRouter) Login(c echo.Context) error {
	state := time.Now().String()
	authUrl := o.Config.AuthCodeURL(state)

	return c.Redirect(302, authUrl)
}

func (o *OAuthRouter) Callback(c echo.Context) error {
	slog.Info("Callback",
		slog.Any("query", c.QueryParams()),
		slog.Any("header", c.Request().Header))

	return c.String(200, "Callback")
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

	slog.Info("NewOAuthRouter",
		slog.String("BaseURL", router.BaseURL),
		slog.String("ClientID", router.Config.ClientID),
		slog.String("ClientSecret", router.Config.ClientSecret),
		slog.String("RedirectURL", router.Config.RedirectURL),
	)

	e := group
	e.GET("/login", router.Login)
	e.GET("/response", router.Callback)

	return &router
}
