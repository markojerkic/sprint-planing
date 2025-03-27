package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/server/auth"
)

type JiraTicketResponse struct {
	Issues []JiraTicket `json:"issues"`
	Total  int          `json:"total"`
}

type JiraTicketParagraph struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type JiraTicketDescriptionContent struct {
	Type    string                `json:"type"`
	Content []JiraTicketParagraph `json:"content"`
}

type JiraTicketDescription struct {
	Type    string                         `json:"type"`
	Content []JiraTicketDescriptionContent `json:"content"`
}

func (j JiraTicketDescription) String() string {
	var desc string
	for _, c := range j.Content {
		for _, p := range c.Content {
			desc += p.Text
		}
	}
	return desc
}

type JiraTicketFields struct {
	Summary      string                `json:"summary"`
	TimeEstimate int                   `json:"timeestimate"`
	Description  JiraTicketDescription `json:"description"`
}

type JiraTicket struct {
	ID     string           `json:"id"`
	Key    string           `json:"key"`
	Fields JiraTicketFields `json:"fields"`
}

type JiraService struct {
}

var jiraKeyRegex = regexp.MustCompile(`[a-zA-Z]+-\d+`)

func (j *JiraService) GetIssues(ctx echo.Context, query string) (JiraTicketResponse, error) {
	clientInfo, ok := ctx.Get(auth.JiraClientInfoKey).(*auth.JiraClientInfo)
	if !ok || clientInfo.ResourceID == "" {
		slog.Error("Jira client info not found in context", slog.Any("clientInfo", clientInfo), slog.Any("ok", ok))
		return JiraTicketResponse{}, ctx.String(http.StatusInternalServerError, "Jira client info not found in context")
	}
	baseUrl := os.Getenv("JIRA_BASE_URL")
	url, err := url.Parse(fmt.Sprintf("%s/%s/rest/api/3/search?", baseUrl, clientInfo.ResourceID))
	if err != nil {
		slog.Error("Error parsing url", slog.Any("error", err))
		return JiraTicketResponse{}, err
	}

	q := url.Query()
	if query != "" {
		// Escape special JQL characters
		escapedQuery := strings.ReplaceAll(query, "\"", "\\\"")

		isKey := jiraKeyRegex.MatchString(query)

		jqlQuery := fmt.Sprintf("text ~ \"%s\"", escapedQuery)

		if isKey {
			jqlQuery += fmt.Sprintf(" OR key = \"%s\"", escapedQuery)
		}

		q.Set("jql", jqlQuery)

	}

	q.Set("maxResults", "50")

	url.RawQuery = q.Encode()

	slog.Info("URL", slog.Any("url", url.String()), slog.Any("query", url.Query().Encode()))

	resp, err := clientInfo.HttpClient(ctx).Get(url.String())
	if err != nil {
		slog.Error("Error getting issues", slog.Any("error", err))
		return JiraTicketResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to get issues", slog.Any("status", resp.StatusCode))
		if resp.Header.Get("Content-Type") == "application/json" {
			var errorResponse map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
				slog.Error("Failed to get issues", slog.Any("error", errorResponse))
			}
		}

		return JiraTicketResponse{}, fmt.Errorf("Failed to get issues")
	}

	var searchResult JiraTicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		slog.Error("Failed to decode issues", slog.Any("error", err))
		return JiraTicketResponse{}, err
	}

	return searchResult, nil
}

func (j *JiraService) UpdateTicketEstimation(ctx echo.Context, ticketKey string, estimation int) error {
	clientInfo, ok := ctx.Get(auth.JiraClientInfoKey).(*auth.JiraClientInfo)
	if !ok || clientInfo.ResourceID == "" {
		slog.Error("Jira client info not found in context", slog.Any("clientInfo", clientInfo), slog.Any("ok", ok))
		return fmt.Errorf("jira client info not found in context")
	}

	baseUrl := os.Getenv("JIRA_BASE_URL")
	url, err := url.Parse(fmt.Sprintf("%s/%s/rest/api/3/issue/%s", baseUrl, clientInfo.ResourceID, ticketKey))
	if err != nil {
		slog.Error("Error parsing url", slog.Any("error", err))
		return err
	}

	// Prepare the request body to update the time estimate
	slog.Info("Updating ticket estimation", slog.String("ticketKey", ticketKey), slog.String("estimation", fmt.Sprintf("%ds", estimation)))
	requestBody := map[string]any{
		"fields": map[string]any{
			"timetracking": map[string]any{
				"originalEstimate": estimation,
			},
		},
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		slog.Error("Error marshalling request body", slog.Any("error", err))
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url.String(), strings.NewReader(string(requestJSON)))
	if err != nil {
		slog.Error("Error creating request", slog.Any("error", err))
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := clientInfo.HttpClient(ctx).Do(req)
	if err != nil {
		slog.Error("Error updating ticket estimation", slog.Any("error", err))
		var errorResponse map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			slog.Error("Failed to update ticket estimation", slog.Any("error", errorResponse))
		}

		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		slog.Error("Failed to update ticket estimation", slog.Any("status", resp.StatusCode))

		// Try to parse error response if it's JSON
		var errorResponse map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			slog.Error("Failed to update ticket estimation", slog.Any("error", errorResponse))
		}

		return fmt.Errorf("failed to update ticket estimation: status code %d", resp.StatusCode)
	}

	slog.Info("Successfully updated ticket estimation",
		slog.String("ticketKey", ticketKey),
		slog.Int("estimationSeconds", estimation))

	return nil
}

func NewJiraService() *JiraService {
	return &JiraService{}
}
