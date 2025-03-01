package room

import (
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
)

func (r *RoomRouter) createRoom(ctx echo.Context) error {
	user := ctx.Get("user").(dbgen.User)
	name := ctx.FormValue("roomName")

	tx, err := r.db.DB.BeginTx(ctx.Request().Context(), nil)
	if err != nil {
		ctx.Logger().Errorf("Error creating transaction: %v", err)
		return ctx.String(500, "Error creating transaction")
	}
	defer tx.Rollback()

	q := r.db.Queries.WithTx(tx)
	room, err := q.CreateRoom(ctx.Request().Context(), dbgen.CreateRoomParams{
		Name:      name,
		CreatedBy: user.ID,
	})
	if err != nil {
		tx.Rollback()
		ctx.Logger().Errorf("Error creating room: %v", err)
		return ctx.String(500, "Error creating room")
	}

	err = q.AddUserToRoom(ctx.Request().Context(), dbgen.AddUserToRoomParams{
		RoomID: room.ID,
		UserID: user.ID,
	})
	if err != nil {
		tx.Rollback()
		ctx.Logger().Errorf("Error adding user to room: %v", err)
		return ctx.String(500, "Error adding user to room")
	}

	if err := tx.Commit(); err != nil {
		ctx.Logger().Errorf("Error committing transaction: %v", err)
		return ctx.String(500, "Error committing transaction")
	}
	return ctx.Redirect(302, "/")
}
