package server

import (
	"fmt"
	"strconv"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/cmd/web/components/room"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
	"github.com/markojerkic/spring-planing/internal/service"
)

type RoomRouter struct {
	roomService   *service.RoomService
	ticketService *service.TicketService
	group         *echo.Group
}

func (r *RoomRouter) createRoomHandler(ctx echo.Context) error {
	user := ctx.Get("user").(dbgen.User)
	name := ctx.FormValue("roomName")

	createdRoom, err := r.roomService.CreateRoom(ctx.Request().Context(), user.ID, name)
	if err != nil {
		ctx.Logger().Errorf("Error creating room: %v", err)
		return ctx.String(500, "Error creating room")
	}

	return ctx.Redirect(302, fmt.Sprintf("/room/%d", createdRoom.ID))
}

func (r *RoomRouter) roomDetailsHandler(ctx echo.Context) error {
	user := ctx.Get("user").(dbgen.User)
	roomID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.Logger().Errorf("Error parsing room id: %v", err)
		return ctx.String(400, "Invalid room id")
	}

	roomDetails, err := r.roomService.GetRoom(ctx.Request().Context(), roomID, user.ID)
	if err != nil {
		ctx.Logger().Errorf("Error getting room: %v", err)
		return ctx.String(500, "Error getting room")
	}

	roomTickets, err := r.ticketService.GetTicketsOfRoom(ctx.Request().Context(), roomID, user.ID, nil)
	if err != nil {
		ctx.Logger().Errorf("Error getting room tickets: %v", err)
		return ctx.String(500, "Error getting room tickets")
	}

	return room.RoomPage(room.RoomPageProps{
		ID:                 roomDetails.ID,
		Name:               roomDetails.Name,
		CreatedAt:          roomDetails.CreatedAt.Time,
		IsCurrentUserOwner: roomDetails.IsOwner,
		Tickets:            roomTickets,
	}).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func newRoomRouter(db *database.Database, group *echo.Group) *RoomRouter {
	r := &RoomRouter{
		roomService:   service.NewRoomService(db),
		ticketService: service.NewTicketService(db),
		group:         group,
	}
	e := r.group
	e.GET("", echo.WrapHandler(templ.Handler(room.CreateRoom())))
	e.POST("", r.createRoomHandler)
	e.GET("/:id", r.roomDetailsHandler)

	return r
}
