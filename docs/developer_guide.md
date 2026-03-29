# Neon Developer Guide (v0.1.0)

Welcome to the Neon Developer Guide! This document provides a detailed explanation of the core features, patterns, and best practices for building robust REST services and tests using the Neon HTTP framework.

## Table of Contents
1. [Core Concepts](#core-concepts)
2. [Building Services](#building-services)
3. [REST API Best Practices](#rest-api-best-practices)
4. [Routing and Parameters](#routing-and-parameters)
5. [Middleware System](#middleware-system)
6. [Service Lifecycle](#service-lifecycle)
7. [API Versioning](#api-versioning)
8. [Observability and Logging](#observability-and-logging)
9. [Testing Your Services](#testing-your-services)

---

## Core Concepts

Neon is designed to be **declarative** and **lightweight**. It uses Go's reflection capabilities to map your struct definitions directly to HTTP endpoints. This approach reduces the "glue code" required to wire up handlers, allowing you to focus on your business logic.

- **App**: The central engine of your application. It manages route registration, middleware execution, and the server lifecycle.
- **Module**: The base unit of organization. Every service must embed `neon.Module`. This provides the necessary metadata (tags) for Neon to discover your endpoints.
- **Service**: A logical grouping of endpoints (e.g., `UserService`, `OrderService`). Services are registered with the `App` using `app.AddService()`.

---

## Building Services

To create a service, define a struct that embeds `neon.Module` anonymously.

```go
type OrderService struct {
    neon.Module `base:"/orders" v:"1" middleware:"auth"`
}
```

### Supported Endpoint Tags
Use the following tags to define your routes:
- `neon.Get`: For retrieval.
- `neon.Post`: For creation.
- `neon.Put`: For full updates.
- `neon.Patch`: For partial updates.
- `neon.Delete`: For removal.
- `neon.Options`: For CORS/metadata.

### Handler Resolution
Neon maps a field like `createOrder neon.Post` to a method named `CreateOrder`. Handlers must satisfy the `http.HandlerFunc` signature:
`func(w http.ResponseWriter, r *http.Request)`.

---

## Neon Framework Best Practices

### 1. Pointer Receivers for Services
Always use pointer receivers for your service methods if they need to maintain state (like a database connection or a cache). This is crucial for thread safety when combined with mutexes.

```go
// Good: Pointer receiver
func (s *UserService) GetUser(w http.ResponseWriter, r *http.Request) { ... }
```

### 2. Leverage `Init()` for Dependency Injection
Avoid manual initialization in `main`. Use the `Init()` method of your service struct to set up databases, clients, and internal state. This makes your services more self-contained and easier to test.

### 3. Middleware Granularity
- **Global**: Use only for concerns that apply to **every** request (e.g., Panic Recovery, Request ID injection).
- **Service-level**: Use for cross-cutting concerns within a module (e.g., Authentication, CORS).
- **Endpoint-level**: Use for specific requirements (e.g., Admin authorization, Rate limiting).

### 4. Group Related Endpoints
Keep your services focused. If a service handle too many unrelated resources, split it. Neon's base URL and versioning make it easy to manage multiple small services.

### 5. Graceful Shutdown is Not Optional
Always use `signal.NotifyContext` and pass it to `app.Run(ctx)`. Implement the `Close()` method in your services to ensure no data loss occurs when the server stops.

### 6. Centralized Error Handling Concepts
While Neon v0.1.0 uses standard `http.Error`, a best practice is to create a small helper within your service (or global) to ensure consistent error formats across all your endpoints.

```go
func (s *BaseService) JSONError(w http.ResponseWriter, code string, msg string, status int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": map[string]string{"code": code, "message": msg},
    })
}
```

### 7. Resource Management in `Init`
If you initialize multiple resources (e.g., DB and Redis), ensure you handle partial failures. If the second resource fails to initialize, clean up the first one before returning the error from `Init()`.

---

## REST API Best Practices

When building services with Neon, follow these industry-standard REST practices:

### 1. Use Plural Nouns for Resources
Avoid verbs in URLs. Use `/users` instead of `/getUsers`.
- **Good**: `GET /orders`, `POST /orders`
- **Bad**: `GET /get-orders`, `POST /create-order`

### 2. Standardize JSON Responses
Always wrap your data in a consistent structure, especially for errors.

**Success Response:**
```json
{
  "data": { "id": "123", "status": "shipped" }
}
```

**Error Response:**
```json
{
  "error": {
    "code": "ORDER_NOT_FOUND",
    "message": "The requested order does not exist."
  }
}
```

### 3. Return Correct HTTP Status Codes
- `200 OK`: Successful retrieval or update.
- `201 Created`: Successful creation (usually returns the new resource).
- `204 No Content`: Successful deletion (no body returned).
- `400 Bad Request`: Client-side input error (e.g., invalid JSON).
- `401 Unauthorized`: Authentication missing or failed.
- `403 Forbidden`: Authenticated but lack permission.
- `404 Not Found`: Resource does not exist.
- `500 Internal Server Error`: Unexpected server-side failure.

---

## Routing and Parameters

### Path Parameters
Named parameters are defined in the `url` tag: `url:"/{id}/items/{itemId}"`.
Extract them using Go 1.22's `r.PathValue(name)`:

```go
func (s *OrderService) GetItem(w http.ResponseWriter, r *http.Request) {
    orderID := r.PathValue("id")
    itemID := r.PathValue("itemId")
    // ...
}
```

---

## Middleware System

Neon's multi-level middleware allows for granular control. Middlewares are executed in order: **Global -> Service -> Endpoint**.

### Custom Middleware Example
```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token != "secret-key" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

Register it once: `app.RegisterMiddleware("auth", AuthMiddleware)`.

---

## Service Lifecycle

### `Init() error`
Use this for **synchronous** setup (e.g., pinging a DB, loading config). If `Init` returns an error, the server won't start.

### `Close(ctx context.Context) error`
Essential for **cleanup**. Ensure you close database connections, stop background goroutines, and flush logs.

---

## Testing Your Services

Testing Neon services is straightforward using the standard `net/http/httptest` package.

### Unit Testing a Handler
Since handlers are just standard Go methods, you can test them in isolation:

```go
func TestGetUser(t *testing.T) {
    service := &UserService{}
    service.Init() // Initialize memory storage

    req := httptest.NewRequest("GET", "/users/1", nil)
    req.SetPathValue("id", "1") // Manually set path value for the test
    w := httptest.NewRecorder()

    service.GetUser(w, req)

    resp := w.Result()
    if resp.StatusCode != http.StatusOK {
        t.Errorf("expected 200, got %d", resp.StatusCode)
    }
}
```

### Integration Testing with App
You can also test the fully assembled `App` using its `ServeHTTP` method:

```go
func TestAppIntegration(t *testing.T) {
    app := neon.New()
    app.AddService(&UserService{})
    app.Prepare() // Builds routes without starting server

    ts := httptest.NewServer(app)
    defer ts.Close()

    res, _ := http.Get(ts.URL + "/v1/users/1")
    if res.StatusCode != http.StatusOK {
        t.Fail()
    }
}
```
