package neon

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
)

type Env int

func (m Env) String() (out string) {
	switch m {
	case DevEnv:
		out = "Development"
	case TestEnv:
		out = "Test"
	case ProdEnv:
		out = "Production"
	default:
		out = "Not Defined"
	}
	return
}

const (
	DevEnv = iota
	TestEnv
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
		if len(moduleMiddleware) > 0 {
			middlewareNames := strings.Split(moduleMiddleware, ",")
			for _, middlewareName := range middlewareNames {
				name := strings.Trim(middlewareName, " ")
				mw, ok := s.middleware[name]
				if !ok {
					s.Logger.Error(nil, "Middleware not registered", "name", name)
					continue
				}
				moduleMiddlewares = append(moduleMiddlewares, mw)
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

			// Build full path
			fullPath := strings.TrimSuffix(baseURL, "/") + apiURL
			if fullPath == "" {
				fullPath = "/"
			}

			apiVersion := fieldType.Tag.Get("v")
			if apiVersion == "" {
				if version != "" {
					apiVersion = version
				} else {
					apiVersion = "1"
				}
			}

			// Get endpoint-level middlewares
			apiMiddleware := fieldType.Tag.Get("middleware")
			var endpointMiddlewares []Middleware
			if len(apiMiddleware) > 0 {
				names := strings.Split(apiMiddleware, ",")
				for _, middlewareName := range names {
					name := strings.Trim(middlewareName, " ")
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
			s.registerRoute(method, fullPath, wrappedHandler)
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
func (s *App) registerRoute(method, path string, handler http.HandlerFunc) {
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

	// Count total middlewares for this route
	totalMiddlewares := len(s.globalMiddlewares)
	fmt.Printf("%s\t%s\t(%d middlewares)\n", blue(method), yellow(path), totalMiddlewares)
}

func (s *App) Run() error {

	printLogo()
	printInfo(s)
	fmt.Println("Server Starting on Port:", blue(s.Port))

	// Add built-in middlewares to global middlewares
	s.globalMiddlewares = append([]Middleware{requestLogger, panicRecovery}, s.globalMiddlewares...)

	// Build all Endpoints after middleware registration
	// This ensures all changes(middlewares) after adding services are also included
	s.loadAllServices()

	if s.Env == ProdEnv && s.Port == 443 && s.TLSCert != "" && s.TLSKey != "" {
		err := http.ListenAndServeTLS(fmt.Sprintf(":%d", s.Port), s.TLSCert, s.TLSKey, s.mux)
		if err != nil {
			log.Printf("TLS server error: %v", err)
		}
		return err
	}
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.mux)
	if err != nil {
		log.Printf("Server error: %v", err)
	}
	return err
}

// Request logger middleware
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// Panic recovery middleware
func panicRecovery(next http.Handler) http.Handler {
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
