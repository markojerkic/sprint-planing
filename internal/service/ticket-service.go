package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
	"gorm.io/gorm"
)

type TicketService struct {
	db               *database.Database
	webSocketService *WebSocketService
}

type CreateTicketForm struct {
	TicketName        string `json:"ticketName" form:"ticketName" validate:"required"`
	TicketDescription string `json:"ticketDescription" form:"ticketDescription" validate:"required"`
	RoomID            int32  `json:"roomID" form:"roomID" validate:"required"`
}

type EstimateTicketForm struct {
	TicketID     int32 `json:"ticketID" form:"ticketID" validate:"required"`
	WeekEstimate int32 `json:"weekEstimate" form:"weekEstimate" default:"0"`
	DayEstimate  int32 `json:"dayEstimate" form:"dayEstimate" default:"0"`
	HourEstimate int32 `json:"hourEstimate" form:"hourEstimate" default:"0"`
}

type HideTicketDto struct {
	TicketID int32 `json:"ticketID" form:"ticketID"`
	IsHidden bool  `json:"isHidden" form:"isHidden"`
}

func (t *TicketService) HideTicket(ctx context.Context, ticketID int32) (*database.Ticket, error) {
	var ticket database.Ticket
	err := t.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&ticket, ticketID).Error; err != nil {
			return err
		}

		ticket.Hidden = !ticket.Hidden
		if err := tx.Save(&ticket).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		slog.Error("Error hiding ticket", slog.Any("error", err))
		return nil, err
	}

	return &ticket, nil

	// ticket, err := t.db.Queries.ToggleTicketHidden(ctx, ticketID)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// t.webSocketService.HideTicket(ticketID, ticket.RoomID, ticket.Hidden)
	// return &ticket, nil
}

func (t *TicketService) EstimateTicket(ctx context.Context, userID int32, form EstimateTicketForm) (string, error) {
	var prettyEstimate string
	err := t.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		estimate := database.Estimate{
			TicketID: uint(form.TicketID),
			Estimate: int(form.WeekEstimate*5*8 + form.DayEstimate*8 + form.HourEstimate),
			UserID:   uint(userID),
		}

		if err := tx.Create(&estimate).Error; err != nil {
			return err
		}

		ticketEstimates := []database.Estimate{}
		if err := tx.Where("ticket_id = ?", form.TicketID).Find(&ticketEstimates).Error; err != nil {
			return err
		}

		prettyEstimate = prettyPrintEstimate(estimate.Estimate)
		return nil
	})
	if err != nil {
		slog.Error("Error estimating ticket", slog.Any("error", err))
		return "", err
	}

	return prettyEstimate, nil
	// tx, err := t.db.DB.BeginTx(ctx, pgx.TxOptions{})
	// if err != nil {
	// 	slog.Error("Error creating transaction", slog.Any("error", err))
	// 	return "", err
	// }
	// defer tx.Rollback(ctx)
	//
	// q := t.db.Queries.WithTx(tx)
	//
	// // Assumes a day is 8 hours, and a week is 5 days
	// estimate, err := q.EstimateTicket(ctx, dbgen.EstimateTicketParams{
	// 	Estimate: form.WeekEstimate*5*8 + form.DayEstimate*8 + form.HourEstimate,
	// 	UserID:   userID,
	// 	TicketID: form.TicketID,
	// })
	// if err != nil {
	// 	tx.Rollback(ctx)
	// 	slog.Error("Error estimating ticket", slog.Any("error", err))
	// 	return "", err
	// }
	//
	// avgEst, err := q.GetTicketAverageEstimation(ctx, form.TicketID)
	// if err != nil {
	// 	tx.Rollback(ctx)
	// 	slog.Error("Error getting ticket average estimation", slog.Any("error", err))
	// 	return "", err
	// }
	//
	// if err := tx.Commit(ctx); err != nil {
	// 	tx.Rollback(ctx)
	// 	slog.Error("Error committing transaction", slog.Any("error", err))
	// 	return "", err
	// }
	//
	// averageEstimate := avgEst.AvgEstimate.Float64
	// t.webSocketService.UpdateEstimate(avgEst.ID,
	// 	avgEst.RoomID,
	// 	prettyPrintEstimate(pgtype.Int4{Int32: int32(averageEstimate), Valid: true}),
	// 	prettyPrintEstimatef(avgEst.MedianEstimate),
	// 	prettyPrintStd(avgEst.StdDevEstimate),
	// 	fmt.Sprintf("%d/%d", avgEst.UsersEstimated, avgEst.TotalUsersInRoom))
	//
	// return prettyPrintEstimate(pgtype.Int4{Int32: estimate.Estimate, Valid: true}), nil
}

