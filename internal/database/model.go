package database

import (
	"time"

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
	Tickets   []Ticket
}

type Ticket struct {
	gorm.Model
	Name        string
	Description string
	ClosedAt    *time.Time
	Hidden      bool `gorm:"default:false"`
	RoomID      uint
	Estimates   []Estimate
}

type Estimate struct {
	gorm.Model
	TicketID uint
	UserID   uint
	Estimate int
}
