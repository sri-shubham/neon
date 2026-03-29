package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-logr/zapr"
	"github.com/sri-shubham/neon"
	"go.uber.org/zap"
)

type HelloService struct {
	neon.Module `base:"/hello"`
	world       neon.Get `url:"/world"`
}

func (s HelloService) World(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func main() {
	// 1. Initialize Zap logger
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()

	// 2. Wrap it in logr (Neon uses the logr interface)
	logger := zapr.NewLogger(zapLogger)

	// 3. Create app and set the logger
	app := neon.New()
	app.SetLogger(logger)

	app.AddService(&HelloService{})

	// Context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Println("Custom Logger Example running (using Zap)")
	fmt.Println("Check the terminal for Zap-structured logs when you visit:")
	fmt.Println("  curl http://localhost:8080/hello/world")

	if err := app.Run(ctx); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
