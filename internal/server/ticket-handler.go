package server

import (
	"github.com/labstack/echo/v4"
	"github.com/markojerkic/spring-planing/internal/database"
)

type TicketRouter struct {
	db    *database.Database
	group *echo.Group
}

func newTicketRouter(db *database.Database, group *echo.Group) *TicketRouter {
	r := &TicketRouter{
		db:    db,
		group: group,
	}
	// e := r.group

	return r
}
