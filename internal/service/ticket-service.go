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
	db *database.Database
}

type CreateTicketForm struct {
	TicketName        string `json:"ticketName" form:"ticketName" validate:"required"`
	TicketDescription string `json:"ticketDescription" form:"ticketDescription" validate:"required"`
	RoomID            int64  `json:"roomID" form:"roomID" validate:"required"`
}

func (t *TicketService) CreateTicket(ctx context.Context, userID int64, form CreateTicketForm) ([]ticket.TicketDetailProps, error) {
	tx, err := t.db.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Error creating transaction", slog.Any("error", err))
		return nil, err
	}
	defer tx.Rollback()

	q := t.db.Queries.WithTx(tx)
	ticket, err := q.CreateTicket(ctx, dbgen.CreateTicketParams{
		Name:        form.TicketName,
		Description: form.TicketDescription,
		RoomID:      form.RoomID,
	})
	if err != nil {
		tx.Rollback()
		slog.Error("Error creating ticket", slog.Any("error", err))
		return nil, err
	}
	slog.Info("Created ticket", slog.Any("ticket", ticket))

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
			ID:           t.ID,
			Name:         t.Name,
			Description:  t.Description,
			HasEstimate:  t.HasEstimate,
			IsClosed:     false,
			AnsweredBy:   "TBD",
			UserEstimate: prettyPrintEstimate(t.UserEstimate),
		}
	}

	return details, nil
}

func NewTicketService(db *database.Database) *TicketService {
	return &TicketService{
		db: db,
	}
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
