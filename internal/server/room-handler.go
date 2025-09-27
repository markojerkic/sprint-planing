package server

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/cmd/web/components/room"
	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
	"github.com/markojerkic/spring-planing/cmd/web/homepage"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/server/auth"
	"github.com/markojerkic/spring-planing/internal/service"
	"github.com/markojerkic/spring-planing/internal/util"
	"gorm.io/gorm"
)

type RoomRouter struct {
	roomService   *service.RoomService
	ticketService *service.TicketService
	db            *gorm.DB
	group         *echo.Group
}

type RoomTicket struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	IsClosed bool   `json:"isClosed"`
}

func (r *RoomRouter) createRoomHandler(ctx echo.Context) error {
	user := ctx.Get("user").(database.User)
	name := ctx.FormValue("roomName")
	allowLLM := ctx.FormValue("allowLLM") == "on"

	createdRoom, err := r.roomService.CreateRoom(ctx.Request().Context(), user.ID, name, allowLLM)
	if err != nil {
		ctx.Logger().Errorf("Error creating room: %v", err)
		return ctx.String(500, "Error creating room")
	}

	return ctx.Redirect(302, fmt.Sprintf("/room/%d", createdRoom.ID))
}

func (r *RoomRouter) roomTicketsHandler(ctx echo.Context) error {
	tickets := make([]database.Ticket, 0)
	if err := r.db.Model(&database.Ticket{}).
		Select("id, name, closed_at").
		Where("room_id = ?", ctx.Param("id")).
		Order("id desc").
		Find(&tickets).Error; err != nil {
		ctx.Logger().Errorf("Error getting tickets: %v", err)
		return ctx.String(500, "Error getting tickets")
	}
	ticketDtos := make([]RoomTicket, len(tickets))
	for i, t := range tickets {
		ticketDtos[i] = RoomTicket{
			ID:       t.ID,
			Name:     t.Name,
			IsClosed: t.ClosedAt != nil,
		}
	}

	return ctx.JSON(200, ticketDtos)
}

func (r *RoomRouter) roomDetailsHandler(ctx echo.Context) error {
	user := ctx.Get("user").(database.User)
	roomID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.Logger().Errorf("Error parsing room id: %v", err)
		util.AddToastHeader(ctx, "Invalid room id", util.INFO)
		return ctx.String(400, "Invalid room id")
	}

	roomDetails, err := r.roomService.GetRoom(ctx.Request().Context(), uint(roomID), user.ID)
	if err != nil {
		ctx.Logger().Errorf("Error getting room: %v", err)
		util.AddToastHeader(ctx, "Room not found", util.INFO)
		return ctx.String(500, "Error getting room")
	}

	tickets := roomDetails.TicketsWithStatistics
	ticketDetails := make([]ticket.TicketDetailProps, len(tickets))
	isOwner := roomDetails.CreatedBy == user.ID
	for i, t := range tickets {
		ticketDetails[i] = t.ToDetailProp(isOwner)
	}

	totalEstimated, err := r.roomService.GetTotalEstimateOfRoom(ctx.Request().Context(), uint(roomID))

	_, isJiraUser := ctx.Get(auth.JiraClientInfoKey).(*auth.JiraClientInfo)
	return room.RoomPage(room.RoomPageProps{
		ID:                 roomDetails.ID,
		Name:               roomDetails.Name,
		CreatedAt:          roomDetails.CreatedAt,
		IsCurrentUserOwner: isOwner,
		TotalEstimated:     totalEstimated,
		IsJiraUser:         isJiraUser,
		IsLlmEnabled:       roomDetails.AllowLLMEstimation,
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

func (r *RoomRouter) allowLlmEstimationHandler(ctx echo.Context) error {
	user := ctx.Get("user").(database.User)
	roomID, err := strconv.Atoi(ctx.FormValue("roomId"))
	if err != nil {
		return ctx.String(400, "Invalid room id")
	}
	allowLlmEstimation := ctx.FormValue("allowLLM") == "on"
	slog.Debug("Allow LLM estimation", "roomId", roomID, "allowLlmEstimation", allowLlmEstimation)

	if err := r.db.WithContext(ctx.Request().Context()).
		Transaction(func(tx *gorm.DB) error {
			var room database.Room
			if err := tx.First(&room, uint(roomID)).Error; err != nil {
				return err
			}

			if room.CreatedBy != user.ID {
				return gorm.ErrRecordNotFound
			}

			if err := tx.Model(&database.Room{}).
				Where("id = ?", roomID).
				Update("allow_llm_estimation", allowLlmEstimation).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
		slog.Error("Error updating room", "error", err)
		return ctx.String(500, "Error updating room")
	}

	var toastMessage string
	if allowLlmEstimation {
		toastMessage = "Successfully enabled LLM estimation"
	} else {
		toastMessage = "Disabled LLM estimation"
	}
	if err = util.AddToastHeader(ctx, toastMessage, util.INFO); err != nil {
		slog.Error("Error adding toast header", "error", err)
		return ctx.String(500, "Error adding toast header")
	}

	return room.AllowLlmEstimationForm(allowLlmEstimation).Render(ctx.Request().Context(), ctx.Response().Writer)
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
	e.GET("", func(c echo.Context) error {
		_, isJiraUser := c.Get(auth.JiraClientInfoKey).(*auth.JiraClientInfo)
		return room.CreateRoom(isJiraUser).Render(c.Request().Context(), c.Response().Writer)
	})
	e.POST("", r.createRoomHandler)
	e.GET("/:id", func(c echo.Context) error {
		if c.Request().Header.Get("Accept") == "application/json" {
			return r.roomTicketsHandler(c)
		}

		return r.roomDetailsHandler(c)
	})
	e.DELETE("/:id", r.deleteRoomHandler)
	e.POST("/allow-llm-estimation", r.allowLlmEstimationHandler)

	return r
}
