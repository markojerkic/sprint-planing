package service

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
)

var rooms = make(map[uint]map[*websocket.Conn]bool)
var mutex = sync.RWMutex{}

type message struct {
	conn   *websocket.Conn
	data   *[]byte
	roomID uint
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
				removeConnection(msg.conn, msg.roomID)
			}
		}
	}
}

// removeConnection removes a connection from a room and cleans up empty rooms
func removeConnection(conn *websocket.Conn, roomID uint) {
	mutex.Lock()
	defer mutex.Unlock()

	// Check if the room exists
	if conns, ok := rooms[roomID]; ok {
		// Remove the connection
		delete(conns, conn)

		// If the room is empty, remove it
		if len(conns) == 0 {
			delete(rooms, roomID)
			log.Printf("Room %d is empty and has been removed", roomID)
		}
	}
}

// readPump reads from the websocket connection to detect disconnects
func (w *WebSocketService) readPump(conn *websocket.Conn, roomID uint) {
	defer func() {
		conn.Close()
		removeConnection(conn, roomID)
		log.Printf("Connection closed for room %d", roomID)
	}()

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Read messages from the websocket
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}
	}
}

// CleanupInactiveConnections periodically checks and removes inactive connections
func (w *WebSocketService) CleanupInactiveConnections() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		mutex.Lock()
		// Log current state
		slog.Debug("Checking for inactive connections. Current rooms", slog.Int("rooms", len(rooms)))

		// For each room, ping each connection
		for roomID, conns := range rooms {
			for conn := range conns {
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
					log.Printf("Failed to ping client in room %d: %v", roomID, err)
					delete(conns, conn)
					conn.Close()
				}
			}

			// If room is empty after checking connections, remove it
			if len(conns) == 0 {
				delete(rooms, roomID)
				log.Printf("Room %d is now empty and has been removed", roomID)
			}
		}
		mutex.Unlock()
	}
}

func (w *WebSocketService) HideTicketsOfRoom(roomID uint, isHidden bool) {
	dto := HideTicketDto{
		IsHidden: isHidden,
	}
	jsonDto, err := json.Marshal(dto)
	if err != nil {
		log.Printf("Error marshalling dto: %v", err)
		return
	}

	mutex.RLock()
	conns := rooms[roomID]
	mutex.RUnlock()
	for conn := range conns {
		buffer <- message{conn: conn, data: &jsonDto, roomID: roomID}
	}
}

func (w *WebSocketService) HideTicket(ticketID uint, roomID uint, isHidden bool) {
	dto := HideTicketDto{
		TicketID: ticketID,
		IsHidden: isHidden,
	}
	jsonDto, err := json.Marshal(dto)
	if err != nil {
		log.Printf("Error marshalling dto: %v", err)
		return
	}

	mutex.RLock()
	conns := rooms[roomID]
	mutex.RUnlock()
	for conn := range conns {
		buffer <- message{conn: conn, data: &jsonDto, roomID: roomID}
	}

}

func (w *WebSocketService) CloseTicket(tticket ticket.TicketDetailProps) {
	renderedTicket := new(bytes.Buffer)
	if err := ticket.ClosedTicketUpdate(tticket, false).
		Render(context.Background(), renderedTicket); err != nil {
		log.Printf("Error rendering ticket thumbnail: %v", err)
		return
	}
	log.Printf("Closing ticket and sending render %d for roomID %d", tticket.ID, tticket.RoomID)

	bytes := renderedTicket.Bytes()

	mutex.RLock()
	conns := rooms[tticket.RoomID]
	mutex.RUnlock()
	for conn := range conns {
		buffer <- message{conn: conn, data: &bytes, roomID: tticket.RoomID}
	}
}

func (w *WebSocketService) UpdateEstimate(ticketID uint,
	jiraKey *string,
	roomID uint,
	averageEstimate string,
	medianEstimate string,
	stdEstimate string,
	estimatedBy string) {
	renderedTicket := new(bytes.Buffer)
	if err := ticket.UpdatedEstimationDetail(ticketID, averageEstimate, medianEstimate, stdEstimate, estimatedBy).
		Render(context.Background(), renderedTicket); err != nil {
		log.Printf("Error rendering ticket thumbnail: %v", err)
		return
	}
	bytes := renderedTicket.Bytes()
	mutex.RLock()
	conns := rooms[roomID]
	mutex.RUnlock()
	for conn := range conns {
		buffer <- message{conn: conn, data: &bytes, roomID: roomID}
	}
}

func (w *WebSocketService) SendNewTicket(tticket ticket.TicketDetailProps) {
	renderedTicket := new(bytes.Buffer)
	if err := ticket.CreatedTicketUpdate(tticket, true).
		Render(context.Background(), renderedTicket); err != nil {
		log.Printf("Error rendering ticket thumbnail: %v", err)
		return
	}
	bytes := renderedTicket.Bytes()
	mutex.RLock()
	conns := rooms[tticket.RoomID]
	mutex.RUnlock()
	for conn := range conns {
		buffer <- message{conn: conn, data: &bytes, roomID: tticket.RoomID}
	}
}

func (w *WebSocketService) BulkImportTickets(tickets []ticket.TicketDetailProps) {
	aggregatedRenderedTickets := new(bytes.Buffer)
	slog.Debug("Bulk importing tickets", slog.Any("ticket num", len(tickets)))
	for i := len(tickets) - 1; i >= 0; i-- {
		renderedTicket := new(bytes.Buffer)
		tticket := tickets[i]
		if err := ticket.CreatedTicketUpdate(tticket, i == 0).
			Render(context.Background(), renderedTicket); err != nil {
			log.Printf("Error rendering ticket thumbnail: %v", err)
			return
		}
		bytes := renderedTicket.Bytes()
		aggregatedRenderedTickets.Write(bytes)
	}
	mutex.RLock()
	conns := rooms[tickets[0].RoomID]
	mutex.RUnlock()
	bytes := aggregatedRenderedTickets.Bytes()

	for conn := range conns {
		buffer <- message{conn: conn, data: &bytes, roomID: tickets[0].RoomID}
	}
}

func (w *WebSocketService) Register(conn *websocket.Conn, roomID uint) {
	mutex.Lock()
	if _, ok := rooms[roomID]; !ok {
		rooms[roomID] = make(map[*websocket.Conn]bool)
	}
	rooms[roomID][conn] = true
	mutex.Unlock()

	// Set up ping/pong to keep connection alive
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start a goroutine to read from the websocket to detect disconnects
	go w.readPump(conn, roomID)
}

func NewWebSocketService(ticketService *TicketService) *WebSocketService {
	service := &WebSocketService{
		TicketService: ticketService,
	}

	for range 30 {
		go writePump()
	}

	// Start the cleanup routine
	go service.CleanupInactiveConnections()

	return service
}