func (t *TicketService) CloseTicket(ctx context.Context, ticketID int32) (*database.Ticket, error) {
	var ticket database.Ticket
	err := t.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find the ticket with preloaded estimates
		if err := tx.Preload("Estimates").First(&ticket, ticketID).Error; err != nil {
			return err
		}
		now := time.Now()
		ticket.ClosedAt = &now

		if err := tx.Save(&ticket).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ticket, nil
	//
	// ticket, err := t.db.Queries.CloseTicket(ctx, ticketID)
	// if err != nil {
	// 	slog.Error("Error closing ticket", slog.Any("error", err))
	// 	return nil, err
	// }
	// avgEst, err := t.db.Queries.GetTicketAverageEstimation(ctx, ticketID)
	//
	// averageEstimate := avgEst.AvgEstimate.Float64
	// t.webSocketService.CloseTicket(ticketID,
	// 	avgEst.RoomID,
	// 	prettyPrintEstimate(pgtype.Int4{Int32: int32(averageEstimate), Valid: true}),
	// 	prettyPrintEstimatef(avgEst.MedianEstimate),
	// 	prettyPrintStd(avgEst.StdDevEstimate),
	// 	fmt.Sprintf("%d/%d", avgEst.UsersEstimated, avgEst.TotalUsersInRoom))
	//
	// return &ticket, nil
}

func (t *TicketService) GetTicketEstimates(ctx context.Context, ticketID int32) ([]string, error) {
	estimates := make([]string, 0)

	err := t.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dbEstimates []database.Estimate
		if err := tx.Where("ticket_id = ?", ticketID).Find(&dbEstimates).Error; err != nil {
			return err
		}

		for _, e := range dbEstimates {
			estimates = append(estimates, prettyPrintEstimate(e.Estimate))
		}

		return nil
	})
	if err != nil {
		slog.Error("Error getting ticket estimates", slog.Any("error", err))
		return nil, err
	}

	return estimates, nil

	// estimates, err := t.db.Queries.GetTicketEstimates(ctx, ticketID)
	// if err != nil {
	// 	slog.Error("Error getting ticket estimates", slog.Any("error", err))
	// 	return nil, err
	// }
	//
	// prettyEstimates := make([]string, len(estimates))
	// for i, e := range estimates {
	// 	prettyEstimates[i] = prettyPrintEstimate(pgtype.Int4{Int32: e, Valid: true})
	// }
	//
	// return prettyEstimates, nil
}

func (t *TicketService) CreateTicket(ctx context.Context, userID int32, form CreateTicketForm) (*database.Ticket, error) {
	var ticket database.Ticket

	err := t.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ticket = database.Ticket{
			Name:        form.TicketName,
			Description: form.TicketDescription,
			RoomID:      uint(form.RoomID),
			CreatedBy:   uint(userID),
		}

		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		slog.Error("Error creating ticket", slog.Any("error", err))
		return nil, err
	}

	return &ticket, nil

	// tx, err := t.db.DB.BeginTx(ctx, pgx.TxOptions{})
	// if err != nil {
	// 	slog.Error("Error creating transaction", slog.Any("error", err))
	// 	return nil, err
	// }
	// defer tx.Rollback(ctx)
	//
	// q := t.db.Queries.WithTx(tx)
	// tticket, err := q.CreateTicket(ctx, dbgen.CreateTicketParams{
	// 	Name:        form.TicketName,
	// 	Description: form.TicketDescription,
	// 	RoomID:      form.RoomID,
	// })
	// if err != nil {
	// 	tx.Rollback(ctx)
	// 	slog.Error("Error creating ticket", slog.Any("error", err))
	// 	return nil, err
	// }
	// slog.Info("Created ticket", slog.Any("ticket", tticket))
	//
	// tickets, err := t.GetTicketsOfRoom(ctx, form.RoomID, userID, q)
	//
	// if err != nil {
	// 	tx.Rollback(ctx)
	// 	slog.Error("Error getting tickets", slog.Any("error", err))
	// 	return nil, err
	// }
	//
	// if err := tx.Commit(ctx); err != nil {
	// 	slog.Error("Error committing transaction", slog.Any("error", err))
	// 	return nil, err
	// }
	//
	// t.webSocketService.SendNewTicket(tickets[0])
	//
	// return tickets, nil
}

