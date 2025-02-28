package room

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database"
)

type RoomRouter struct {
	db    database.Service
	group *echo.Group
}

func NewRoomRouter(db database.Service, group *echo.Group) *RoomRouter {
	r := &RoomRouter{
		db:    db,
		group: group,
	}
	e := r.group
	e.GET("/", echo.WrapHandler(templ.Handler(CreateRoom())))
	e.POST("/", r.createRoom)

	return r
}

func (r *RoomRouter) createRoom(ctx echo.Context) error {
	return ctx.String(200, "Room created")
}
