package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/markojerkic/spring-planing/internal/database"
	"gorm.io/gorm"
)

var ticketQuery = `
    SELECT t.*,
           AVG(e.estimate)                                         AS average_estimate,
           PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY e.estimate) AS median_estimate,
           STDDEV(e.estimate)                                      AS std_dev_estimate,
           COUNT(DISTINCT e.id)                                    AS estimate_count,
           COUNT(DISTINCT room_users.user_id)                      AS user_count,
           users_estimate.estimate                                 AS users_estimate
    FROM tickets t
             LEFT JOIN estimates e ON t.id = e.ticket_id
             LEFT JOIN estimates users_estimate ON t.id = users_estimate.ticket_id AND users_estimate.user_id = ?
             LEFT JOIN room_users ON t.room_id = room_users.room_id
    WHERE t.room_id = ?
      AND t.deleted_at IS NULL
    GROUP BY t.id, t.created_at, users_estimate.estimate
    ORDER BY t.created_at DESC;`

type TicketService struct {
	db               *database.Database
	webSocketService *WebSocketService
}

type CreateTicketForm struct {
	TicketName        string `json:"ticketName" form:"ticketName" validate:"required"`
	TicketDescription string `json:"ticketDescription" form:"ticketDescription" validate:"required"`
	RoomID            uint   `json:"roomID" form:"roomID" validate:"required"`
	JiraKey           string `json:"jiraKey" form:"jiraKey"`
}

type EstimateTicketForm struct {
	TicketID     uint  `json:"ticketID" form:"ticketID" validate:"required"`
	RoomID       uint  `json:"roomID" form:"roomID" validate:"required"`
	WeekEstimate int32 `json:"weekEstimate" form:"weekEstimate" default:"0"`
	DayEstimate  int32 `json:"dayEstimate" form:"dayEstimate" default:"0"`
	HourEstimate int32 `json:"hourEstimate" form:"hourEstimate" default:"0"`
}

type HideTicketDto struct {
	TicketID uint `json:"ticketID" form:"ticketID"`
	IsHidden bool `json:"isHidden" form:"isHidden"`
}

func (t *TicketService) HideTicket(ctx context.Context, ticketID uint) (*database.Ticket, error) {
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

	t.webSocketService.HideTicket(ticketID, ticket.RoomID, ticket.Hidden)

	return &ticket, nil
}

func (t *TicketService) EstimateTicket(ctx context.Context, userID uint, form EstimateTicketForm) (string, error) {
	var prettyEstimate string
	err := t.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		estimate := database.Estimate{
			TicketID: uint(form.TicketID),
			Estimate: int(form.WeekEstimate*5*8 + form.DayEstimate*8 + form.HourEstimate),
			UserID:   uint(userID),
		}

		if err := tx.Create(&estimate).Error; err != nil {
			slog.Error("Error creating estimate", slog.Any("error", err))
			return err
		}

		updatedTicket, err := t.GetTicket(ctx, tx, userID, &form.RoomID, form.TicketID)
		if err != nil {
			slog.Error("Error getting ticket", slog.Any("error", err))
			return err
		}

		var usersInRoom int
		if err := tx.Raw("SELECT COUNT(*) FROM room_users WHERE room_id = ?",
			updatedTicket.RoomID).Scan(&usersInRoom).Error; err != nil {
			slog.Error("Error getting users in room", slog.Any("error", err))
			return err
		}
		slog.Debug("Estimate ticket", slog.Any("users", usersInRoom))

		prettyEstimate = prettyPrintEstimate(estimate.Estimate)
		t.webSocketService.UpdateEstimate(updatedTicket.ID,
			updatedTicket.JiraKey,
			updatedTicket.RoomID,
			prettyPrintEstimate(int(updatedTicket.AverageEstimate)),
			prettyPrintEstimate(int(updatedTicket.MedianEstimate)),
			fmt.Sprintf("%.2fh", updatedTicket.StdDevEstimate),
			fmt.Sprintf("%d/%d", updatedTicket.EstimateCount, updatedTicket.UserCount),
		)
		return nil
	})
	if err != nil {
		slog.Error("Error estimating ticket", slog.Any("error", err))
		return "", err
	}

	return prettyEstimate, nil
}

