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
	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
	"github.com/markojerkic/spring-planing/internal/server/auth"
	"github.com/markojerkic/spring-planing/internal/util"
)

type JiraTicketResponse struct {
	Issues []JiraTicket `json:"issues"`
	Total  int          `json:"total"`
	IsLast bool         `json:"isLast"`
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

type JiraProjectSearchResult struct {
	Values []ticket.JiraProject `json:"values"`
	Total  int                  `json:"total"`
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
	ticketService *TicketService
}

var jiraKeyRegex = regexp.MustCompile(`[a-zA-Z]+-\d+`)

// cal /rest/api/3/serverInfo and get field baseUrl
func (j *JiraService) GetResourceServerBaseUrl(ctx echo.Context) (string, error) {

	clientInfo, ok := ctx.Get(auth.JiraClientInfoKey).(*auth.JiraClientInfo)
	if !ok || clientInfo.ResourceID == "" {
		slog.Error("Jira client info not found in context", slog.Any("clientInfo", clientInfo), slog.Any("ok", ok))
		return "", ctx.String(http.StatusInternalServerError, "Jira client info not found in context")
	}
	baseUrl := os.Getenv("JIRA_BASE_URL")
	url, err := url.Parse(fmt.Sprintf("%s/%s/rest/api/3/serverInfo", baseUrl, clientInfo.ResourceID))
	if err != nil {
		slog.Error("Error parsing url", slog.Any("error", err))
		return "", err
	}

	q := url.Query()
	q.Set("maxResults", "50")

	url.RawQuery = q.Encode()

	slog.Debug("URL", slog.Any("url", url.String()), slog.Any("query", url.Query().Encode()))

	resp, err := clientInfo.HttpClient(ctx).Get(url.String())
	if err != nil {
		slog.Error("Error getting issues", slog.Any("error", err))
		return "", err
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

		return "", fmt.Errorf("Failed to get issues")
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("Failed to decode issues", slog.Any("error", err))
		return "", err
	}

	return result["baseUrl"].(string), nil

}

// Batch import tickets from Jira
func (j *JiraService) BulkImportTickets(ctx echo.Context, userID uint, roomID uint, filter JiraIssueFilter) ([]ticket.TicketDetailProps, error) {
	issues, err := j.GetIssues(ctx, filter)
	if err != nil {
		return nil, err
	}

	tickets := make([]CreateTicketForm, len(issues.Issues))
	for i, t := range issues.Issues {
		tickets[i] = CreateTicketForm{
			TicketName:        t.Key,
			TicketDescription: t.Fields.Summary,
			RoomID:            roomID,
			JiraKey:           t.Key,
		}
	}

	ticketDetails, err := j.ticketService.BulkImportTickets(ctx.Request().Context(), userID, roomID, tickets)
	if err != nil {
		slog.Error("Error bulk importing tickets", slog.Any("error", err))
		return nil, err
	}

	return ticketDetails, nil
}

func (j *JiraService) GetProjectStories(ctx echo.Context, projectID string) (JiraTicketResponse, error) {
	return j.GetIssues(ctx, JiraIssueFilter{
		ProjectID: projectID,
		IssueType: "story",
	})
}

func (j *JiraService) GetProjects(ctx echo.Context) ([]ticket.JiraProject, error) {
	clientInfo, ok := ctx.Get(auth.JiraClientInfoKey).(*auth.JiraClientInfo)
	if !ok || clientInfo.ResourceID == "" {
		slog.Error("Jira client info not found in context", slog.Any("clientInfo", clientInfo), slog.Any("ok", ok))
		return nil, ctx.String(http.StatusInternalServerError, "Jira client info not found in context")
	}
	baseUrl := os.Getenv("JIRA_BASE_URL")
	url, err := url.Parse(fmt.Sprintf("%s/%s/rest/api/3/project", baseUrl, clientInfo.ResourceID))
	if err != nil {
		slog.Error("Error parsing url", slog.Any("error", err))
		return nil, err
	}

	q := url.Query()
	q.Set("maxResults", "50")

	url.RawQuery = q.Encode()

	slog.Debug("URL", slog.Any("url", url.String()), slog.Any("query", url.Query().Encode()))

	resp, err := clientInfo.HttpClient(ctx).Get(url.String())
	if err != nil {
		slog.Error("Error getting issues", slog.Any("error", err))
		return nil, err
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

		return nil, fmt.Errorf("Failed to get issues")
	}

	var projects []ticket.JiraProject
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		slog.Error("Failed to decode issues", slog.Any("error", err))
		return nil, err
	}

	return projects, nil
}

