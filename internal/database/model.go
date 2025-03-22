package database

import (
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
	CreatedBy             uint
	Name                  string
	Tickets               []Ticket
	TicketsWithStatistics []TicketWithEstimateStatistics `gorm:"-"`
	Users                 []User                         `gorm:"many2many:room_users;"`
}

type Estimate struct {
	gorm.Model
	TicketID uint
	UserID   uint
	Estimate int
}
