package database

import (
	"fmt"
	"time"

	"github.com/markojerkic/spring-planing/cmd/web/components/ticket"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	CreatedRoom []Room `gorm:"foreignKey:CreatedBy"`
	InRoom      []Room `gorm:"many2many:room_users;"`
	Estimates   []Estimate
}

type Room struct {
	gorm.Model
	CreatedBy uint
	Name      string
	Tickets   []Ticket
	Users     []User `gorm:"many2many:room_users;"`
}

type Ticket struct {
	gorm.Model
	Name        string
	Description string
	ClosedAt    *time.Time
	Hidden      bool `gorm:"default:false"`
	RoomID      uint
	Room        Room `gorm:"foreignKey:RoomID"`
	CreatedBy   uint
	Estimates   []Estimate
}

type Estimate struct {
	gorm.Model
	TicketID uint
	UserID   uint
	Estimate int
}

func (t *Ticket) ToDetailProp(numUsersInRoom int) ticket.TicketDetailProps {
	return ticket.TicketDetailProps{
		ID:          t.ID,
		Name:        t.Name,
		RoomID:      t.RoomID,
		Description: t.Description,
		EstimatedBy: fmt.Sprintf("%d/%d", len(t.Estimates), numUsersInRoom),
		IsClosed:    t.ClosedAt != nil,
	}

}
