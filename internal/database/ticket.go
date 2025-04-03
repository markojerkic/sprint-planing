package database

import (
	"fmt"
	"time"

	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
	"gorm.io/gorm"
)

type Ticket struct {
	gorm.Model
	Name        string
	Description string
	JiraKey     *string
	ClosedAt    *time.Time
	Hidden      bool `gorm:"default:false"`
	RoomID      uint
	Room        Room `gorm:"foreignKey:RoomID"`
	CreatedBy   uint
	Estimates   []Estimate
}

type TicketWithEstimateStatistics struct {
	Ticket
	AverageEstimate float64
	MedianEstimate  float64
	StdDevEstimate  float64
	UsersEstimate   *int
	EstimateCount   int
	UserCount       int
}

func (t *TicketWithEstimateStatistics) ToDetailProp(isOwner bool) ticket.TicketDetailProps {
	ticket := &ticket.TicketDetailProps{
		ID:              t.ID,
		JiraKey:         t.JiraKey,
		Name:            t.Name,
		RoomID:          t.RoomID,
		Description:     t.Description,
		EstimatedBy:     fmt.Sprintf("%d/%d", t.EstimateCount, t.UserCount),
		IsClosed:        t.ClosedAt != nil,
		IsHidden:        t.Hidden,
		AverageEstimate: prettyPrintEstimate(t.AverageEstimate),
		MedianEstimate:  prettyPrintEstimate(t.MedianEstimate),
		StdEstimate:     fmt.Sprintf("%.2fh", t.StdDevEstimate),
		HasEstimate:     t.UsersEstimate != nil,
	}

	if !isOwner {
		ticket.JiraKey = nil
	}

	return *ticket
}

func prettyPrintEstimate(estimate float64) string {
	weeks := int(estimate / 40)
	days := int((int(estimate) % 40) / 8)
	hours := int(estimate) % 8

	return fmt.Sprintf("%dw %dd %dh", weeks, days, hours)
}
