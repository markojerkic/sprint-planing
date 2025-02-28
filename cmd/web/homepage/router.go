package homepage

import (
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
)

func HomepageHandler(db *dbgen.Queries) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*dbgen.User)

		rooms, err := db.GetMyRooms(c.Request().Context(), dbgen.GetMyRoomsParams{
			ID: user.ID,
		})
		if err != nil {
			c.Logger().Error(err)
			return c.String(500, "Error getting rooms")
		}

		return c.JSON(200, rooms)
	}
}
