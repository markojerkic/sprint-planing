package homepage

import (
	"fmt"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/service"
)

func HomepageHandler(roomService *service.RoomService) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(database.User)

		roomId := c.QueryParam("roomId")

		if roomId != "" {
			return c.Redirect(302, fmt.Sprintf("/room/%s", roomId))
		}

		rooms, err := roomService.GetUsersRooms(c.Request().Context(), int32(user.ID))
		if err != nil {
			c.Logger().Error(err)
			return c.String(500, "Error getting rooms")
		}
		slog.Debug("Rooms", slog.Any("rooms", rooms))
		slog.Debug("User", slog.Any("user", user))

		return RoomList(rooms, user.ID).Render(c.Request().Context(), c.Response().Writer)
	}
}
