package service

import (
	"context"

	"github.com/markojerkic/spring-planing/internal/database"
	"gorm.io/gorm"
)

type RoomService struct {
	db *database.Database
}

func (r *RoomService) GetUsersRooms(ctx context.Context, userID int32) ([]database.Room, error) {
	var rooms []database.Room
	err := r.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Users").Where("created_by = ?", userID).Find(&rooms).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *RoomService) CreateRoom(ctx context.Context, userID uint, roomName string) (*database.Room, error) {
	var room database.Room
	err := r.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user database.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		room = database.Room{
			CreatedBy: userID,
			Name:      roomName,
			Users:     []database.User{user},
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

		if err := tx.Preload("Users").First(&room, roomID).Error; err != nil {
			return err
		}

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

func NewRoomService(db *database.Database) *RoomService {
	return &RoomService{
		db: db,
	}
}
