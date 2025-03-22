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

		db: database.New(),
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

	// Run every 10 days
	_, err := cron.AddFunc("@every 240h", func() {
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
	slog.Warn("cleanup function not implemented")
	// tx, err := s.db.DB.BeginTx(ctx, pgx.TxOptions{})
	// if err != nil {
	// 	log.Printf("Failed to begin transaction: %v", err)
	// 	return err
	// }
	// defer tx.Rollback(ctx)
	// qtx := s.db.Queries.WithTx(tx)
	//
	// if err := qtx.CleanupOldTickets(ctx); err != nil {
	// 	return fmt.Errorf("failed to cleanup ticket estimates: %w", err)
	// }
	//
	// if err := qtx.CleanupClosedTickets(ctx); err != nil {
	// 	return fmt.Errorf("failed to cleanup closed tickets: %w", err)
	// }
	//
	// if err := qtx.CleanupUnusedRooms(ctx); err != nil {
	// 	return fmt.Errorf("failed to cleanup unused rooms: %w", err)
	// }
	//
	// if err := qtx.CleanupUnusedUsers(ctx); err != nil {
	// 	return fmt.Errorf("failed to cleanup unused users: %w", err)
	// }
	//
	// if err := tx.Commit(ctx); err != nil {
	// 	log.Printf("Failed to commit transaction: %v", err)
	// 	return err
	// }
	// return nil
	return nil
}
