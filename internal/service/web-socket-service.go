package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
)

var subscriptions = make(map[*websocket.Conn]Route)
var mutex = sync.RWMutex{}

func getMatchingSubscriptions(route Route) []*websocket.Conn {
	mutex.RLock()
	defer mutex.RUnlock()

	conns := make([]*websocket.Conn, 0, len(subscriptions))

	for conn, subRoute := range subscriptions {
		if subRoute.Matches(route) {
			conns = append(conns, conn)
		}
	}

	return conns
}

type message struct {
	conn   *websocket.Conn
	data   *[]byte
	roomID uint
}

var buffer = make(chan message, 100)

type WebSocketService struct {
	roomService *RoomService
}

func writePump() {
	for msg := range buffer {
		if err := msg.conn.WriteMessage(websocket.TextMessage, *msg.data); err != nil {
			log.Printf("Error writing message: %v", err)
			removeConnection(msg.conn)
		}
	}
}

// removeConnection removes a connection from a room and cleans up empty rooms
func removeConnection(conn *websocket.Conn) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(subscriptions, conn)
}

// readPump reads from the websocket connection to detect disconnects
func (w *WebSocketService) readPump(conn *websocket.Conn, route Route) {
	defer func() {
		conn.Close()
		removeConnection(conn)
		log.Printf("Connection closed for room %s", string(route))
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
		slog.Debug("Checking for inactive connections. Current rooms", slog.Int("rooms", len(subscriptions)))

		for conn, topic := range subscriptions {
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				log.Printf("Failed to ping client in room %s: %v", topic, err)
				delete(subscriptions, conn)
				conn.Close()
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
	conns := getMatchingSubscriptions(Route(fmt.Sprintf("room/%d/*", roomID)))
	mutex.RUnlock()
	for _, conn := range conns {
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
	conns := getMatchingSubscriptions(Route(fmt.Sprintf("room/%d/estimator", roomID)))
	mutex.RUnlock()
	for _, conn := range conns {
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
	estimatorConns := getMatchingSubscriptions(Route(fmt.Sprintf("room/%d/estimator", tticket.RoomID)))
	allConns := getMatchingSubscriptions(Route(fmt.Sprintf("room/%d/*", tticket.RoomID)))
	mutex.RUnlock()
	for _, conn := range estimatorConns {
		buffer <- message{conn: conn, data: &bytes, roomID: tticket.RoomID}
	}

	if estimate, err := w.roomService.GetTotalEstimateOfRoom(context.Background(), tticket.RoomID); err == nil {
		bytes = fmt.Appendf(nil, `<div hx-swap-oob="innerHtml:#total-estimated">%s</div>`, estimate)
		for _, conn := range allConns {
			buffer <- message{conn: conn, data: &bytes, roomID: tticket.RoomID}
		}
	}

}

func (w *WebSocketService) SendLLMRecommendation(ticketID uint, jiraKey *string, roomID uint, llmRecommendation string) {
	delta := fmt.Sprintf(`<div hx-swap-oob="innerHtml:form[data-estimation-form='%d' ] > span.llm-recommendation">%s</div>`, ticketID, llmRecommendation)

	mutex.RLock()
	conns := getMatchingSubscriptions(Route(fmt.Sprintf("room/%d/*", roomID)))
	mutex.RUnlock()

	deltaBytes := []byte(delta)

	for _, conn := range conns {
		buffer <- message{conn: conn, data: &deltaBytes, roomID: roomID}
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
	conns := getMatchingSubscriptions(Route(fmt.Sprintf("room/%d/*", roomID)))
	mutex.RUnlock()
	for _, conn := range conns {
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
	conns := getMatchingSubscriptions(Route(fmt.Sprintf("room/%d/estimator", tticket.RoomID)))
	mutex.RUnlock()
	for _, conn := range conns {
		buffer <- message{conn: conn, data: &bytes, roomID: tticket.RoomID}
	}
}

func (w *WebSocketService) BulkImportTickets(tickets []ticket.TicketDetailProps) {
	if len(tickets) == 0 {
		return
	}

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
	conns := getMatchingSubscriptions(Route(fmt.Sprintf("room/%d/estimator", tickets[0].RoomID)))
	mutex.RUnlock()
	bytes := aggregatedRenderedTickets.Bytes()

	for _, conn := range conns {
		buffer <- message{conn: conn, data: &bytes, roomID: tickets[0].RoomID}
	}
}

func (w *WebSocketService) Register(conn *websocket.Conn, roomID uint, isOwner bool) {
	mutex.Lock()

	var routeSuffix string
	if isOwner {
		routeSuffix = "owner"
	} else {
		routeSuffix = "estimator"
	}

	route := Route(fmt.Sprintf("room/%d/%s", roomID, routeSuffix))
	subscriptions[conn] = route
	mutex.Unlock()

	// Set up ping/pong to keep connection alive
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start a goroutine to read from the websocket to detect disconnects
	go w.readPump(conn, route)
}

func NewWebSocketService(roomService *RoomService) *WebSocketService {
	service := &WebSocketService{
		roomService: roomService,
	}

	for range 30 {
		go writePump()
	}

	// Start the cleanup routine
	go service.CleanupInactiveConnections()

	return service
}
