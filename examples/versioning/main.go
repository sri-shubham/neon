package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/sri-shubham/neon"
)

// UserV1 represents the first version of the user service
type UserV1 struct {
	neon.Module `base:"/users" v:"1"`
	getProfile  neon.Get `url:"/profile"`
}

func (s UserV1) GetProfile(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "[v1] User Profile: Alice")
}

// UserV2 represents the second version of the user service
type UserV2 struct {
	neon.Module `base:"/users" v:"2"`
	getProfile  neon.Get `url:"/profile"`
}

func (s UserV2) GetProfile(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "[v2] User Profile: Alice (with more details)")
}

func main() {
	app := neon.New()
	app.SetEnv(neon.DevEnv)

	// Add both versions of the service
	app.AddService(&UserV1{})
	app.AddService(&UserV2{})

	// Context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Println("API Versioning Example running")
	fmt.Println("Neon automatically handles paths with versions from the 'v' tag:")
	fmt.Println("1. V1 Profile: curl http://localhost:8080/v1/users/profile")
	fmt.Println("2. V2 Profile: curl http://localhost:8080/v2/users/profile")

	if err := app.Run(ctx); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
