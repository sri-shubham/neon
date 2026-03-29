package neon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

		req := httptest.NewRequest("GET", "/v1/integration/test", nil)
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

		req := httptest.NewRequest("POST", "/v1/integration/create", nil)
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
		req := httptest.NewRequest("PUT", "/v1/integration/test", nil)
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

	// Test Path Parameters
	t.Run("Path parameters", func(t *testing.T) {
		app := New()
		service := &PathParameterService{}
		app.AddService(service)
		app.loadAllServices()

		req := httptest.NewRequest("GET", "/v1/path/alice/posts/123", nil)
		w := httptest.NewRecorder()
		app.mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		expected := "User: alice, Post: 123"
		if w.Body.String() != expected {
			t.Errorf("Expected body '%s', got '%s'", expected, w.Body.String())
		}
	})

	// Test Versioning
	t.Run("Versioning", func(t *testing.T) {
		app := New()
		app.AddService(&VersionedServiceV1{})
		app.AddService(&VersionedServiceV2{})
		app.loadAllServices()

		// Test V1
		req1 := httptest.NewRequest("GET", "/v1/api/info", nil)
		w1 := httptest.NewRecorder()
		app.mux.ServeHTTP(w1, req1)
		if !strings.Contains(w1.Body.String(), "v1") {
			t.Errorf("Expected v1 response, got %s", w1.Body.String())
		}

		// Test V2
		req2 := httptest.NewRequest("GET", "/v2/api/info", nil)
		w2 := httptest.NewRecorder()
		app.mux.ServeHTTP(w2, req2)
		if !strings.Contains(w2.Body.String(), "v2") {
			t.Errorf("Expected v2 response, got %s", w2.Body.String())
		}
	})

	// Test Lifecycle Hooks
	t.Run("Lifecycle hooks", func(t *testing.T) {
		app := New()
		service := &LifecycleService{}
		app.AddService(service)

		// Run starts Init
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		if err := app.Run(ctx); err != nil {
			t.Fatalf("App run failed: %v", err)
		}

		if !service.initialized {
			t.Error("Service Init() was not called")
		}

		if !service.closed {
			t.Error("Service Close() was not called")
		}

		// Test HealthCheck
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		app.mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected health check 200, got %d", w.Code)
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

// Additional test services
type PathParameterService struct {
	Module  `base:"/path" v:"1"`
	getPost Get `url:"/{user}/posts/{id}"`
}

func (s PathParameterService) GetPost(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	id := r.PathValue("id")
	w.Write([]byte("User: " + user + ", Post: " + id))
}

type VersionedServiceV1 struct {
	Module  `base:"/api" v:"1"`
	getInfo Get `url:"/info"`
}

func (s VersionedServiceV1) GetInfo(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("v1 response"))
}

type VersionedServiceV2 struct {
	Module  `base:"/api" v:"2"`
	getInfo Get `url:"/info"`
}

func (s VersionedServiceV2) GetInfo(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("v2 response"))
}

type LifecycleService struct {
	Module      `base:"/lifecycle"`
	initialized bool
	closed      bool
}

func (s *LifecycleService) Init() error {
	s.initialized = true
	return nil
}

func (s *LifecycleService) Close(ctx context.Context) error {
	s.closed = true
	return nil
}

func (s *LifecycleService) HealthCheck() error {
	return nil
}

type IntegrationTestModuleService struct {
	Module  `base:"/test"`
	getTest Get `url:"/test"`
}

func (s IntegrationTestModuleService) GetTest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test response"))
}

func TestServiceLoading(t *testing.T) {
	app := New()

	// Test multiple services
	service1 := &IntegrationTestService{}
	service2 := &IntegrationTestModuleService{}

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
	req := httptest.NewRequest("GET", "/v1/integration/test", nil) // Fixed path with /v1
	w := httptest.NewRecorder()
	app.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for loaded service, got %d", w.Code)
	}
}
