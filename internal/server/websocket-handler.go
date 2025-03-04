package server

import (
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/service"
)

type WebSocketRouter struct {
	service *service.WebSocketService
	group   *echo.Group
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (r *WebSocketRouter) webSocketHandler(c echo.Context) error {
	roomId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(400, "Invalid room roomID")
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	r.service.Register(conn, roomId)

	return nil

}

func newWebsocketRouter(db *database.Database, group *echo.Group) *WebSocketRouter {
	ticketService := service.NewTicketService(db)
	r := &WebSocketRouter{
		service: service.NewWebSocketService(ticketService),
		group:   group,
	}
	e := r.group

	e.GET("/:id", r.webSocketHandler)

	return r
}
