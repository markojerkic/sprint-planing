package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/markojerkic/spring-planing/cmd/web"
	"github.com/markojerkic/spring-planing/cmd/web/components/privacy"
	"github.com/markojerkic/spring-planing/cmd/web/homepage"
	"github.com/markojerkic/spring-planing/internal/server/auth"
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

	roomTicketService := service.NewRoomTicketService(s.db)
	roomService := service.NewRoomService(s.db, roomTicketService)
	websocketService := service.NewWebSocketService(roomService)
	llmService := service.NewLLMService(websocketService, s.db)
	ticketService := service.NewTicketService(s.db, roomTicketService, websocketService)
	jiraService := service.NewJiraService(ticketService)

	auth.NewOAuthRouter(e.Group("/auth/jira"))
	newRoomRouter(roomService, ticketService, s.db.DB, e.Group("/room"))
	newTicketRouter(ticketService, jiraService, llmService, s.db.DB, e.Group("/ticket"))
	newWebsocketRouter(websocketService, roomService, e.Group("/ws"))
	newJiraRouter(jiraService, s.db.DB, e.Group("/jira"))
	e.GET("/", homepage.HomepageHandler(roomService))
	e.GET("/rooms", homepage.RoomsHandler(roomService))
	e.GET("/privacy", echo.WrapHandler(templ.Handler(privacy.PrivacyPage())))
	e.GET("/terms-of-service", echo.WrapHandler(templ.Handler(privacy.TermsPage())))

	return e
}