func (t *TicketService) CloseTicket(ctx context.Context, ticketID uint, userID uint) (*database.Ticket, error) {
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

	ticketWithStats, err := t.GetTicket(ctx, t.db.DB, userID, nil, ticketID)
	jiraKey := ticket.JiraKey

	ticketProps := ticketWithStats.ToDetailProp(true)

	t.webSocketService.CloseTicket(ticketID,
		jiraKey,
		ticket.RoomID,
		ticketProps.AverageEstimate,
		ticketProps.MedianEstimate,
		ticketProps.StdEstimate,
		ticketProps.EstimatedBy,
	)

	return &ticket, nil
}

func (t *TicketService) GetTicketEstimates(ctx context.Context, ticketID int32) ([]string, error) {
	estimates := make([]string, 0)

	err := t.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dbEstimates []database.Estimate
		if err := tx.Where("ticket_id = ?", ticketID).Order("estimate ASC").Find(&dbEstimates).Error; err != nil {
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
}

func (t *TicketService) GetTicket(ctx context.Context, db *gorm.DB, userID uint, roomID *uint, ticketID uint) (*database.TicketWithEstimateStatistics, error) {
	var tickets []database.TicketWithEstimateStatistics
	var foundRoomId *uint

	if roomID != nil {
		foundRoomId = roomID
	} else {
		var ticket database.Ticket
		if err := db.WithContext(ctx).Find(&ticket, ticketID).Error; err != nil {
			slog.Error("Error getting ticket for statistics", slog.Any("id", ticketID), slog.Any("error", err))
			return nil, err
		}
		foundRoomId = &ticket.RoomID
	}

	if err := db.WithContext(ctx).
		Raw(ticketQuery, userID, *foundRoomId).
		Scan(&tickets).Error; err != nil {
		slog.Error("Error getting ticket", slog.Int("userID", int(userID)),
			slog.Int("ticketID", int(ticketID)), slog.Any("error", err))
		return nil, err
	}

	for _, ticket := range tickets {
		if ticket.ID == ticketID {
			return &ticket, nil
		}
	}

	slog.Error("Ticket not found",
		slog.Int("ticketID", int(ticketID)),
		slog.Int("roomID", int(*foundRoomId)),
		slog.Any("tickets", tickets))

	return nil, fmt.Errorf("ticket not found")
}

func (t *TicketService) GetTicketsOfRoom(ctx context.Context, db *gorm.DB, userID uint, roomID uint) ([]database.TicketWithEstimateStatistics, error) {
	var tickets []database.TicketWithEstimateStatistics
	if err := db.WithContext(ctx).
		Raw(ticketQuery, userID, roomID).
		Scan(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func (t *TicketService) CreateTicket(ctx context.Context, userID uint, form CreateTicketForm) ([]database.TicketWithEstimateStatistics, error) {
	var tickets []database.TicketWithEstimateStatistics

	err := t.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ticket := database.Ticket{
			Name:        form.TicketName,
			Description: form.TicketDescription,
			RoomID:      uint(form.RoomID),
			CreatedBy:   uint(userID),
		}

		if form.JiraKey != "" {
			ticket.JiraKey = &form.JiraKey
		}

		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}

		roomTickets, err := t.GetTicketsOfRoom(ctx, tx, userID, form.RoomID)
		if err != nil {
			return err
		}
		tickets = roomTickets

		return nil
	})
	if err != nil {
		slog.Error("Error creating ticket", slog.Any("error", err))
		return nil, err
	}

	savedTicket := tickets[0]
	isOwner := savedTicket.CreatedBy == userID

	t.webSocketService.SendNewTicket(savedTicket.ToDetailProp(isOwner))

	return tickets, nil
}

func NewTicketService(db *database.Database) *TicketService {
	ticketService := &TicketService{
		db: db,
	}
	ticketService.webSocketService = NewWebSocketService(ticketService)
	return ticketService
}

func prettyPrintEstimate(estimate int) string {
	weeks := estimate / 40
	days := (estimate % 40) / 8
	hours := estimate % 8

	return fmt.Sprintf("%dw %dd %dh", weeks, days, hours)
}
