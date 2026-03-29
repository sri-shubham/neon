package neon

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

type Env int

func (m Env) String() (out string) {
	switch m {
	case DevEnv:
		out = "Development"
	case ProdEnv:
		out = "Production"
	default:
		out = "Not Defined"
	}
	return
}

const (
	DevEnv = iota
	ProdEnv
)

type App struct {
	Env     Env
	Port    int
	TLSCert string
	TLSKey  string

	Logger logr.Logger

	mux               *http.ServeMux
	services          []Moduler
	middleware        map[string]Middleware
	globalMiddlewares []Middleware
	routes            map[string]map[string]http.HandlerFunc // path -> method -> handler
}

func (s *App) SetEnv(e Env) {
	s.Env = e
}

// Config :
type Config struct {
	Port    int
	TLSCert string
	TLSKey  string
}

type Middleware func(http.Handler) http.Handler

// New : Create a New Server
func New(conf ...*Config) *App {
	app := new(App)
	app.middleware = make(map[string]Middleware)
	app.mux = http.NewServeMux()
	app.globalMiddlewares = make([]Middleware, 0)
	app.routes = make(map[string]map[string]http.HandlerFunc)
	app.Logger = logr.Discard() // Initialize with no-op logger by default
	if len(conf) > 0 {
		if conf[0].Port != 0 {
			app.Port = conf[0].Port
		} else {
			app.Port = 8080
		}
		app.TLSCert = conf[0].TLSCert
		app.TLSKey = conf[0].TLSKey
	} else {
		app.Port = 8080
	}
	return app
}

// SetLogger allows the user to provide their own logr.Logger implementation
func (s *App) SetLogger(logger logr.Logger) {
	s.Logger = logger
}

// Add a middleware for services
func (s *App) AddMiddleware(fun Middleware) {
	s.globalMiddlewares = append(s.globalMiddlewares, fun)
}

func (s *App) RegisterMiddleware(name string, fn Middleware) {
	if s.middleware == nil {
		s.middleware = make(map[string]Middleware)
	}
	s.middleware[name] = fn
}

// AddService : Add Service to app
// Service must embed neon.Module
func (s *App) AddService(servicePtr Moduler) {
	serviceTypeOf := reflect.TypeOf(servicePtr)
	if serviceTypeOf.Kind() != reflect.Ptr {
		s.Logger.Error(nil, "All services should be passed as a pointer")
		return
	}

	serviceTypeOf = serviceTypeOf.Elem()
	if serviceTypeOf.Kind() != reflect.Struct {
		s.Logger.Error(nil, "service should be a struct")
		return
	}

	module := reflect.TypeOf(Module{})
	field, ok := serviceTypeOf.FieldByName(module.Name())
	if !ok || !field.Anonymous || !field.Type.ConvertibleTo(module) {
		s.Logger.Error(nil, "service struct must add Module Anonymously")
		return
	}

	s.services = append(s.services, servicePtr)
}

// loadAllServices : Builds Routes for all input service structures
func (s *App) loadAllServices() {
	for _, service := range s.services {

		// Reflect Service Data
		module := reflect.TypeOf(Module{})
		serviceType := reflect.TypeOf(service)
		if serviceType.Kind() == reflect.Ptr {
			serviceType = serviceType.Elem()
		}

		serviceValue := reflect.ValueOf(service)
		// Get Module level Configurations
		field, _ := serviceType.FieldByName(module.Name())

		baseURL := field.Tag.Get("base")
		if baseURL == "" {
			baseURL = "/"
		}

		version := field.Tag.Get("v")

		// These middlewares run for specified modules only
		moduleMiddleware := field.Tag.Get("middleware")

		// Get module-level middlewares
		var moduleMiddlewares []Middleware
		var moduleMiddlewareNames []string
		if len(moduleMiddleware) > 0 {
			middlewareNames := strings.Split(moduleMiddleware, ",")
			seen := make(map[string]bool)
			for _, middlewareName := range middlewareNames {
				name := strings.Trim(middlewareName, " ")
				if name == "" {
					continue
				}

				if seen[name] {
					s.Logger.Info("Duplicate middleware registered for module", "base", baseURL, "middleware", name)
				}
				seen[name] = true

				mw, ok := s.middleware[name]
				if !ok {
					s.Logger.Error(nil, "Middleware not registered", "name", name)
					continue
				}
				moduleMiddlewares = append(moduleMiddlewares, mw)
				moduleMiddlewareNames = append(moduleMiddlewareNames, name)
			}
		}

		for i := 0; i < serviceType.NumField(); i++ {
			fieldType := serviceType.FieldByIndex([]int{i})
			switch fieldType.Type.String() {
			case "neon.Get", "neon.Put", "neon.Post", "neon.Patch", "neon.Delete", "neon.Options":
			default:
				continue
			}

			apiURL := fieldType.Tag.Get("url")
			if apiURL == "" {
				apiURL = "/"
			}

			apiVersion := fieldType.Tag.Get("v")
			if apiVersion == "" {
				if version != "" {
					apiVersion = version
				} else {
					apiVersion = "1"
				}
			}

			// Build full path with versioning support
			prefix := ""
			if apiVersion != "" {
				prefix = "/v" + apiVersion
			}
			fullPath := prefix + strings.TrimSuffix(baseURL, "/") + apiURL
			if fullPath == "" {
				fullPath = "/"
			}

			// Get endpoint-level middlewares
			apiMiddleware := fieldType.Tag.Get("middleware")
			var endpointMiddlewares []Middleware
			if len(apiMiddleware) > 0 {
				names := strings.Split(apiMiddleware, ",")

				seenMiddlewares := make(map[string]bool)
				for _, mName := range moduleMiddlewareNames {
					seenMiddlewares[mName] = true
				}

				for _, middlewareName := range names {
					name := strings.Trim(middlewareName, " ")
					if name == "" {
						continue
					}

					if seenMiddlewares[name] {
						s.Logger.Info("Duplicate middleware registered for endpoint", "path", fullPath, "middleware", name)
					}
					seenMiddlewares[name] = true

					mw, ok := s.middleware[name]
					if !ok {
						s.Logger.Error(nil, "Middleware not registered", "name", name)
						continue
					}
					endpointMiddlewares = append(endpointMiddlewares, mw)
				}
			}

			handler, ok := checkAPIMethodExists(serviceValue, serviceType, fieldType)
			if !ok {
				s.Logger.Error(nil, "Handler not found", "name", fieldType.Name)
				continue
			}

			// Combine all middlewares: global + module + endpoint
			allMiddlewares := make([]Middleware, 0)
			allMiddlewares = append(allMiddlewares, s.globalMiddlewares...)
			allMiddlewares = append(allMiddlewares, moduleMiddlewares...)
			allMiddlewares = append(allMiddlewares, endpointMiddlewares...)

			// Wrap handler with all middlewares
			wrappedHandler := s.wrapWithMiddlewares(*handler, allMiddlewares)

			// Register the route with method-specific handling
			method := strings.ToUpper(fieldType.Type.String()[5:]) // Remove "neon." prefix
			s.registerRoute(method, fullPath, wrappedHandler, len(allMiddlewares))
		}
	}
}

