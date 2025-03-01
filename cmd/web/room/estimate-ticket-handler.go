package room

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
)

type EstimateTicketParams struct {
	TicketID     int64 `json:"ticketID" form:"ticketID" validate:"required"`
	WeekEstimate int64 `json:"weekEstimate" form:"weekEstimate" default:"0"`
	DayEstimate  int64 `json:"dayEstimate" form:"dayEstimate" default:"0"`
	HourEstimate int64 `json:"hourEstimate" form:"hourEstimate" default:"0"`
}

func (r *RoomRouter) estimateTicketHandler(c echo.Context) error {
	log.Println("Estimating ticket")
	user := c.Get("user").(dbgen.User)
	var estimateParams EstimateTicketParams
	if err := c.Bind(&estimateParams); err != nil {
		return c.String(400, "Invalid request")
	}
	if err := c.Validate(estimateParams); err != nil {
		c.Logger().Errorf("Error validating form: %v", err)
		return c.String(400, "Form validation failed. Please check your input.")
	}

	log.Printf("Estimating ticket %+v", estimateParams)

	return r.estimateTicket(c, estimateParams, user)
}

func (r *RoomRouter) estimateTicket(c echo.Context, params EstimateTicketParams, user dbgen.User) error {
	tx, err := r.db.DB.BeginTx(c.Request().Context(), nil)
	if err != nil {
		c.Logger().Errorf("Error creating transaction: %v", err)
		return c.String(500, "Error creating transaction")
	}
	defer tx.Rollback()

	q := r.db.Queries.WithTx(tx)

	// Assumes a day is 8 hours, and a week is 5 days
	estimate, err := q.EstimateTicket(c.Request().Context(), dbgen.EstimateTicketParams{
		Estimate: params.WeekEstimate*5*8 + params.DayEstimate*8 + params.HourEstimate,
		UserID:   user.ID,
		TicketID: params.TicketID,
	})
	if err != nil {
		tx.Rollback()
		c.Logger().Errorf("Error estimating ticket: %v", err)
		return c.String(500, "Error estimating ticket")
	}

	prettyEstimate := formatEstimate(estimate.Estimate)
	log.Printf("Estimated ticket: %+v from %+v", prettyEstimate, estimate)

	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("Error committing transaction: %v", err)
		return c.String(500, "Error committing transaction")
	}

	go func() {
		r.updateAverageEstimateForTicket(params.TicketID)
	}()

	return myEstimation(params.TicketID, prettyEstimate).Render(c.Request().Context(), c.Response().Writer)
}

func formatEstimate(estimate int64) string {
	weeks := estimate / (5 * 8)
	days := (estimate % (5 * 8)) / 8
	hours := estimate % 8

	return fmt.Sprintf("%dw %dd %dh", weeks, days, hours)
}
