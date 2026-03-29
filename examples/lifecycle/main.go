package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sri-shubham/neon"
)

// ManagedService demonstrates lifecycle hooks
type ManagedService struct {
	neon.Module `base:"/system"`

	// Dependencies
	dbConnected bool
	stopChan    chan struct{}

	// Endpoints
	status neon.Get `url:"/status"`
}

// Init is called automatically by Neon before the server starts
func (s *ManagedService) Init() error {
	fmt.Println("[ManagedService] Initializing resources...")
	time.Sleep(1 * time.Second) // Simulate DB connection
	s.dbConnected = true
	s.stopChan = make(chan struct{})

	// Start a background worker
	go s.backgroundWorker()

	fmt.Println("[ManagedService] Initialization complete.")
	return nil
}

// HealthCheck is used by the automatic /health endpoint
func (s *ManagedService) HealthCheck() error {
	if !s.dbConnected {
		return fmt.Errorf("database not connected")
	}
	return nil
}

// Close is called automatically during graceful shutdown
func (s *ManagedService) Close(ctx context.Context) error {
	fmt.Println("[ManagedService] Shutting down...")
	close(s.stopChan)

	// Simulate cleanup
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("[ManagedService] Cleanup finished.")
	case <-ctx.Done():
		fmt.Println("[ManagedService] Shutdown timed out!")
	}
	return nil
}

func (s *ManagedService) Status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "System is running. Database connected: %v", s.dbConnected)
}

func (s *ManagedService) backgroundWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("[ManagedService] Performing background maintenance...")
		case <-s.stopChan:
			fmt.Println("[ManagedService] Background worker stopping...")
			return
		}
	}
}

func main() {
	// Create app in Development mode to see detailed logs
	app := neon.New()
	app.SetEnv(neon.DevEnv)

	app.AddService(&ManagedService{})

	// Setup context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Println("Neon Lifecycle Demo running at http://localhost:8080")
	fmt.Println("1. Check status: curl http://localhost:8080/system/status")
	fmt.Println("2. Check health: curl http://localhost:8080/health")
	fmt.Println("3. Press Ctrl+C to see graceful shutdown in action")

	if err := app.Run(ctx); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
