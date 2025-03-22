package database

import (
	"database/sql"
	"embed"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string
	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
}

type Database struct {
	DB    *gorm.DB
	sqlDB *sql.DB
}

var (
	dburl      = os.Getenv("DB_URL")
	dbInstance *Database
)

func New() *Database {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	db, err := gorm.Open(postgres.Open(dburl), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	// AutoMigrate
	db.AutoMigrate(&User{}, &Room{}, &Ticket{}, &Estimate{})

	dbInstance = &Database{
		DB: db,
	}

	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *Database) Health() map[string]string {
	stats := make(map[string]string)

	// Ping the database
	if err := s.sqlDB.Ping(); err != nil {
		stats["status"] = "down"
		stats["error"] = err.Error()
		return stats
	}

	// Get the database version
	var version string
	if err := s.sqlDB.QueryRow("SELECT version()").Scan(&version); err != nil {
		stats["status"] = "down"
		stats["error"] = err.Error()
		return stats
	}
	stats["status"] = "up"
	stats["version"] = version

	// Get the number of active connections
	var connections int
	if err := s.sqlDB.QueryRow("SELECT COUNT(*) FROM pg_stat_activity").Scan(&connections); err != nil {
		stats["connections"] = "unknown"
	} else {
		stats["connections"] = strconv.Itoa(connections)
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *Database) Close() error {
	log.Printf("Disconnected from database: %s", dburl)
	s.sqlDB.Close()
	return nil
}
