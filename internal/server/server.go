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

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"

	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/robfig/cron/v3"
)

type Server struct {
	port int

	db *database.Database
}

func NewServer() *http.Server {
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
		if err := tx.Exec(`
            DELETE FROM tickets
            WHERE created_at < NOW() - INTERVAL '10 days';
        `).Error; err != nil {
			slog.Error("Failed to delete tickets", slog.Any("error", err))
			return err
		}
		// Delete rooms which have no tickets and are older than 10 days
		if err := tx.Exec(`
            DELETE FROM rooms
            WHERE id NOT IN (
                SELECT room_id FROM tickets WHERE room_id IS NOT NULL
            ) AND created_at < NOW() - INTERVAL '10 days';
        `).Error; err != nil {
			slog.Error("Failed to delete rooms", slog.Any("error", err))
			return err
		}
		// Delete room_users entries for users not in any rooms
		if err := tx.Exec(`
            DELETE FROM room_users
            WHERE user_id NOT IN (
                SELECT DISTINCT user_id FROM room_users ru
                JOIN rooms r ON ru.room_id = r.id
            );
            `).Error; err != nil {
			slog.Error("Failed to delete room_users", slog.Any("error", err))
			return err
		}
		// Delete users which have no rooms and are older than 10 days
		if err := tx.Exec(`
            DELETE FROM users
            WHERE id NOT IN (
                SELECT user_id FROM room_users
            ) AND created_at < NOW() - INTERVAL '10 days';
        `).Error; err != nil {
			slog.Error("Failed to delete users", slog.Any("error", err))
			return err
		}
		return nil
	}); err != nil {
		slog.Error("Failed to run cleanup transaction", slog.Any("error", err))
		return err // Added to properly return the error
	}
	return nil
}
