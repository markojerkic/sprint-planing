package server

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
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
	return errors.New("not implemented")
	// var form service.CreateTicketForm
	// if err := c.Bind(&form); err != nil {
	// 	return c.String(400, "Invalid request")
	// }
	// if err := c.Validate(form); err != nil {
	// 	c.Logger().Errorf("Error validating form: %v", err)
	// 	return c.HTML(400, "<div class='error-message'>Form validation failed. Please check your input.</div>")
	// }
	//
	// user := c.Get("user").(dbgen.User)
	//
	// ticketList, err := r.service.CreateTicket(c.Request().Context(), user.ID, form)
	// if err != nil {
	// 	c.Logger().Errorf("Error creating ticket: %v", err)
	// 	return c.String(500, "Error creating ticket")
	// }
	//
	// return ticket.TicketList(ticketList, true).Render(c.Request().Context(), c.Response().Writer)
}

func (r *TicketRouter) ticketEstimatesHandler(c echo.Context) error {
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(400, "Invalid ticket id")
	}
	estimates, err := r.service.GetTicketEstimates(c.Request().Context(), int32(ticketID))
	if err != nil {
		return c.String(500, "Error getting ticket estimates")
	}

	return ticket.EstimatesPopupContent(estimates).Render(c.Request().Context(), c.Response().Writer)
}

func (r *TicketRouter) closeTicketHandler(c echo.Context) error {
	ticketID, err := strconv.ParseInt(c.FormValue("id"), 10, 64)
	if err != nil {
		return c.String(400, "Invalid ticket id")
	}
	if _, err := r.service.CloseTicket(c.Request().Context(), int32(ticketID)); err != nil {
		return c.String(500, "Error closing ticket")
	}

	return c.String(200, fmt.Sprintf("Closed ticket %d", ticketID))
}

func (r *TicketRouter) hideTicketHandler(c echo.Context) error {
	c.Logger().Info("Hiding ticket")
	ticketID, err := strconv.ParseInt(c.FormValue("id"), 10, 64)
	if err != nil {
		return c.String(400, "Invalid ticket id")
	}
	updatedTicket, err := r.service.HideTicket(c.Request().Context(), int32(ticketID))
	if err != nil {
		return c.String(500, "Error hiding ticket")
	}

	return ticket.HideToggle(int32(ticketID), updatedTicket.Hidden).Render(c.Request().Context(), c.Response().Writer)
}

func newTicketRouter(ticketService *service.TicketService, group *echo.Group) *TicketRouter {
	r := &TicketRouter{
		service: ticketService,
		group:   group,
	}
	e := r.group

	e.POST("", r.createTicketHandler)
	e.POST("/hide", r.hideTicketHandler)
	e.POST("/estimate", r.estimateTicketHandler)
	e.POST("/close", r.closeTicketHandler)
	e.GET("/estimates/:id", r.ticketEstimatesHandler)

	return r
}
