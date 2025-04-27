package server

import (
	"fmt"
	"strconv"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/cmd/web/components/room"
	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
	"github.com/markojerkic/spring-planing/cmd/web/homepage"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/server/auth"
	"github.com/markojerkic/spring-planing/internal/service"
	"gorm.io/gorm"
)

type RoomRouter struct {
	roomService   *service.RoomService
	ticketService *service.TicketService
	db            *gorm.DB
	group         *echo.Group
}

func (r *RoomRouter) createRoomHandler(ctx echo.Context) error {
	user := ctx.Get("user").(database.User)
	name := ctx.FormValue("roomName")

	createdRoom, err := r.roomService.CreateRoom(ctx.Request().Context(), user.ID, name)
	if err != nil {
		ctx.Logger().Errorf("Error creating room: %v", err)
		return ctx.String(500, "Error creating room")
	}

	return ctx.Redirect(302, fmt.Sprintf("/room/%d", createdRoom.ID))
}

func (r *RoomRouter) roomDetailsHandler(ctx echo.Context) error {
	user := ctx.Get("user").(database.User)
	roomID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.Logger().Errorf("Error parsing room id: %v", err)
		return ctx.String(400, "Invalid room id")
	}

	roomDetails, err := r.roomService.GetRoom(ctx.Request().Context(), uint(roomID), user.ID)
	if err != nil {
		ctx.Logger().Errorf("Error getting room: %v", err)
		return ctx.String(500, "Error getting room")
	}

	tickets := roomDetails.TicketsWithStatistics
	ticketDetails := make([]ticket.TicketDetailProps, len(tickets))
	isOwner := roomDetails.CreatedBy == user.ID
	for i, t := range tickets {
		ticketDetails[i] = t.ToDetailProp(isOwner)
	}

	_, isJiraUser := ctx.Get(auth.JiraClientInfoKey).(*auth.JiraClientInfo)
	return room.RoomPage(room.RoomPageProps{
		ID:                 roomDetails.ID,
		Name:               roomDetails.Name,
		CreatedAt:          roomDetails.CreatedAt,
		IsCurrentUserOwner: isOwner,
		IsJiraUser:         isJiraUser,
		Tickets:            ticketDetails,
	}, isOwner).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (r *RoomRouter) deleteRoomHandler(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.String(400, "Invalid room id")
	}

	user := ctx.Get("user").(database.User)
	rooms, err := r.roomService.DeleteRoom(ctx.Request().Context(), uint(id), user.ID)
	if err != nil {
		return ctx.String(500, "Error deleting room")
	}

	return homepage.RoomsPage(rooms, user.ID).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func newRoomRouter(roomService *service.RoomService,
	ticketService *service.TicketService,
	db *gorm.DB,
	group *echo.Group) *RoomRouter {
	r := &RoomRouter{
		roomService:   roomService,
		ticketService: ticketService,
		db:            db,
		group:         group,
	}
	e := r.group
	e.GET("", echo.WrapHandler(templ.Handler(room.CreateRoom())))
	e.POST("", r.createRoomHandler)
	e.GET("/:id", r.roomDetailsHandler)
	e.DELETE("/:id", r.deleteRoomHandler)

	return r
}
