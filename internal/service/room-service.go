package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/markojerkic/spring-planing/internal/database"
	"gorm.io/gorm"
)

type RoomService struct {
	db                *database.Database
	roomTicketService *RoomTicketService
}

type RoomTicket struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	IsHidden bool   `json:"isHidden"`
}

func (r *RoomService) GetTotalEstimateOfRoom(ctx context.Context, roomID uint) (string, error) {
	var totalEstimatedHours int
	if err := r.db.DB.Raw(`
		WITH
		  avg_estimates AS (
			SELECT
			  PERCENTILE_CONT(0.5) WITHIN GROUP (
				ORDER BY
				  estimate
			  ) AS median_estimate,
			  room_id
			FROM
			  estimates
			  JOIN tickets ON tickets.id = estimates.ticket_id
			  AND tickets.closed_at IS NOT NULL
			WHERE estimates.user_id IS NOT NULL
			GROUP BY
			  ticket_id,
			  room_id
		  )
		SELECT
		  COALESCE(SUM(median_estimate), 0) AS total_avg_estimate
		FROM
		  avg_estimates
		WHERE
		  room_id = ?;
		`, roomID).
		First(&totalEstimatedHours).
		Error; err != nil {
		return "", err
	}

	return prettyPrintEstimate(totalEstimatedHours), nil
}

func (r *RoomService) GetTicketList(ctx context.Context, roomID uint) ([]RoomTicket, error) {
	tickets := make([]database.Ticket, 0)
	if err := r.db.DB.Model(&database.Ticket{}).
		Select("id, name, hidden").
		Where("room_id = ?", roomID).
		Order("id desc").
		Find(&tickets).Error; err != nil {
		return nil, errors.Join(err, errors.New("Error getting tickets"))
	}
	ticketDtos := make([]RoomTicket, len(tickets))
	for i, t := range tickets {
		ticketDtos[i] = RoomTicket{
			ID:       t.ID,
			Name:     t.Name,
			IsHidden: t.Hidden,
		}
	}

	return ticketDtos, nil
}

func (r *RoomService) GetIsOwner(ctx context.Context, roomID uint, userID uint) bool {
	var room database.Room
	if err := r.db.DB.First(&room, roomID).Error; err != nil {
		slog.Error("Error reading room", slog.Any("err", err))
		return false
	}

	return userID == room.CreatedBy
}

func (r *RoomService) DeleteRoom(ctx context.Context, roomID uint, userID uint) ([]database.Room, error) {
	var rooms []database.Room
	err := r.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var room database.Room
		if err := tx.First(&room, roomID).Error; err != nil {
			return err
		}

		if room.CreatedBy != userID {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Delete(&room).Error; err != nil {
			return err
		}

		if err := tx.Preload("Users").Joins("JOIN room_users ON room_users.room_id = rooms.id").
			Where("room_users.user_id = ?", userID).
			Find(&rooms).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *RoomService) GetUsersRooms(ctx context.Context, userID int32) ([]database.Room, error) {
	var rooms []database.Room
	err := r.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Order("created_at desc").
			Preload("Users").
			Joins("JOIN room_users ON room_users.room_id = rooms.id").
			Where("room_users.user_id = ?", userID).
			Find(&rooms).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *RoomService) CreateRoom(ctx context.Context, userID uint, roomName string, allowLLM bool) (*database.Room, error) {
	var room database.Room
	err := r.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user database.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		room = database.Room{
			CreatedBy:          userID,
			AllowLLMEstimation: allowLLM,
			Name:               roomName,
			Users:              []database.User{user},
		}

		if err := tx.Create(&room).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *RoomService) GetRoom(ctx context.Context, roomID uint, userID uint) (*database.Room, error) {
	var room database.Room
	err := r.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user database.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		if err := tx.Preload("Users").
			First(&room, roomID).Error; err != nil {
			return err
		}

		ticketsWithStatistics, err := r.roomTicketService.GetTicketsOfRoom(ctx, tx, userID, room.ID)
		if err != nil {
			return err
		}
		room.TicketsWithStatistics = ticketsWithStatistics

		// Check if user is in the room
		// If not, add user to the room
		var found bool
		for _, u := range room.Users {
			if u.ID == user.ID {
				found = true
				break
			}
		}
		if !found {
			room.Users = append(room.Users, user)
			if err := tx.Save(&room).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &room, nil
}

func NewRoomService(db *database.Database, roomTicketService *RoomTicketService) *RoomService {
	return &RoomService{
		db:                db,
		roomTicketService: roomTicketService,
	}
}
