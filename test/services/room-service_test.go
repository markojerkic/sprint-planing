package services

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/markojerkic/spring-planing/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

type RoomServiceSuite struct {
	suite.Suite
	postgresContainer *postgres.PostgresContainer
	roomService       *service.RoomService
	db                *database.Database
}

// TearDownSubTest implements suite.TearDownSubTest.
func (r *RoomServiceSuite) TearDownSubTest() {
	// Truncate all tables
	var allTables []string
	err := r.db.DB.Raw("SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema'").Scan(&allTables).Error
	assert.NoError(r.T(), err)

	for _, table := range allTables {
		err := r.db.DB.Exec("TRUNCATE TABLE " + table + " CASCADE").Error
		assert.NoError(r.T(), err)
	}
	assert.NoError(r.T(), err)
}

// SetupSubTest implements suite.SetupSubTest.
func (r *RoomServiceSuite) SetupTest() {
	// Prepare user with id 1
	var savedUser database.User
	err := r.db.DB.Create(&database.User{
		Model: gorm.Model{
			ID: 1,
		},
	}).Scan(&savedUser).Error
	assert.NoError(r.T(), err)
}

// SetupSuite implements suite.SetupAllSuite.
func (r *RoomServiceSuite) SetupSuite() {
	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "trust",
			"POSTGRES_SSL":              "false",
		}),
	)
	if err != nil {
		log.Fatalf("Failed to start postgres container: %v", err)
	}
	r.postgresContainer = postgresContainer

	connString, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		r.T().Fatal(err)
	}

	db := database.New(connString)
	r.db = db // Add this line
	r.roomService = service.NewRoomService(db)

}

// TearDownSuite implements suite.TearDownAllSuite.
func (r *RoomServiceSuite) TearDownSuite() {
	if r.postgresContainer != nil {
		if err := r.postgresContainer.Terminate(r.T().Context()); err != nil {
			r.T().Fatalf("failed to terminate postgres container: %v", err)
		}
	}
}

var _ suite.TearDownAllSuite = &RoomServiceSuite{}
var _ suite.SetupAllSuite = &RoomServiceSuite{}
var _ suite.SetupTestSuite = &RoomServiceSuite{}
var _ suite.TearDownSubTest = &RoomServiceSuite{}

func (r *RoomServiceSuite) TestCreateRoom() {
	t := r.T()
	ctx := t.Context()

	room, err := r.roomService.CreateRoom(ctx, 1, "roomName")

	assert.NoError(t, err, "Error creating room")
	assert.Equal(t, "roomName", room.Name)
	assert.Equal(t, uint(1), room.CreatedBy)
	assert.Equal(t, 1, len(room.Users))
	assert.Equal(t, 0, len(room.Tickets))

}

func TestRoomServiceSuite(t *testing.T) {
	suite.Run(t, new(RoomServiceSuite))
}
