package server

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/cmd/web/components/room"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
	"github.com/markojerkic/spring-planing/internal/service"
)

type RoomRouter struct {
	roomService *service.RoomService
	group       *echo.Group
}

func (r *RoomRouter) createRoomHandler(ctx echo.Context) error {
	user := ctx.Get("user").(dbgen.User)
	name := ctx.FormValue("roomName")

	_, err := r.roomService.CreateRoom(ctx.Request().Context(), user.ID, name)
	if err != nil {
		ctx.Logger().Errorf("Error creating room: %v", err)
		return ctx.String(500, "Error creating room")
	}

	return ctx.Redirect(302, fmt.Sprintf("/room/%s", name))
}

func newRoomRouter(db *database.Database, group *echo.Group) *RoomRouter {
	r := &RoomRouter{
		roomService: service.NewRoomService(db),
		group:       group,
	}
	e := r.group
	e.GET("", echo.WrapHandler(templ.Handler(room.CreateRoom())))

	return r
}
