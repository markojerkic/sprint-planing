package room

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database"
)

type RoomRouter struct {
	db    *database.Database
	group *echo.Group
}

func NewRoomRouter(db *database.Database, group *echo.Group) *RoomRouter {
	r := &RoomRouter{
		db:    db,
		group: group,
	}
	e := r.group
	e.GET("", echo.WrapHandler(templ.Handler(CreateRoom())))
	e.GET("/:id", r.roomDetailsHandler)
	e.POST("", r.createRoom)
	e.POST("/ticket", r.createTicketHandler)

	return r
}
