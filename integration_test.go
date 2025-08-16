package neon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFullIntegration(t *testing.T) {
	app := New()

	// Track middleware execution order
	var executionOrder []string

	// Add global middleware
	globalMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "global")
			next.ServeHTTP(w, r)
		})
	}
	app.AddMiddleware(globalMW)

	// Register named middlewares
	authMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "auth")
			next.ServeHTTP(w, r)
		})
	}
	app.RegisterMiddleware("Auth", authMW)

	rateLimitMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "rateLimit")
			next.ServeHTTP(w, r)
		})
	}
	app.RegisterMiddleware("RateLimit", rateLimitMW)

	// Add test service
	service := &IntegrationTestService{}
	app.AddService(service)

	// Load services
	app.loadAllServices()

	// Test GET endpoint with service-level middleware
	t.Run("GET endpoint with service middleware", func(t *testing.T) {
		executionOrder = []string{} // Reset

		req := httptest.NewRequest("GET", "/integration/test", nil)
		w := httptest.NewRecorder()

		app.mux.ServeHTTP(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if !strings.Contains(w.Body.String(), "integration test response") {
			t.Errorf("Expected response to contain 'integration test response', got '%s'", w.Body.String())
		}

		// Check middleware execution order
		expectedOrder := []string{"global", "auth"}
		if len(executionOrder) != len(expectedOrder) {
			t.Errorf("Expected %d middleware executions, got %d: %v", len(expectedOrder), len(executionOrder), executionOrder)
		}

		for i, expected := range expectedOrder {
			if i < len(executionOrder) && executionOrder[i] != expected {
				t.Errorf("Expected middleware %s at position %d, got %s", expected, i, executionOrder[i])
			}
		}
	})

	// Test POST endpoint with service + endpoint middleware
	t.Run("POST endpoint with service and endpoint middleware", func(t *testing.T) {
		executionOrder = []string{} // Reset

		req := httptest.NewRequest("POST", "/integration/create", nil)
		w := httptest.NewRecorder()

		app.mux.ServeHTTP(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if !strings.Contains(w.Body.String(), "create response") {
			t.Errorf("Expected response to contain 'create response', got '%s'", w.Body.String())
		}

		// Check middleware execution order (should include endpoint-level RateLimit)
		expectedOrder := []string{"global", "auth", "rateLimit"}
		if len(executionOrder) != len(expectedOrder) {
			t.Errorf("Expected %d middleware executions, got %d: %v", len(expectedOrder), len(executionOrder), executionOrder)
		}

		for i, expected := range expectedOrder {
			if i < len(executionOrder) && executionOrder[i] != expected {
				t.Errorf("Expected middleware %s at position %d, got %s", expected, i, executionOrder[i])
			}
		}
	})

	// Test method not allowed
	t.Run("Method not allowed", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/integration/test", nil)
		w := httptest.NewRecorder()

		app.mux.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	// Test route not found
	t.Run("Route not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nonexistent", nil)
		w := httptest.NewRecorder()

		app.mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

// Integration test service
type IntegrationTestService struct {
	Module     `base:"/integration" v:"1" middleware:"Auth"`
	getTest    Get  `url:"/test" middleware:""`
	createTest Post `url:"/create" middleware:"RateLimit"`
}

func (s IntegrationTestService) GetTest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("integration test response"))
}

func (s IntegrationTestService) CreateTest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("create response"))
}

func TestServiceLoading(t *testing.T) {
	app := New()

	// Test multiple services
	service1 := &IntegrationTestService{}
	service2 := &TestModuleService{}

	app.AddService(service1)
	app.AddService(service2)

	if len(app.services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(app.services))
	}

	// Register middleware
	app.RegisterMiddleware("Auth", func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	})

	// Load services (should not panic)
	app.loadAllServices()

	// Test that routes were registered
	req := httptest.NewRequest("GET", "/integration/test", nil)
	w := httptest.NewRecorder()
	app.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for loaded service, got %d", w.Code)
	}
}
