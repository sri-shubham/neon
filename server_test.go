package neon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
)

func TestNew(t *testing.T) {
	t.Run("New without config", func(t *testing.T) {
		app := New()

		if app == nil {
			t.Fatal("Expected app to be created, got nil")
		}

		if app.Port != 8080 {
			t.Errorf("Expected default port 8080, got %d", app.Port)
		}

		if app.mux == nil {
			t.Error("Expected mux to be initialized")
		}

		if app.middleware == nil {
			t.Error("Expected middleware map to be initialized")
		}

		if app.globalMiddlewares == nil {
			t.Error("Expected globalMiddlewares slice to be initialized")
		}
	})

	t.Run("New with config", func(t *testing.T) {
		config := &Config{
			Port:    9090,
			TLSCert: "cert.pem",
			TLSKey:  "key.pem",
		}

		app := New(config)

		if app.Port != 9090 {
			t.Errorf("Expected port 9090, got %d", app.Port)
		}

		if app.TLSCert != "cert.pem" {
			t.Errorf("Expected TLS cert 'cert.pem', got '%s'", app.TLSCert)
		}

		if app.TLSKey != "key.pem" {
			t.Errorf("Expected TLS key 'key.pem', got '%s'", app.TLSKey)
		}
	})

	t.Run("New with zero port in config", func(t *testing.T) {
		config := &Config{Port: 0}
		app := New(config)

		if app.Port != 8080 {
			t.Errorf("Expected default port 8080 when config port is 0, got %d", app.Port)
		}
	})
}

func TestSetLogger(t *testing.T) {
	app := New()

	// Create a test logger
	testLogger := logr.Discard()

	app.SetLogger(testLogger)

	// Since logr.Logger is not directly comparable, we test by using it
	// This ensures the logger was set correctly
	app.Logger.Info("test message")
}

func TestSetEnv(t *testing.T) {
	app := New()

	app.SetEnv(ProdEnv)
	if app.Env != ProdEnv {
		t.Errorf("Expected environment %v, got %v", ProdEnv, app.Env)
	}

	app.SetEnv(DevEnv)
	if app.Env != DevEnv {
		t.Errorf("Expected environment %v, got %v", DevEnv, app.Env)
	}
}

func TestAddMiddleware(t *testing.T) {
	app := New()

	testMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Test", "middleware")
			next.ServeHTTP(w, r)
		})
	}

	initialCount := len(app.globalMiddlewares)
	app.AddMiddleware(testMiddleware)

	if len(app.globalMiddlewares) != initialCount+1 {
		t.Errorf("Expected %d middlewares, got %d", initialCount+1, len(app.globalMiddlewares))
	}
}

func TestRegisterMiddleware(t *testing.T) {
	app := New()

	testMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	app.RegisterMiddleware("test", testMiddleware)

	if _, exists := app.middleware["test"]; !exists {
		t.Error("Expected middleware 'test' to be registered")
	}
}

func TestAddService(t *testing.T) {
	app := New()

	// Test valid service
	t.Run("Valid service", func(t *testing.T) {
		service := &TestService{}
		initialCount := len(app.services)

		app.AddService(service)

		if len(app.services) != initialCount+1 {
			t.Errorf("Expected %d services, got %d", initialCount+1, len(app.services))
		}
	})

	// Test invalid service (not a pointer)
	t.Run("Invalid service - not pointer", func(t *testing.T) {
		service := TestService{}
		initialCount := len(app.services)

		app.AddService(service) // This should fail

		if len(app.services) != initialCount {
			t.Error("Expected service count to remain the same for invalid service")
		}
	})
}

func TestWrapWithMiddlewares(t *testing.T) {
	app := New()

	// Create test middlewares that add headers
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Middleware-1", "executed")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Middleware-2", "executed")
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler executed"))
	})

	middlewares := []Middleware{middleware1, middleware2}
	wrappedHandler := app.wrapWithMiddlewares(handler, middlewares)

	// Test the wrapped handler
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrappedHandler(w, req)

	// Check that middlewares were executed
	if w.Header().Get("Middleware-1") != "executed" {
		t.Error("Expected Middleware-1 to be executed")
	}

	if w.Header().Get("Middleware-2") != "executed" {
		t.Error("Expected Middleware-2 to be executed")
	}

	// Check that handler was executed
	if w.Body.String() != "handler executed" {
		t.Errorf("Expected 'handler executed', got '%s'", w.Body.String())
	}
}

// Test service for testing purposes
type TestService struct {
	Module  `base:"/test" v:"1" middleware:""`
	getTest Get `url:"/endpoint" middleware:""`
}

func (s TestService) GetTest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test response"))
}

func TestEnvString(t *testing.T) {
	tests := []struct {
		env      Env
		expected string
	}{
		{DevEnv, "Development"},
		{TestEnv, "Test"},
		{ProdEnv, "Production"},
		{Env(999), "Not Defined"}, // Invalid env
	}

	for _, test := range tests {
		result := test.env.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s for env %v", test.expected, result, test.env)
		}
	}
}

func TestRegisterRoute(t *testing.T) {
	app := New()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	app.registerRoute("GET", "/test", handler)

	// Test that the route was registered
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	app.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test" {
		t.Errorf("Expected 'test', got '%s'", w.Body.String())
	}

	// Test wrong method
	req = httptest.NewRequest("POST", "/test", nil)
	w = httptest.NewRecorder()

	app.mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}
