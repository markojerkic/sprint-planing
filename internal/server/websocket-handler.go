package server

import (
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/service"
)

type WebSocketRouter struct {
	service     *service.WebSocketService
	roomService *service.RoomService
	group       *echo.Group
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (r *WebSocketRouter) webSocketHandler(c echo.Context) error {
	user := c.Get("user").(database.User)
	roomId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid room roomID")
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	isOwner := r.roomService.GetIsOwner(c.Request().Context(), uint(roomId), user.ID)
	r.service.Register(conn, uint(roomId), isOwner)

	return nil

}

func newWebsocketRouter(
	webSocketService *service.WebSocketService,
	roomService *service.RoomService,
	group *echo.Group) *WebSocketRouter {
	r := &WebSocketRouter{
		service:     webSocketService,
		roomService: roomService,
		group:       group,
	}
	e := r.group

	e.GET("/:id", r.webSocketHandler)

	return r
}
