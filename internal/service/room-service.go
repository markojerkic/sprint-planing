package service

import (
	"context"
	"log/slog"

	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
)

type RoomService struct {
	db *database.Database
}

func (r *RoomService) CreateRoom(ctx context.Context, userID int64, roomName string) (*dbgen.Room, error) {
	tx, err := r.db.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Error creating transaction", slog.Any("error", err))
		return nil, err
	}
	defer tx.Rollback()

	q := r.db.Queries.WithTx(tx)
	room, err := q.CreateRoom(ctx, dbgen.CreateRoomParams{
		Name:      roomName,
		CreatedBy: userID,
	})
	if err != nil {
		tx.Rollback()
		slog.Error("Error creating room", slog.Any("error", err))
		return nil, err
	}

	// Add user to room
	err = q.AddUserToRoom(ctx, dbgen.AddUserToRoomParams{
		RoomID: room.ID,
		UserID: userID,
	})
	if err != nil {
		tx.Rollback()
		slog.Error("Error adding user to room", slog.Any("error", err))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Error committing transaction", slog.Any("error", err))
		return nil, err
	}

	return &room, nil
}

func (r *RoomService) GetRoom(ctx context.Context, roomID int64, userID int64) (*dbgen.GetRoomDetailsRow, error) {
	tx, err := r.db.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Error creating transaction", slog.Any("error", err))
		return nil, err
	}
	defer tx.Rollback()
	q := r.db.Queries.WithTx(tx)

	// Add user to room
	err = q.AddUserToRoom(ctx, dbgen.AddUserToRoomParams{
		RoomID: roomID,
		UserID: userID,
	})
	if err != nil {
		tx.Rollback()
		slog.Error("Error adding user to room", slog.Any("error", err))
		return nil, err
	}

	detail, err := q.GetRoomDetails(ctx, dbgen.GetRoomDetailsParams{
		UserID: userID,
		ID:     roomID,
	})

	if err != nil {
		tx.Rollback()
		slog.Error("Error getting room details", slog.Any("error", err))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &detail, err
}

func NewRoomService(db *database.Database) *RoomService {
	return &RoomService{
		db: db,
	}
}
