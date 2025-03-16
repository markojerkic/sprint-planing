package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/markojerkic/spring-planing/internal/database/dbgen"
	"github.com/pressly/goose/v3"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	Pool    *pgxpool.Pool
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
	config, err := pgxpool.ParseConfig(dburl)
	if err != nil {
		log.Fatalf("Unable to parse database URL: %v", err)
	}

	// Configure pool settings
	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// Connect to database
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	log.Printf("Connected to database: %s", dburl)

	dbInstance = &Database{
		Pool: pool,
	}
	dbInstance.runMigrations()

	// Initialize queries with the connection pool
	dbInstance.Queries = dbgen.New(pool)

	return dbInstance
}

func (s *Database) runMigrations() {
	// Set the goose environment
	goose.SetBaseFS(embedMigrations)

	// Get a single connection from the pool for migrations
	conn, err := s.Pool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("failed to acquire connection for migrations: %v", err)
	}
	defer conn.Release()

	// Optional: Set goose dialect to postgres
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}

	// Get the underlying *sql.DB from pgx
	sqlDB, err := conn.Conn().PgConn().ConnConfig().ConnString()
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	// Run the migrations using a temporary sql.DB
	db, err := goose.OpenDBWithDriver("postgres", sqlDB)
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
	err := s.Pool.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats
	poolStats := s.Pool.Stat()
	stats["total_connections"] = strconv.Itoa(poolStats.TotalConns())
	stats["acquired_connections"] = strconv.Itoa(poolStats.AcquiredConns())
	stats["idle_connections"] = strconv.Itoa(poolStats.IdleConns())

	// Evaluate stats to provide a health message
	if poolStats.AcquiredConns() > 8 { // Assuming 10 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if poolStats.TotalConns() == 10 && poolStats.IdleConns() == 0 {
		stats["message"] = "All connections are in use, indicating potential bottlenecks."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *Database) Close() error {
	log.Printf("Disconnected from database: %s", dburl)
	s.Pool.Close()
	return nil
}
