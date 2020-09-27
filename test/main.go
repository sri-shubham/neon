package main

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"

	"github.com/go-chi/chi/middleware"
	"github.com/sri-shubham/neon"
)

func main() {
	app := neon.New()
	app.Port = 9999
	app.AddMiddleware(middleware.Logger)
	app.RegisterMiddleware("UserCtx", UserService{}.UserCtx)
	app.RegisterMiddleware("ReqID", middleware.RequestID)
	app.AddService(&UserService{})
	fmt.Println(app.Run())
}

// UserService : List User Services
type UserService struct {
	neon.Module `base:"/user" v:"1" middleware:"ReqID,UserCtx"`
	getUser     neon.Get `middleware:""`
	// createUser  neon.Post `url:"/"`
}

// UserCtx : Middleware create request ctx
func (s UserService) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received Request on User", runtime.FuncForPC(reflect.ValueOf(next).Pointer()).Name())
	})
}

// GetUser : Get User
func (s UserService) GetUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	fmt.Fprintf(w, fmt.Sprint(r.Header))
}
