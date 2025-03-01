package room

import (
	"bytes"
	"context"
	"log"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
)

type roomsHub struct {
	// Subscribers to room
	mutex sync.Mutex
	rooms map[int64]map[*websocket.Conn]bool
}

var hub = roomsHub{
	rooms: make(map[int64]map[*websocket.Conn]bool),
	mutex: sync.Mutex{},
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (r *RoomRouter) webSocketHandler(c echo.Context) error {
	roomId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(400, "Invalid room roomID")
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	hub.mutex.Lock()
	if _, ok := hub.rooms[roomId]; !ok {
		hub.rooms[roomId] = make(map[*websocket.Conn]bool)
	}
	hub.rooms[roomId][conn] = true
	hub.mutex.Unlock()

	return nil
}

func (r *RoomRouter) sendNewTicket(ticket dbgen.Ticket) {
	renderedTicketThumbnail := new(bytes.Buffer)
	if err := toTopOfListTicketThumbnail(dbgen.GetTicketsOfRoomRow{
		ID:          ticket.ID,
		Name:        ticket.Name,
		Description: ticket.Description,
		CreatedAt:   ticket.CreatedAt,
		HasEstimate: false,
	}).Render(context.Background(), renderedTicketThumbnail); err != nil {
		log.Printf("Error rendering ticket thumbnail: %v", err)
		return
	}

	hub.mutex.Lock()
	for conn := range hub.rooms[ticket.RoomID] {
		if err := conn.WriteMessage(websocket.TextMessage, renderedTicketThumbnail.Bytes()); err != nil {
			log.Printf("Error writing message to websocket: %v", err)
			conn.Close()
			delete(hub.rooms[ticket.RoomID], conn)
		}
	}
	hub.mutex.Unlock()

}
