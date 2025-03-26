package server

import (
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
	"github.com/markojerkic/spring-planing/internal/service"
	"gorm.io/gorm"
)

type JiraRouter struct {
	jiraService *service.JiraService
	db          *gorm.DB
	group       *echo.Group
}

func (j *JiraRouter) searchIssuesHandler(ctx echo.Context) error {
	issues, err := j.jiraService.GetIssues(ctx, ctx.QueryParam("q"))
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

func newJiraRouter(jiraService *service.JiraService, db *gorm.DB, group *echo.Group) *JiraRouter {
	router := &JiraRouter{
		jiraService: jiraService,
		db:          db,
		group:       group,
	}

	router.group.GET("/search", router.searchIssuesHandler)

	return router
}
