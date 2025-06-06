package service

import (
	"context"

	"github.com/markojerkic/spring-planing/internal/database"
	"gorm.io/gorm"
)

type RoomTicketService struct {
	db *database.Database
}

func (r *RoomTicketService) GetTicketsOfRoom(ctx context.Context, db *gorm.DB, userID uint, roomID uint) ([]database.TicketWithEstimateStatistics, error) {
	var tickets []database.TicketWithEstimateStatistics
	if err := db.WithContext(ctx).
		Raw(ticketQuery, userID, roomID).
		Scan(&tickets).Error; err != nil {
		return nil, err
	}
	return tickets, nil
}

func NewRoomTicketService(db *database.Database) *RoomTicketService {
	return &RoomTicketService{db: db}
}
