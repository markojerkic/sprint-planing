package service

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
)

type RoomService struct {
	db *database.Database
}

func (r *RoomService) CreateRoom(ctx context.Context, userID int32, roomName string) (*dbgen.Room, error) {
	tx, err := r.db.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		slog.Error("Error creating transaction", slog.Any("error", err))
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := r.db.Queries.WithTx(tx)
	room, err := q.CreateRoom(ctx, dbgen.CreateRoomParams{
		Name:      roomName,
		CreatedBy: userID,
	})
	if err != nil {
		tx.Rollback(ctx)
		slog.Error("Error creating room", slog.Any("error", err))
		return nil, err
	}

	// Add user to room
	err = q.AddUserToRoom(ctx, dbgen.AddUserToRoomParams{
		RoomID: room.ID,
		UserID: userID,
	})
	if err != nil {
		tx.Rollback(ctx)
		slog.Error("Error adding user to room", slog.Any("error", err))
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("Error committing transaction", slog.Any("error", err))
		return nil, err
	}

	return &room, nil
}

func (r *RoomService) GetRoom(ctx context.Context, roomID int32, userID int32) (*dbgen.GetRoomDetailsRow, error) {
	tx, err := r.db.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		slog.Error("Error creating transaction", slog.Any("error", err))
		return nil, err
	}
	defer tx.Rollback(ctx)
	q := r.db.Queries.WithTx(tx)

	// Add user to room
	err = q.AddUserToRoom(ctx, dbgen.AddUserToRoomParams{
		RoomID: roomID,
		UserID: userID,
	})
	if err != nil {
		tx.Rollback(ctx)
		slog.Error("Error adding user to room", slog.Any("error", err))
		return nil, err
	}

	detail, err := q.GetRoomDetails(ctx, dbgen.GetRoomDetailsParams{
		CreatedBy: userID,
		ID:        roomID,
	})

	if err != nil {
		tx.Rollback(ctx)
		slog.Error("Error getting room details", slog.Any("error", err))
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &detail, err
}

func NewRoomService(db *database.Database) *RoomService {
	return &RoomService{
		db: db,
	}
}
