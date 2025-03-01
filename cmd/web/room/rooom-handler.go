package room

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
)

func (r *RoomRouter) roomDetailsHandler(c echo.Context) error {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(400, "Invalid room roomID")
	}

	room, tickets, err := r.roomDetailData(c, roomID)
	if err != nil {
		return c.String(500, "Error getting room details")
	}

	if err := r.addUserToRoomIfNotAlreadyThere(c, c.Get("user").(dbgen.User), roomID); err != nil {
		return c.String(500, "Error adding user to room")
	}

	return roomDetail(room, tickets).Render(c.Request().Context(), c.Response().Writer)
}

func (r *RoomRouter) roomDetailData(c echo.Context, roomID int64) (dbgen.GetRoomDetailsRow, []dbgen.GetTicketsOfRoomRow, error) {
	tx, err := r.db.DB.BeginTx(c.Request().Context(), nil)
	q := r.db.Queries.WithTx(tx)
	defer tx.Rollback()

	user := c.Get("user").(dbgen.User)

	room, err := q.GetRoomDetails(c.Request().Context(), dbgen.GetRoomDetailsParams{
		UserID: user.ID,
		ID:     roomID,
	})

	if err != nil {
		tx.Rollback()
		c.Logger().Errorf("Error getting room details: %v", err)
		return dbgen.GetRoomDetailsRow{}, []dbgen.GetTicketsOfRoomRow{}, err
	}

	tickets, err := q.GetTicketsOfRoom(c.Request().Context(), dbgen.GetTicketsOfRoomParams{
		RoomID: roomID,
		UserID: user.ID,
	})

	if err != nil {
		tx.Rollback()
		c.Logger().Errorf("Error getting room tickets: %v", err)
		return dbgen.GetRoomDetailsRow{}, []dbgen.GetTicketsOfRoomRow{}, err
	}

	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("Error committing transaction: %v", err)
		return dbgen.GetRoomDetailsRow{}, []dbgen.GetTicketsOfRoomRow{}, err
	}

	return room, tickets, nil
}

func (r *RoomRouter) addUserToRoomIfNotAlreadyThere(c echo.Context, user dbgen.User, roomID int64) error {
	return r.db.Queries.AddUserToRoom(c.Request().Context(), dbgen.AddUserToRoomParams{
		UserID: user.ID,
		RoomID: roomID,
	})
}
