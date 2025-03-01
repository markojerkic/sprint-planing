package room

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
)

func (r *RoomRouter) roomDetailsHandler(c echo.Context) error {
	user := c.Get("user").(dbgen.User)
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(400, "Invalid room roomID")
	}

	c.Logger().Infof("User: %v", user)
	c.Logger().Infof("Room ID: %v", roomID)

	room, err := r.db.Queries.GetRoomDetails(c.Request().Context(), roomID)
	if err != nil {
		c.Logger().Errorf("Error getting room details: %v", err)
		return c.String(500, "Error getting room details")
	}

	return roomDetail(room).Render(c.Request().Context(), c.Response().Writer)
}
