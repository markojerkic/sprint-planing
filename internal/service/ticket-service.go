package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/markojerkic/spring-planing/internal/database"
	"gorm.io/gorm"
)

type TicketService struct {
	db               *database.Database
	webSocketService *WebSocketService
}

type CreateTicketForm struct {
	TicketName        string `json:"ticketName" form:"ticketName" validate:"required"`
	TicketDescription string `json:"ticketDescription" form:"ticketDescription" validate:"required"`
	RoomID            uint   `json:"roomID" form:"roomID" validate:"required"`
}

type EstimateTicketForm struct {
	TicketID     uint  `json:"ticketID" form:"ticketID" validate:"required"`
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
}

func (t *TicketService) CreateTicket(ctx context.Context, userID uint, form CreateTicketForm) ([]database.Ticket, int, error) {
	var tickets []database.Ticket
	var usersInRoom int

	err := t.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ticket := database.Ticket{
			Name:        form.TicketName,
			Description: form.TicketDescription,
			RoomID:      uint(form.RoomID),
			CreatedBy:   uint(userID),
		}

		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}

		if err := tx.Raw("SELECT COUNT(*) FROM room_users WHERE room_id = ?", form.RoomID).Scan(&usersInRoom).Error; err != nil {
			return err
		}

		if err := tx.Where("room_id = ?", form.RoomID).
			Order("created_at desc").
			Find(&tickets).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		slog.Error("Error creating ticket", slog.Any("error", err))
		return nil, 0, err
	}

	return tickets, usersInRoom, nil
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
