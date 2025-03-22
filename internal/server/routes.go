package server

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/markojerkic/spring-planing/cmd/web"
	"github.com/markojerkic/spring-planing/cmd/web/homepage"
	"github.com/markojerkic/spring-planing/internal/service"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Use(middleware.Recover())
	s.InitSessions(e)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	fileServer := http.FileServer(http.FS(web.Files))
	e.GET("/assets/*", echo.WrapHandler(fileServer))

	roomService := service.NewRoomService(s.db)
	ticketService := service.NewTicketService(s.db)
	// websocketService := service.NewWebSocketService(ticketService)

	newRoomRouter(roomService, ticketService, s.db.DB, e.Group("/room"))
	newTicketRouter(ticketService, s.db.DB, e.Group("/ticket"))
	// newWebsocketRouter(websocketService, e.Group("/ws"))
	e.GET("/", homepage.HomepageHandler(roomService))

	return e
}