type JiraIssueFilter struct {
	IssueType         string `json:"issueType" form:"jira-issue-type" query:"jira-issue-type"`
	HasEstimate       string `json:"hasEstimate" form:"has-estimate" query:"has-estimate"`
	Query             string `json:"query" form:"q" query:"q"`
	ProjectID         string `json:"projectId" form:"jira-project-id" query:"jira-project-id"`
	StoryKey          string `json:"jiraStory" form:"jira-story" query:"jira-story"`
	CreatedWithinDays string `json:"createdWithinDays" form:"created-within-days" query:"created-within-days"`
	HasAssignee       string `json:"hasAssignee" form:"has-assignee" query:"has-assignee"`
}

func (j *JiraService) GetIssues(ctx echo.Context, filter JiraIssueFilter) (JiraTicketResponse, error) {
	clientInfo, ok := ctx.Get(auth.JiraClientInfoKey).(*auth.JiraClientInfo)
	if !ok || clientInfo.ResourceID == "" {
		slog.Error("Jira client info not found in context", slog.Any("clientInfo", clientInfo), slog.Any("ok", ok))
		return JiraTicketResponse{}, ctx.String(http.StatusInternalServerError, "Jira client info not found in context")
	}
	baseUrl := os.Getenv("JIRA_BASE_URL")
	url, err := url.Parse(fmt.Sprintf("%s/%s/rest/api/3/search/jql?", baseUrl, clientInfo.ResourceID))
	if err != nil {
		slog.Error("Error parsing url", slog.Any("error", err))
		return JiraTicketResponse{}, err
	}

	q := url.Query()
	jqlQueries := make([]string, 0)
	if filter.Query != "" {
		// Escape special JQL characters
		escapedQuery := strings.ReplaceAll(filter.Query, "\"", "\\\"")

		isKey := jiraKeyRegex.MatchString(filter.Query)

		// Search in text and summary fields
		jqlQuery := fmt.Sprintf("text ~ \"%s\" OR summary ~ \"%s\"", escapedQuery, escapedQuery)

		if isKey {
			// Exact key match
			jqlQuery += fmt.Sprintf(" OR key = \"%s\"", escapedQuery)
		} else {
			// For partial key searches, try a broader approach
			upperQuery := strings.ToUpper(escapedQuery)
			jqlQuery += fmt.Sprintf(" OR key ~ \"%s\" OR summary ~ \"%s\"", upperQuery, upperQuery)
		}

		jqlQueries = append(jqlQueries, fmt.Sprintf("(%s)", jqlQuery))
	}

	if filter.ProjectID != "" {
		jqlQueries = append(jqlQueries, fmt.Sprintf("project = %s", filter.ProjectID))
	}

	if filter.IssueType != "" {
		switch filter.IssueType {
		case "all":
			jqlQueries = append(jqlQueries, "type IN (story, task, subtask, bug)")
		case "task":
			jqlQueries = append(jqlQueries, "type IN (task, subtask)")
		default:
			jqlQueries = append(jqlQueries, fmt.Sprintf("type = %s", filter.IssueType))
		}
	}

	switch filter.HasEstimate {
	case "yes":
		jqlQueries = append(jqlQueries, "originalEstimate != 0")
	case "no":
		jqlQueries = append(jqlQueries, "(originalEstimate = 0 OR originalEstimate IS EMPTY)")
	}

	if filter.StoryKey != "" {
		jqlQueries = append(jqlQueries, fmt.Sprintf("(parent = \"%s\" OR key = \"%s\")", filter.StoryKey, filter.StoryKey))
	}

	if filter.CreatedWithinDays != "" {
		jqlQueries = append(jqlQueries, fmt.Sprintf("created >= -%sd", filter.CreatedWithinDays))
	}

	switch filter.HasAssignee {
	case "yes":
		jqlQueries = append(jqlQueries, "assignee IS NOT EMPTY")
	case "no":
		jqlQueries = append(jqlQueries, "assignee IS EMPTY")
	}

	jqlQuery := strings.Join(jqlQueries, " AND ")

	// If no filters provided, add a default condition to get recent issues
	if jqlQuery == "" {
		jqlQuery = "created >= -30d ORDER BY created DESC"
	}

	slog.Debug("JQL", slog.String("jql", jqlQuery))
	q.Set("jql", jqlQuery)

	q.Set("maxResults", "75")
	q.Set("fields", "summary,description,key,id")

	url.RawQuery = q.Encode()

	slog.Debug("URL", slog.Any("url", url.String()), slog.Any("query", url.Query().Encode()))

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
	slog.Debug("Updating ticket estimation", slog.String("ticketKey", ticketKey), slog.String("estimation", fmt.Sprintf("%ds", estimation)))
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
			slog.Error("Failed to connect to JIRA", slog.Any("error", errorResponse))
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
			if _, hasErrors := errorResponse["errors"]; hasErrors {
				if _, isMap := errorResponse["errors"].(map[string]any); isMap {
					if _, hasTimetrackingError := errorResponse["errors"].(map[string]any)["timetracking"]; hasTimetrackingError {
						util.AddToastHeader(ctx, "Failed writing time estimation. Check to make sure you're not using story points!", util.ERROR)
					}
				}
			}

		} else {
			slog.Error("Error parsing jira error", slog.Any("error", err))
		}

		return fmt.Errorf("failed to update ticket estimation: status code %d", resp.StatusCode)
	}

	slog.Debug("Successfully updated ticket estimation",
		slog.String("ticketKey", ticketKey),
		slog.Int("estimationSeconds", estimation))

	return nil
}

