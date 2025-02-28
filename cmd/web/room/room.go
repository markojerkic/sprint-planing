package room

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
)

type RoomRouter struct {
	db    *dbgen.Queries
	group *echo.Group
}

func NewRoomRouter(db *dbgen.Queries, group *echo.Group) *RoomRouter {
	r := &RoomRouter{
		db:    db,
		group: group,
	}
	e := r.group
	e.GET("", echo.WrapHandler(templ.Handler(CreateRoom())))
	e.POST("", r.createRoom)

	return r
}

func (r *RoomRouter) createRoom(ctx echo.Context) error {
	name := ctx.FormValue("name")
	room, err := r.db.CreateRoom(ctx.Request().Context(), dbgen.CreateRoomParams{
		Name: name,
	})
	if err != nil {
		ctx.Logger().Error(err)
		return ctx.String(500, "Error creating room")
	}

	return ctx.Redirect(302, fmt.Sprintf("/room/%d", room.ID))
}
