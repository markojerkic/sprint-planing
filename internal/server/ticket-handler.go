package server

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
	"github.com/markojerkic/spring-planing/internal/service"
)

type TicketRouter struct {
	service *service.TicketService
	group   *echo.Group
}

func (r *TicketRouter) estimateTicketHandler(c echo.Context) error {
	var form service.EstimateTicketForm
	if err := c.Bind(&form); err != nil {
		return c.String(400, "Invalid request")
	}
	if err := c.Validate(form); err != nil {
		c.Logger().Errorf("Error validating form: %v", err)
		return c.String(400, "Form validation failed. Please check your input.")
	}
	user := c.Get("user").(dbgen.User)

	estimate, err := r.service.EstimateTicket(c.Request().Context(), user.ID, form)
	if err != nil {
		c.Logger().Errorf("Error estimating ticket: %v", err)
		return c.String(500, "Error estimating ticket")
	}

	return ticket.UsersEstimate(form.TicketID, estimate).Render(c.Request().Context(), c.Response().Writer)
}

func (r *TicketRouter) createTicketHandler(c echo.Context) error {
	var form service.CreateTicketForm
	if err := c.Bind(&form); err != nil {
		return c.String(400, "Invalid request")
	}
	if err := c.Validate(form); err != nil {
		c.Logger().Errorf("Error validating form: %v", err)
		return c.HTML(400, "<div class='error-message'>Form validation failed. Please check your input.</div>")
	}

	user := c.Get("user").(dbgen.User)

	ticketList, err := r.service.CreateTicket(c.Request().Context(), user.ID, form)
	if err != nil {
		c.Logger().Errorf("Error creating ticket: %v", err)
		return c.String(500, "Error creating ticket")
	}

	return ticket.TicketList(ticketList, true).Render(c.Request().Context(), c.Response().Writer)
}

func (r *TicketRouter) closeTicketHandler(c echo.Context) error {
	ticketID, err := strconv.ParseInt(c.FormValue("id"), 10, 64)
	if err != nil {
		return c.String(400, "Invalid ticket id")
	}
	if _, err := r.service.CloseTicket(c.Request().Context(), ticketID); err != nil {
		return c.String(500, "Error closing ticket")
	}

	return c.String(200, fmt.Sprintf("Closed ticket %d", ticketID))
}

func newTicketRouter(db *database.Database, group *echo.Group) *TicketRouter {
	r := &TicketRouter{
		service: service.NewTicketService(db),
		group:   group,
	}
	e := r.group

	e.POST("", r.createTicketHandler)
	e.POST("/estimate", r.estimateTicketHandler)
	e.POST("/close", r.closeTicketHandler)

	return r
}

func prettyPrintEstimate(nEstimate sql.NullInt64) string {
	if !nEstimate.Valid {
		return "No estimate"
	}
	estimate := nEstimate.Int64

	weeks := estimate / 40
	days := (estimate % 40) / 8
	hours := estimate % 8

	return fmt.Sprintf("%dw %dd %dh", weeks, days, hours)
}
