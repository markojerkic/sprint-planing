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
