package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/sri-shubham/neon"
	"go.uber.org/zap"
)

func main() {
	app := neon.New(&neon.Config{
		Port: 8080,
	})
	app.SetEnv(neon.ProdEnv)

	var log logr.Logger

	zapLog, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	log = zapr.NewLogger(zapLog)

	app.Logger = log

	// 1. GLOBAL MIDDLEWARE - Applied to ALL endpoints
	app.AddMiddleware(simpleLogger)
	app.AddMiddleware(globalMiddleware)

	// Register named middlewares for service/endpoint use
	app.RegisterMiddleware("Auth", authMiddleware)
	app.RegisterMiddleware("CORS", corsMiddleware)
	app.RegisterMiddleware("RateLimit", rateLimitMiddleware)

	app.AddService(&UserService{})
	app.AddService(&ProductService{})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	app.Run(ctx)
}

// UserService with service-level middleware
type UserService struct {
	neon.Module  `base:"/users" v:"1" middleware:"Auth,CORS,Auth"`
	getUser      neon.Get  `url:"/{id}" middleware:"Auth,Auth"`       // Single parameter
	getUserPosts neon.Get  `url:"/{id}/posts/{postId}" middleware:""` // Multiple parameters
	createUser   neon.Post `url:"/" middleware:"RateLimit"`           // Additional endpoint middleware
	updateUser   neon.Put  `url:"/{id}" middleware:""`                // Update with ID parameter
}

// ProductService with different service-level middleware
type ProductService struct {
	neon.Module   `base:"/products" v:"1" middleware:"RateLimit"`
	getProduct    neon.Get  `url:"/{id}" middleware:""`      // Named parameter
	createProduct neon.Post `url:"/" middleware:"Auth,CORS"` // Additional endpoint middleware
}

// Global middleware (applied to ALL endpoints)
func simpleLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("📝 Simple Logger - %s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func globalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("🌍 Global Middleware - Applied to ALL endpoints")
		next.ServeHTTP(w, r)
	})
}

// Named middlewares for service/endpoint use
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("🔐 Auth Middleware - Checking authentication")
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("🌐 CORS Middleware - Setting CORS headers")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("⏱️  Rate Limit Middleware - Checking rate limits")
		next.ServeHTTP(w, r)
	})
}

// UserService handlers
func (s UserService) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract path parameter
	userID := r.PathValue("id")
	fmt.Printf("👤 UserService.GetUser - Handler executed for user ID: %s\n", userID)
	w.Write([]byte(fmt.Sprintf("User data for ID: %s", userID)))
}

func (s UserService) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	// Extract multiple path parameters
	userID := r.PathValue("id")
	postID := r.PathValue("postId")
	fmt.Printf("👤 UserService.GetUserPosts - User: %s, Post: %s\n", userID, postID)
	w.Write([]byte(fmt.Sprintf("Post %s from User %s", postID, userID)))
}

func (s UserService) CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("👤 UserService.CreateUser - Handler executed")
	w.Write([]byte("User created"))
}

func (s UserService) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	fmt.Printf("👤 UserService.UpdateUser - Handler executed for user ID: %s\n", userID)
	w.Write([]byte(fmt.Sprintf("User %s updated", userID)))
}

// ProductService handlers
func (s ProductService) GetProduct(w http.ResponseWriter, r *http.Request) {
	// Extract path parameter
	productID := r.PathValue("id")
	fmt.Printf("📦 ProductService.GetProduct - Handler executed for product ID: %s\n", productID)
	w.Write([]byte(fmt.Sprintf("Product data for ID: %s", productID)))
}

func (s ProductService) CreateProduct(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📦 ProductService.CreateProduct - Handler executed")
	w.Write([]byte("Product created"))
}
