package service

import (
	"bytes"
	"context"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
)

var rooms = make(map[int64]map[*websocket.Conn]bool)
var mutex = sync.Mutex{}

type message struct {
	conn   *websocket.Conn
	data   *[]byte
	roomID int64
}

var buffer = make(chan message, 100)

type WebSocketService struct {
	TicketService *TicketService
}

func writePump() {
	for {
		select {
		case msg := <-buffer:
			if err := msg.conn.WriteMessage(websocket.TextMessage, *msg.data); err != nil {
				log.Printf("Error writing message: %v", err)
				mutex.Lock()
				delete(rooms[msg.roomID], msg.conn)
				mutex.Unlock()
			}

		}
	}
}

func (w *WebSocketService) CloseTicket(ticketID int64, roomID int64, averageEstimate string, estimatedBy string) {
	removedTicketForm := new(bytes.Buffer)
	if err := ticket.ClosedEstimation(ticketID, averageEstimate, estimatedBy).
		Render(context.Background(), removedTicketForm); err != nil {
		log.Printf("Error rendering ticket thumbnail: %v", err)
		return
	}

	log.Printf("Closing ticket and sending render %d for roomID %d", ticketID, roomID)
	bytes := removedTicketForm.Bytes()
	mutex.Lock()
	conns := rooms[roomID]
	mutex.Unlock()
	for conn := range conns {
		buffer <- message{conn: conn, data: &bytes, roomID: roomID}
	}
}

func (w *WebSocketService) UpdateEstimate(ticketID int64, roomID int64, averageEstimate string, estimatedBy string) {
	renderedTicket := new(bytes.Buffer)
	if err := ticket.UpdatedEstimationDetail(ticketID, averageEstimate, estimatedBy).
		Render(context.Background(), renderedTicket); err != nil {
		log.Printf("Error rendering ticket thumbnail: %v", err)
		return
	}

	bytes := renderedTicket.Bytes()
	mutex.Lock()
	conns := rooms[roomID]
	mutex.Unlock()
	for conn := range conns {
		buffer <- message{conn: conn, data: &bytes, roomID: roomID}
	}
}

func (w *WebSocketService) SendNewTicket(tticket ticket.TicketDetailProps) {
	renderedTicket := new(bytes.Buffer)
	if err := ticket.CreatedTicketUpdate(tticket).
		Render(context.Background(), renderedTicket); err != nil {
		log.Printf("Error rendering ticket thumbnail: %v", err)
		return
	}

	bytes := renderedTicket.Bytes()
	mutex.Lock()
	conns := rooms[tticket.RoomID]
	mutex.Unlock()
	for conn := range conns {
		buffer <- message{conn: conn, data: &bytes}
	}
}

func (w *WebSocketService) Register(conn *websocket.Conn, roomID int64) {
	mutex.Lock()
	if _, ok := rooms[roomID]; !ok {
		rooms[roomID] = make(map[*websocket.Conn]bool)
	}
	rooms[roomID][conn] = true
	mutex.Unlock()
}

func NewWebSocketService(ticketService *TicketService) *WebSocketService {
	service := &WebSocketService{
		TicketService: ticketService,
	}

	// Start N writePump goroutines
	for range 30 {
		go writePump()
	}

	return service
}
