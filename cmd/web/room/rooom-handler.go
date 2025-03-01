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

	return roomDetail(room, tickets).Render(c.Request().Context(), c.Response().Writer)
}

func (r *RoomRouter) roomDetailData(c echo.Context, roomID int64) (dbgen.GetRoomDetailsRow, []dbgen.Ticket, error) {
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
		return dbgen.GetRoomDetailsRow{}, []dbgen.Ticket{}, err
	}

	tickets, err := q.GetTicketsOfRoom(c.Request().Context(), roomID)
	if err != nil {
		tx.Rollback()
		c.Logger().Errorf("Error getting room tickets: %v", err)
		return dbgen.GetRoomDetailsRow{}, []dbgen.Ticket{}, err
	}

	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("Error committing transaction: %v", err)
		return dbgen.GetRoomDetailsRow{}, []dbgen.Ticket{}, err
	}

	return room, tickets, nil
}
