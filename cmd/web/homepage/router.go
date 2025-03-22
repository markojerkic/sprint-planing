package homepage

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
	"github.com/markojerkic/spring-planing/internal/service"
)

func HomepageHandler(roomService *service.RoomService) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(dbgen.User)

		roomId := c.QueryParam("roomId")

		if roomId != "" {
			return c.Redirect(302, fmt.Sprintf("/room/%s", roomId))
		}

		rooms, err := roomService.GetUsersRooms(c.Request().Context(), user.ID)
		if err != nil {
			c.Logger().Error(err)
			return c.String(500, "Error getting rooms")
		}
		log.Printf("User: %v", user)
		log.Printf("Rooms: %v", rooms)

		return c.JSON(200, rooms)
		// return RoomList(rooms).Render(c.Request().Context(), c.Response().Writer)
	}
}
