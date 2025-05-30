package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/server/auth"
	"github.com/markojerkic/spring-planing/internal/service"
	"github.com/markojerkic/spring-planing/internal/util"
	"gorm.io/gorm"
)

type JiraRouter struct {
	jiraService   *service.JiraService
	ticketService *service.TicketService
	db            *gorm.DB
	group         *echo.Group
}

func (j *JiraRouter) bulkImportJiraTicketsHandler(ctx echo.Context) error {
	var filter service.JiraIssueFilter
	if err := ctx.Bind(&filter); err != nil {
		return ctx.String(400, "Invalid filter")
	}

	if err := ctx.Validate(filter); err != nil {
		return ctx.String(400, "Invalid filter")
	}

	user := ctx.Get("user").(database.User)
	sroomId := ctx.FormValue("roomId")
	roomID, err := strconv.Atoi(sroomId)
	if err != nil {
		return ctx.String(400, "Invalid room id")
	}

	tickets, err := j.jiraService.BulkImportTickets(ctx, user.ID, uint(roomID), filter)
	if err != nil {
		return ctx.String(500, "Error bulk importing tickets")
	}

	ctx.Response().Header().Add("Hx-Trigger", `{"createdTicket": true}`)

	util.AddToastHeader(ctx, "Ticket created successfully", util.INFO)

	return ticket.TicketList(tickets, true).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (j *JiraRouter) getBulkImportSearchResultsHandler(ctx echo.Context) error {
	var filter service.JiraIssueFilter
	if err := ctx.Bind(&filter); err != nil {
		return ctx.String(400, "Invalid filter")
	}

	slog.Debug("Getting search results", slog.Any("filter", filter))
	issues, err := j.jiraService.GetIssues(ctx, filter)
	if err != nil {
		return ctx.String(500, "Error getting issues")
	}

	jiraTickes := make([]ticket.JiraTicket, len(issues.Issues))
	for i, t := range issues.Issues {
		jiraTickes[i] = ticket.JiraTicket{
			Key:         t.Key,
			Summary:     t.Fields.Summary,
			Description: t.Fields.Description.String(),
		}
	}

	return ticket.JiraSearchTicketList(ticket.JiraTicketListProps{
		Tickets: jiraTickes,
	}).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (j *JiraRouter) getProjectStoriesHandler(ctx echo.Context) error {
	projectID := ctx.QueryParam("jira-project-id")

	result, err := j.jiraService.GetProjectStories(ctx, projectID)
	if err != nil {
		return ctx.String(500, "Error getting stories")
	}
	issueType := ctx.QueryParam("jira-issue-type")

	if issueType != "all" && issueType != "task" {
		// Retun empty string
		return ctx.String(200, "")
	}

	stories := make([]ticket.JiraTicket, len(result.Issues))
	for i, t := range result.Issues {
		stories[i] = ticket.JiraTicket{
			Key:         t.Key,
			Summary:     t.Fields.Summary,
			Description: t.Fields.Description.String(),
		}
	}

	return ticket.JiraStorySelect(stories).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (j *JiraRouter) getProjectsHandler(ctx echo.Context) error {
	roomID, err := strconv.Atoi(ctx.QueryParam("roomId"))
	if err != nil {
		return ctx.String(400, "Invalid room id")
	}

	projects, err := j.jiraService.GetProjects(ctx)
	if err != nil {
		return ctx.String(500, "Error getting projects")
	}

	return ticket.BulkImportJiraTicketsForm(ticket.BulkImportJiraTicketsProps{
		RoomId:       uint(roomID),
		JiraProjects: projects,
	}).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (j *JiraRouter) searchIssuesHandler(ctx echo.Context) error {
	var filter service.JiraIssueFilter
	if err := ctx.Bind(&filter); err != nil {
		return ctx.String(400, "Invalid filter")
	}

	issues, err := j.jiraService.GetIssues(ctx, filter)
	if err != nil {
		return ctx.String(500, "Error getting issues")
	}

	jiraTickes := make([]ticket.JiraTicket, len(issues.Issues))
	for i, t := range issues.Issues {
		jiraTickes[i] = ticket.JiraTicket{
			Key:         t.Key,
			Summary:     t.Fields.Summary,
			Description: t.Fields.Description.String(),
		}
	}

	return ticket.JiraTicketList(ticket.JiraTicketListProps{
		Tickets: jiraTickes,
	}).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (j *JiraRouter) writeEstimate(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.FormValue("id"))
	if err != nil {
		return ctx.String(400, "Invalid ticket id")
	}
	user, ok := ctx.Get("user").(database.User)
	if !ok {
		slog.Error("Error getting user from context")
		return ctx.String(500, "Error getting user")
	}

	ticket, err := j.ticketService.GetTicket(ctx.Request().Context(), j.db, user.ID, nil, uint(id))
	if err != nil {
		slog.Error("Error getting ticket for estimate", slog.Any("error", err))
	}

	if ticket.JiraKey == nil {
		return ctx.String(400, "Ticket is not linked to Jira")
	}

	var estimateHours int
	estimateType := ctx.Param("type")

	if estimateType == "median" {
		estimateHours = int(ticket.MedianEstimate)
	} else if estimateType == "average" {
		estimateHours = int(ticket.AverageEstimate)
	} else {
		return ctx.String(400, "Invalid estimate type")
	}

	slog.Debug("Updating ticket", slog.Any("estimateHours", estimateHours), slog.Int("estimateSeconds", estimateHours*60))
	if err := j.jiraService.UpdateTicketEstimation(ctx, *ticket.JiraKey, estimateHours*60); err != nil {
		slog.Error("Error updating ticket", slog.Any("error", err))
		return ctx.String(500, "Error updating ticket")
	}

	util.AddToastHeader(ctx, "Estimate successfully written to Jira!", util.INFO)

	return ctx.String(200, "<div>Estimate updated!</div>")
}

func (j *JiraRouter) redirectToJiraIssueHandler(ctx echo.Context) error {
	issueKey := ctx.Param("issueKey")
	resourceBseUrl, err := j.jiraService.GetResourceServerBaseUrl(ctx)
	if err != nil {
		return ctx.String(500, "Error getting resource ID")
	}

	slog.Debug("Redirecting to Jira issue", slog.String("issueKey", issueKey), slog.String("resourceBseUrl", resourceBseUrl))
	return ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/browse/%s", resourceBseUrl, issueKey))

}

func newJiraRouter(jiraService *service.JiraService, db *gorm.DB, group *echo.Group) *JiraRouter {
	router := &JiraRouter{
		jiraService: jiraService,
		db:          db,
		group:       group,
	}

	router.group.Use(auth.JiraAuthMiddleware)

	router.group.GET("/:issueKey", router.redirectToJiraIssueHandler)
	router.group.GET("/search", router.searchIssuesHandler)
	router.group.POST("/ticket/:type", router.writeEstimate)
	router.group.GET("/projects-form", router.getProjectsHandler)
	router.group.GET("/project-stories", router.getProjectStoriesHandler)
	router.group.GET("/bulk/search-results", router.getBulkImportSearchResultsHandler)
	router.group.POST("/bulk/import", router.bulkImportJiraTicketsHandler)

	return router
}
