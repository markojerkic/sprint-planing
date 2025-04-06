package auth

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

// Custom TokenSource that updates session when token changes
type sessionUpdatingTokenSource struct {
	source     oauth2.TokenSource
	echoCtx    echo.Context
	resourceID string
	original   *oauth2.Token
}

func (j *JiraClientInfo) HttpClient(c echo.Context) *http.Client {
	originalToken := &oauth2.Token{
		AccessToken:  j.AccessToken,
		TokenType:    "Bearer",
		RefreshToken: j.RefreshToken,
		Expiry:       j.Expiry,
	}

	ctx := c.Request().Context()
	// Create a token source with the original token
	tokenSource := oauthConf.TokenSource(ctx, originalToken)

	// Wrap it with a custom token source that can update the session
	wrappedSource := &sessionUpdatingTokenSource{
		source:     tokenSource,
		echoCtx:    c,
		resourceID: j.ResourceID,
		original:   originalToken,
	}

	return oauth2.NewClient(ctx, wrappedSource)

}

func (s *sessionUpdatingTokenSource) Token() (*oauth2.Token, error) {
	// Get token, which may trigger a refresh
	newToken, err := s.source.Token()
	if err != nil {
		return nil, err
	}

	// If token changed (refreshed), update the session
	if newToken.AccessToken != s.original.AccessToken {
		slog.Debug("Token refreshed, updating session",
			slog.String("old_token", s.original.AccessToken[:10]+"..."),
			slog.String("new_token", newToken.AccessToken[:10]+"..."))

		clientInfo := JiraClientInfo{
			ResourceID:   s.resourceID,
			AccessToken:  newToken.AccessToken,
			RefreshToken: newToken.RefreshToken,
			Expiry:       newToken.Expiry,
		}

		if err := saveJiraInfoToSession(s.echoCtx, clientInfo); err != nil {
			slog.Error("Failed to save refreshed token to session",
				slog.Any("error", err))
		}

		// Update original token reference
		s.original = newToken
	}

	return newToken, nil
}
