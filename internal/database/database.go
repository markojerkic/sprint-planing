package database

import (
	"context"
	"database/sql"
	"embed"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
	"github.com/pressly/goose/v3"

	// This is the important import for goose to find the postgres driver
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
	Conn    *pgx.Conn
	Queries *dbgen.Queries
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

	// Parse connection string
	conn, err := pgx.Connect(context.Background(), dburl)
	if err != nil {
		log.Fatalf("failed to parse connection string: %v", err)
	}

	dbInstance = &Database{
		Conn: conn,
	}
	dbInstance.runMigrations()

	// Initialize queries with the connection pool
	dbInstance.Queries = dbgen.New(dbInstance.Conn)

	return dbInstance
}

func (s *Database) runMigrations() {
	// Set the goose environment
	goose.SetBaseFS(embedMigrations)

	// Optional: Set goose dialect to postgres
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}

	// Open the database using the pgx driver via its stdlib adapter
	db, err := sql.Open("pgx", dburl)
	if err != nil {
		log.Fatalf("failed to open DB for migrations: %v", err)
	}
	defer db.Close()

	// Run the migrations
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *Database) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	stats := make(map[string]string)

	// Ping the database
	if err := s.Conn.Ping(ctx); err != nil {
		stats["status"] = "down"
		stats["error"] = err.Error()
		return stats
	}

	// Get the database version
	var version string
	if err := s.Conn.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
		stats["status"] = "down"
		stats["error"] = err.Error()
		return stats
	}
	stats["status"] = "up"
	stats["version"] = version

	// Get the number of active connections
	var connections int
	if err := s.Conn.QueryRow(ctx, "SELECT COUNT(*) FROM pg_stat_activity").Scan(&connections); err != nil {
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
	return s.Conn.Close(context.Background())
}