// wrapWithMiddlewares applies middlewares in reverse order (outermost first)
func (s *App) wrapWithMiddlewares(handler http.HandlerFunc, middlewares []Middleware) http.HandlerFunc {
	wrapped := http.Handler(handler)

	// Apply middlewares in reverse order so they execute in correct order
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}

	return wrapped.ServeHTTP
}

// registerRoute registers a route with method checking
func (s *App) registerRoute(method, path string, handler http.HandlerFunc, totalMiddlewares int) {
	if s.routes == nil {
		s.routes = make(map[string]map[string]http.HandlerFunc)
	}
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}
	// Initialize path map if it doesn't exist
	if s.routes[path] == nil {
		s.routes[path] = make(map[string]http.HandlerFunc)

		// Create a dispatcher handler for this path
		dispatcher := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if methodHandler, exists := s.routes[path][r.Method]; exists {
				methodHandler(w, r)
			} else {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		})

		// Register the dispatcher with the mux
		s.mux.HandleFunc(path, dispatcher)
	}

	// Add the method handler to the path
	s.routes[path][method] = handler

	if s.Env == DevEnv {
		fmt.Printf("%s\t%s\t(%d middlewares)\n", blue(method), yellow(path), totalMiddlewares)
	}
}

func (s *App) Run(ctx context.Context) error {

	if s.Env == DevEnv {
		printLogo()
		printInfo(s)
	}

	s.Logger.Info("Server Starting", "port", s.Port, "Enviornment", s.Env)

	// Build all Endpoints after middleware registration
	// This ensures all changes(middlewares) after adding services are also included
	s.loadAllServices()

	// Execute Init() on all services
	for _, service := range s.services {
		if err := service.Init(); err != nil {
			return fmt.Errorf("module init failed: %w", err)
		}
	}

	// Register an automatic healthcheck endpoint for all services
	healthHandler := s.wrapWithMiddlewares(func(w http.ResponseWriter, r *http.Request) {
		for _, service := range s.services {
			if err := service.HealthCheck(); err != nil {
				http.Error(w, fmt.Sprintf("Service unhealthy: %v", err), http.StatusServiceUnavailable)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}, s.globalMiddlewares)
	s.registerRoute("GET", "/health", healthHandler, len(s.globalMiddlewares))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: s.mux,
	}

	serverErrors := make(chan error, 1)

	go func() {
		if s.Env == ProdEnv && s.Port == 443 && s.TLSCert != "" && s.TLSKey != "" {
			err := srv.ListenAndServeTLS(s.TLSCert, s.TLSKey)
			if err != nil && err != http.ErrServerClosed {
				s.Logger.Error(err, "TLS server error")
				serverErrors <- err
			}
		} else {
			err := srv.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				s.Logger.Error(err, "Server error")
				serverErrors <- err
			}
		}
	}()

	select {
	case err := <-serverErrors:
		return err

	case <-ctx.Done():
		s.Logger.Info("Server shutdown started")

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(timeoutCtx); err != nil {
			s.Logger.Error(err, "Server graceful shutdown failed")
			return err
		}

		// Close any services that implement Close(context.Context) error
		for _, service := range s.services {
			if err := service.Close(timeoutCtx); err != nil {
				s.Logger.Error(err, "Service close failed")
			}
		}
	}

	return nil
}

// Prepare builds all Endpoints without starting a server (useful for tests and benchmarks)
func (s *App) Prepare() {
	s.loadAllServices()
}

// ServeHTTP satisfies the standard http.Handler interface
func (s *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}
	s.mux.ServeHTTP(w, r)
}

// RequestLogger middleware
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// PanicRecovery middleware
func PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Recovered from panic: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
