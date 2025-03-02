package room

import (
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
)

type CreateTicketParams struct {
	TicketName        string `json:"ticketName" form:"ticketName" validate:"required"`
	TicketDescription string `json:"ticketDescription" form:"ticketDescription" validate:"required"`
	RoomID            int64  `json:"roomID" form:"roomID" validate:"required"`
}

func (r *RoomRouter) createTicketHandler(c echo.Context) error {
	var params CreateTicketParams
	if err := c.Bind(&params); err != nil {
		return c.String(400, "Invalid request")
	}
	if err := c.Validate(params); err != nil {
		c.Logger().Errorf("Error validating form: %v", err)
		c.Response().Header().Set("HX-Target", "#ticket-form")
		c.Response().Header().Set("HX-Swap", "beforeend")
		return c.HTML(400, "<div class='error-message'>Form validation failed. Please check your input.</div>")
	}

	return r.createTicket(c, params)
}

func (r *RoomRouter) createTicket(c echo.Context, params CreateTicketParams) error {
	tx, err := r.db.DB.BeginTx(c.Request().Context(), nil)
	if err != nil {
		c.Logger().Errorf("Error creating transaction: %v", err)
		return c.String(500, "Error creating transaction")
	}
	defer tx.Rollback()

	q := r.db.Queries.WithTx(tx)

	ticket, err := q.CreateTicket(c.Request().Context(), dbgen.CreateTicketParams{
		Name:        params.TicketName,
		Description: params.TicketDescription,
		RoomID:      params.RoomID,
	})
	if err != nil {
		tx.Rollback()
		c.Logger().Errorf("Error creating ticket: %v", err)
		return c.String(500, "Error creating ticket")
	}
	c.Logger().Infof("Created ticket: %v", ticket)

	tickets, err := q.GetTicketsOfRoom(c.Request().Context(), dbgen.GetTicketsOfRoomParams{
		RoomID: params.RoomID,
		UserID: c.Get("user").(dbgen.User).ID,
	})

	if err != nil {
		tx.Rollback()
		c.Logger().Errorf("Error getting tickets: %v", err)
		return c.String(500, "Error getting tickets")
	}

	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("Error committing transaction: %v", err)
		return c.String(500, "Error committing transaction")
	}
	r.sendNewTicket(ticket)

	return ticketList(tickets).Render(c.Request().Context(), c.Response().Writer)
}
