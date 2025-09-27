package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/joho/godotenv"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/robfig/cron/v3"
)

type Server struct {
	port int

	db *database.Database
}

func NewServer() *http.Server {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("No .env file found or error loading: %v", err)
	}
	if os.Getenv("APP_ENV") == "local" {
		slog.Info("Loading .env.local")
		if err := godotenv.Overload(".env.local"); err != nil {
			log.Fatalf("No .env.local file found or error loading: %v", err)
		}
	}

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,

		db: database.New(os.Getenv("DB_URL")),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Server running on port %d", NewServer.port)

	NewServer.cleanupCRON()

	return server
}

func (s *Server) cleanupCRON() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		err := s.cleanup(ctx)
		if err != nil {
			log.Printf("Failed to cleanup: %v", err)
		}
	}()

	cron := cron.New()

	// Run every 10 hours
	_, err := cron.AddFunc("@every 10h", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		err := s.cleanup(ctx)
		if err != nil {
			log.Printf("Failed to cleanup: %v", err)
		}
	})

	if err != nil {
		log.Fatalf("Failed to add cleanup job: %v", err)
	}

	cron.Start()
	log.Printf("Cleanup cron job started")

}

func (s *Server) cleanup(ctx context.Context) error {
	slog.Info("Cleanup job started")
	if err := s.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete estimates
		if err := tx.Model(&database.Estimate{}).
			Where("created_at < NOW() - INTERVAL '10 days'").
			Delete(&database.Estimate{}).Error; err != nil {
			slog.Error("Failed to delete estimates", slog.Any("error", err))
			return err
		}
		// Delete tickets which are older than 10 days
		if err := tx.Model(&database.Ticket{}).
			Where("created_at < NOW() - INTERVAL '10 days'").
			Delete(&database.Ticket{}).Error; err != nil {
			slog.Error("Failed to delete tickets", slog.Any("error", err))
			return err
		}
		// Delete rooms which have no non-deleted tickets and are older than 10 days
		if err := tx.Model(&database.Room{}).
			Where("id NOT IN (SELECT room_id FROM tickets WHERE room_id IS NOT NULL AND deleted_at IS NULL)").
			Where("created_at < NOW() - INTERVAL '10 days'").
			Delete(&database.Room{}).Error; err != nil {
			slog.Error("Failed to delete rooms", slog.Any("error", err))
			return err
		}
		// Delete users which have no non-deleted rooms or estimates
		if err := tx.Model(&database.User{}).
			Where("id NOT IN (SELECT user_id FROM estimates WHERE user_id IS NOT NULL AND deleted_at IS NULL)").
			Where("id NOT IN (SELECT user_id FROM room_users WHERE user_id IS NOT NULL AND deleted_at IS NULL)").
			Where("created_at < NOW() - INTERVAL '35 days'").
			Delete(&database.User{}).Error; err != nil {
			slog.Error("Failed to delete users", slog.Any("error", err))
			return err
		}
		return nil
	}); err != nil {
		slog.Error("Failed to run cleanup transaction", slog.Any("error", err))
		return err
	}
	return nil
}
