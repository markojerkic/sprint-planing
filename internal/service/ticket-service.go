package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
)

type TicketService struct {
	db               *database.Database
	webSocketService *WebSocketService
}

type CreateTicketForm struct {
	TicketName        string `json:"ticketName" form:"ticketName" validate:"required"`
	TicketDescription string `json:"ticketDescription" form:"ticketDescription" validate:"required"`
	RoomID            int64  `json:"roomID" form:"roomID" validate:"required"`
}

type EstimateTicketForm struct {
	TicketID     int64 `json:"ticketID" form:"ticketID" validate:"required"`
	WeekEstimate int64 `json:"weekEstimate" form:"weekEstimate" default:"0"`
	DayEstimate  int64 `json:"dayEstimate" form:"dayEstimate" default:"0"`
	HourEstimate int64 `json:"hourEstimate" form:"hourEstimate" default:"0"`
}

func (t *TicketService) EstimateTicket(ctx context.Context, userID int64, form EstimateTicketForm) (string, error) {
	tx, err := t.db.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Error creating transaction", slog.Any("error", err))
		return "", err
	}
	defer tx.Rollback()

	q := t.db.Queries.WithTx(tx)

	// Assumes a day is 8 hours, and a week is 5 days
	estimate, err := q.EstimateTicket(ctx, dbgen.EstimateTicketParams{
		Estimate: form.WeekEstimate*5*8 + form.DayEstimate*8 + form.HourEstimate,
		UserID:   userID,
		TicketID: form.TicketID,
	})
	if err != nil {
		tx.Rollback()
		slog.Error("Error estimating ticket", slog.Any("error", err))
		return "", err
	}

	updatedTickets, err := q.GetTicketAverageEstimation(ctx, form.TicketID)
	if err != nil {
		tx.Rollback()
		slog.Error("Error getting ticket average estimation", slog.Any("error", err))
		return "", err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("Error committing transaction", slog.Any("error", err))
		return "", err
	}

	averageEstimate := updatedTickets.AvgEstimate.Float64
	t.webSocketService.UpdateEstimate(updatedTickets.ID,
		updatedTickets.RoomID,
		prettyPrintEstimate(sql.NullInt64{Int64: int64(averageEstimate), Valid: true}),
		fmt.Sprintf("%d/%d", updatedTickets.UsersEstimated, updatedTickets.TotalUsersInRoom))

	return prettyPrintEstimate(sql.NullInt64{Int64: estimate.Estimate, Valid: true}), nil
}

func (t *TicketService) CreateTicket(ctx context.Context, userID int64, form CreateTicketForm) ([]ticket.TicketDetailProps, error) {
	tx, err := t.db.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Error creating transaction", slog.Any("error", err))
		return nil, err
	}
	defer tx.Rollback()

	q := t.db.Queries.WithTx(tx)
	tticket, err := q.CreateTicket(ctx, dbgen.CreateTicketParams{
		Name:        form.TicketName,
		Description: form.TicketDescription,
		RoomID:      form.RoomID,
	})
	if err != nil {
		tx.Rollback()
		slog.Error("Error creating ticket", slog.Any("error", err))
		return nil, err
	}
	slog.Info("Created ticket", slog.Any("ticket", tticket))

	tickets, err := t.GetTicketsOfRoom(ctx, form.RoomID, userID, q)

	if err != nil {
		tx.Rollback()
		slog.Error("Error getting tickets", slog.Any("error", err))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Error committing transaction", slog.Any("error", err))
		return nil, err
	}

	t.webSocketService.SendNewTicket(tickets[0])

	return tickets, nil
}

func (t *TicketService) GetTicketsOfRoom(ctx context.Context, roomID int64, userID int64, q *dbgen.Queries) ([]ticket.TicketDetailProps, error) {
	queries := q
	if q == nil {
		queries = t.db.Queries
	}

	tickets, err := queries.GetTicketsOfRoom(ctx, dbgen.GetTicketsOfRoomParams{
		UserID: userID,
		RoomID: roomID,
	})
	if err != nil {
		slog.Error("Error getting tickets", slog.Any("error", err))
		return nil, err
	}

	details := make([]ticket.TicketDetailProps, len(tickets))
	for i, t := range tickets {
		details[i] = ticket.TicketDetailProps{
			ID:              t.ID,
			RoomID:          t.RoomID,
			Name:            t.Name,
			Description:     t.Description,
			HasEstimate:     t.HasEstimate,
			IsClosed:        false,
			AnsweredBy:      "TBD",
			UserEstimate:    prettyPrintEstimate(t.UserEstimate),
			AverageEstimate: prettyPrintEstimate(sql.NullInt64{Int64: int64(t.AvgEstimate.Float64), Valid: t.AvgEstimate.Valid}),
			EstimatedBy:     fmt.Sprintf("%d/%d", t.UsersEstimated, t.TotalUsersInRoom),
		}
	}

	return details, nil
}

func NewTicketService(db *database.Database) *TicketService {
	ticketService := &TicketService{
		db: db,
	}
	ticketService.webSocketService = NewWebSocketService(ticketService)
	return ticketService
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
