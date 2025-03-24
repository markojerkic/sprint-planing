package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/server/auth"
)

type JiraTicketResponse struct {
	Issues []JiraTicket `json:"issues"`
	Total  int          `json:"total"`
}

type JiraTicket struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
		Summary      string `json:"summary"`
		TimeEstimate int    `json:"timeestimate"`
	}
}

type JiraService struct {
}

func (j *JiraService) GetIssues(ctx echo.Context, query string) error {
	clientInfo, ok := ctx.Get(auth.JiraClientInfoKey).(auth.JiraClientInfo)
	if !ok || clientInfo.ResourceID == "" {
		slog.Error("Jira client info not found in context")
		return ctx.String(http.StatusInternalServerError, "Jira client info not found in context")
	}
	baseUrl := os.Getenv("JIRA_BASE_URL")
	url, err := url.Parse(fmt.Sprintf("%s/%s/rest/api/3/search?", baseUrl, clientInfo.ResourceID))
	if err != nil {
		slog.Error("Error parsing url", slog.Any("error", err))
		return err
	}
	if query != "" {
		q := url.Query()
		q.Set("jql", fmt.Sprintf("summary ~ \"%s\"", query))
		url.RawQuery = q.Encode()
	}
	slog.Info("URL", slog.Any("url", url.String()), slog.Any("query", url.Query().Encode()))

	resp, err := clientInfo.HttpClient(ctx.Request().Context()).Get(url.String())
	if err != nil {
		slog.Error("Error getting issues", slog.Any("error", err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to get issues", slog.Any("status", resp.StatusCode))
		return fmt.Errorf("Failed to get issues")
	}

	var searchResult JiraTicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		slog.Error("Failed to decode issues", slog.Any("error", err))
		return err
	}

	return ctx.JSON(http.StatusOK, searchResult)
}

func NewJiraService() *JiraService {
	return &JiraService{}
}