func (j *JiraService) GetTicketDescription(ctx echo.Context, ticketKey string) (string, error) {
	clientInfo, ok := ctx.Get(auth.JiraClientInfoKey).(*auth.JiraClientInfo)
	if !ok || clientInfo.ResourceID == "" {
		slog.Error("Jira client info not found in context", slog.Any("clientInfo", clientInfo), slog.Any("ok", ok))
		return "", fmt.Errorf("jira client info not found in context")
	}

	baseUrl := os.Getenv("JIRA_BASE_URL")
	url, err := url.Parse(fmt.Sprintf("%s/%s/rest/api/3/issue/%s", baseUrl, clientInfo.ResourceID, ticketKey))
	if err != nil {
		slog.Error("Error parsing url", slog.Any("error", err))
		return "", err
	}

	q := url.Query()
	q.Set("fields", "description")
	url.RawQuery = q.Encode()

	slog.Debug("Getting ticket description", slog.String("url", url.String()))

	resp, err := clientInfo.HttpClient(ctx).Get(url.String())
	if err != nil {
		slog.Error("Error getting ticket", slog.Any("error", err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to get ticket", slog.Any("status", resp.StatusCode))
		if resp.Header.Get("Content-Type") == "application/json" {
			var errorResponse map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
				slog.Error("Failed to get ticket", slog.Any("error", errorResponse))
			}
		}
		return "", fmt.Errorf("failed to get ticket: status code %d", resp.StatusCode)
	}

	var ticket JiraTicket
	if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
		slog.Error("Failed to decode ticket", slog.Any("error", err))
		return "", err
	}

	return ticket.Fields.Description.String(), nil
}

func NewJiraService(ticketService *TicketService) *JiraService {
	if ticketService == nil {
		panic("ticketService cannot be nil")
	}
	return &JiraService{
		ticketService: ticketService,
	}
}
