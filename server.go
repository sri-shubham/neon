package neon

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi"
)

type Env int

func (m Env) String() (out string) {
	switch m {
	case DevEnv:
		out = "Development"
	case TestEnv:
		out = "Test"
	case ProdEnv:
		out = "Produciton"
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
	Env  Env
	Port int

	mux        *chi.Mux
	services   []Moduler
	middleware map[string]Middleware
	// endpoints map[string]endpoint
}

func (s *App) SetEnv(e Env) {
	s.Env = e
}

// Config :
type Config struct {
}

type Middleware func(http.Handler) http.Handler

// New : Create a New Server
func New(conf ...*Config) *App {
	app := new(App)
	app.middleware = make(map[string]Middleware)
	app.mux = chi.NewRouter()
	return app
}

// Add a middleware for services
func (s *App) AddMiddleware(fun Middleware) {
	s.mux.Use(fun)
}

func (s *App) RegisterMiddleware(name string, fn Middleware) {
	s.middleware[name] = fn
}

// AddService : Add Service to app
// Service must embed neon.Module
func (s *App) AddService(servicePtr Moduler) {
	serviceTypeOf := reflect.TypeOf(servicePtr)
	if serviceTypeOf.Kind() != reflect.Ptr {
		log.Fatal("All services should be passed as a pointer")
	}

	serviceTypeOf = serviceTypeOf.Elem()
	if serviceTypeOf.Kind() != reflect.Struct {
		log.Fatal("service should be a struct")
	}

	module := reflect.TypeOf(Module{})
	field, ok := serviceTypeOf.FieldByName(module.Name())
	if !ok || !field.Anonymous || !field.Type.ConvertibleTo(module) {
		log.Fatal("service struct must add Module Anonymously")
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

		s.mux.Route(baseURL, func(r chi.Router) {
			middlewares := []func(http.Handler) http.Handler{}
			if len(moduleMiddleware) > 0 {
				moduleMiddlewares := strings.Split(moduleMiddleware, ",")
				for _, middleware := range moduleMiddlewares {
					name := strings.Trim(middleware, " ")
					if _, ok := s.middleware[name]; !ok {
						log.Fatal("Middleware not registered")
					}
					middlewares = append(middlewares, (func(http.Handler) http.Handler)(s.middleware[name]))
				}
				r.Use(middlewares...)
			}

			for i := 0; i < serviceType.NumField(); i++ {
				fieldType := serviceType.FieldByIndex([]int{i})
				// Check for Supported Methods
				switch fieldType.Type.String() {
				case "neon.Get", "neon.Put", "neon.Post", "neon.Patch", "neon.Delete", "neon.Options":
				default:
					continue
				}

				// Get API URL
				apiURL := fieldType.Tag.Get("url")
				if apiURL == "" {
					apiURL = "/"
				}

				// GET API Verion; If Provided Overwrite module level API verison
				// Only for current API
				version = fieldType.Tag.Get("v")
				if version == "" {
					version = "1"
				}

				// These middlewares run for specified modules only;
				// Note for this API Global Level middleware, Module Level Middleware and
				// API level middle ware all are executed
				apiMiddleware := fieldType.Tag.Get("middleware")

				handler, ok := checkAPIMethodExists(serviceValue, serviceType, fieldType)
				if !ok {
					log.Fatal("Either struct field starts with uppercase or Handler not found: ", red(fieldType.Name))
				}

				if len(apiMiddleware) > 0 {
					apiMiddlewares := strings.Split(apiMiddleware, ",")
					middlewares = []func(http.Handler) http.Handler{}
					for _, middleware := range apiMiddlewares {
						fmt.Println(middleware)
						name := strings.Trim(middleware, " ")
						if _, ok := s.middleware[name]; !ok {
							log.Fatal("Middleware not registered")
						}
						middlewares = append(middlewares, (func(http.Handler) http.Handler)(s.middleware[name]))
					}
				}

				switch fieldType.Type.String() {
				case "neon.Get":
					r.With(middlewares...).Get(apiURL, *handler)
				case "neon.Put":
					r.With(middlewares...).Put(apiURL, *handler)
				case "neon.Post":
					r.With(middlewares...).Post(apiURL, *handler)
				case "neon.Patch":
					r.With(middlewares...).Patch(apiURL, *handler)
				case "neon.Delete":
					r.With(middlewares...).Delete(apiURL, *handler)
				case "neon.Options":
					r.With(middlewares...).Options(apiURL, *handler)
				default:
					continue
				}

			}
		})
	}

}

func (s *App) Run() error {

	// Build all Endpoints before starting server
	// This ensures all changes(middlewares) after adding services are also included
	s.loadAllServices()

	printLogo()
	printInfo(s)
	fmt.Println("Server Starting on Port:", blue(s.Port))
	s.mux.Get("/", hello)

	if err := chi.Walk(s.mux, Walk); err != nil {
		panic(err)
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.mux)
}

func Walk(method string, route string, handler http.Handler,
	middlewares ...func(http.Handler) http.Handler) (err error) {
	fmt.Printf("%s\t%s\t(%d middlewares)\n", blue(method), yellow(route), len(middlewares))
	return
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("welcome"))
}