func (t *TicketService) GetTicketsOfRoom(ctx context.Context, roomID int32, userID int32, q *dbgen.Queries) ([]database.Ticket, error) {
	var tickets []database.Ticket

	err := t.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("room_id = ?", roomID).Find(&tickets).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		slog.Error("Error getting tickets", slog.Any("error", err))
		return nil, err
	}

	return tickets, nil

	// queries := q
	// if q == nil {
	// 	queries = t.db.Queries
	// }
	//
	// tickets, err := queries.GetTicketsOfRoom(ctx, dbgen.GetTicketsOfRoomParams{
	// 	UserID: userID,
	// 	RoomID: roomID,
	// })
	// if err != nil {
	// 	slog.Error("Error getting tickets", slog.Any("error", err))
	// 	return nil, err
	// }
	//
	// details := make([]ticket.TicketDetailProps, len(tickets))
	// for i, t := range tickets {
	// 	details[i] = ticket.TicketDetailProps{
	// 		ID:              t.ID,
	// 		RoomID:          t.RoomID,
	// 		Name:            t.Name,
	// 		Description:     t.Description,
	// 		HasEstimate:     false, // t.HasEstimate,
	// 		IsClosed:        t.ClosedAt.Valid,
	// 		IsHidden:        t.Hidden,
	// 		AnsweredBy:      "0",
	// 		UserEstimate:    prettyPrintEstimate(t.UserEstimate),
	// 		AverageEstimate: prettyPrintEstimate(pgtype.Int4{Int32: int32(t.AvgEstimate.Float64), Valid: t.AvgEstimate.Valid}),
	// 		MedianEstimate:  prettyPrintEstimatef(t.MedianEstimate),
	// 		StdEstimate:     prettyPrintStd(t.StdDevEstimate),
	// 		EstimatedBy:     fmt.Sprintf("%d/%d", t.UsersEstimated, t.TotalUsersInRoom),
	// 	}
	// }
	//
	// return details, nil
}

func NewTicketService(db *database.Database) *TicketService {
	ticketService := &TicketService{
		db: db,
	}
	ticketService.webSocketService = NewWebSocketService(ticketService)
	return ticketService
}

func prettyPrintStd(stdf pgtype.Float8) string {
	if !stdf.Valid {
		return "No estimate"
	}
	estimate := stdf.Float64

	weeks := int32(estimate) / 40
	days := (int32(estimate) % 40) / 8
	hours := float64(int32(estimate)%8) + (estimate - float64(int32(estimate)))

	return fmt.Sprintf("%dw %dd %.2fh", weeks, days, hours)
}

func prettyPrintEstimatef(nEstimate pgtype.Float8) string {
	if !nEstimate.Valid {
		return "No estimate"
	}
	estimate := int32(nEstimate.Float64)

	weeks := estimate / 40
	days := (estimate % 40) / 8
	hours := estimate % 8

	return fmt.Sprintf("%dw %dd %dh", weeks, days, hours)
}

func prettyPrintEstimate(estimate int) string {

	weeks := estimate / 40
	days := (estimate % 40) / 8
	hours := estimate % 8

	return fmt.Sprintf("%dw %dd %dh", weeks, days, hours)
}
