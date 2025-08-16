# Neon Framework Roadmap

This document outlines the planned features and improvements for the Neon HTTP Framework.

## v0.2.0 - Enhanced Context & Auto-Marshaling

### Custom Context Wrapper
Introduce a unified context while maintaining backward compatibility with standard `http.Handler`:

**Option 1 - New Neon Context (Recommended):**
```go
func (s UserService) GetUser(ctx *neon.Context) {
    userID := ctx.PathValue("id")
    ctx.JSON(200, map[string]string{"id": userID})
}
```

**Option 2 - Traditional HTTP Handler (Still Supported):**
```go
func (s UserService) GetUser(w http.ResponseWriter, r *http.Request) {
    userID := r.PathValue("id")
    w.Write([]byte(fmt.Sprintf("User ID: %s", userID)))
}
```

**Option 3 - Mixed Approaches:**
```go
type UserService struct {
    neon.Module `base:"/users"`
    getUser     neon.Get `url:"/{id}"`     // Uses neon.Context
    legacy      neon.Get `url:"/legacy"`  // Uses http.Handler
}

func (s UserService) GetUser(ctx *neon.Context) {
    // New context-based handler
}

func (s UserService) Legacy(w http.ResponseWriter, r *http.Request) {
    // Traditional HTTP handler - still works!
}
```

### Framework Detection
The framework will automatically detect handler signatures and call them appropriately:
- `func(ctx *neon.Context)` - Uses new context wrapper
- `func(w http.ResponseWriter, r *http.Request)` - Uses traditional approach
- `func(ctx *neon.Context) ReturnType` - Auto-marshaling with context
- `func(w http.ResponseWriter, r *http.Request, params...)` - Traditional with dependency injection

### Implementation Details

**Handler Signature Detection:**
The framework will use reflection to detect handler signatures and parameter binding at registration time:

```go
// endpoint.go enhancement
func checkAPIMethodExists(sv reflect.Value, st reflect.Type, ft reflect.StructField) (*func(w http.ResponseWriter, r *http.Request), bool) {
    method := sv.MethodByName(handlerName)
    if !method.IsValid() {
        return nil, false
    }
    
    methodType := method.Type()
    
    // Detect handler signature
    switch {
    case isNeonContextWithParamsHandler(methodType):
        return wrapNeonContextWithParamsHandler(method, methodType), true
    case isNeonContextHandler(methodType):
        return wrapNeonContextHandler(method), true
    case isTraditionalHandler(methodType):
        return wrapTraditionalHandler(method), true
    default:
        return nil, false
    }
}
```

**Parameter Binding Implementation:**
```go
func wrapNeonContextWithParamsHandler(method reflect.Value, methodType reflect.Type) func(w http.ResponseWriter, r *http.Request) {
    // Analyze parameter types at registration time
    paramBindings := analyzeParameters(methodType)
    
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := neon.NewContext(w, r)
        
        // Build parameter values
        params := []reflect.Value{reflect.ValueOf(ctx)}
        
        for _, binding := range paramBindings {
            paramValue, err := binding.Bind(ctx)
            if err != nil {
                // Auto-handle binding errors
                ctx.Error(neon.NewBadRequestError(err.Error()))
                return
            }
            params = append(params, paramValue)
        }
        
        // Call handler with bound parameters
        results := method.Call(params)
        
        // Handle return values (auto-marshaling)
        handleResults(ctx, results)
    }
}

type ParameterBinding struct {
    Type        reflect.Type
    Source      BindingSource
    PathParam   string
    StructTags  map[string]string
}

type BindingSource int
const (
    SourcePath BindingSource = iota
    SourceQuery
    SourceBody
    SourceHeader
)

func (pb *ParameterBinding) Bind(ctx *neon.Context) (reflect.Value, error) {
    switch pb.Source {
    case SourcePath:
        return bindPathParameter(ctx, pb)
    case SourceQuery:
        return bindQueryParameters(ctx, pb)
    case SourceBody:
        return bindRequestBody(ctx, pb)
    case SourceHeader:
        return bindHeaders(ctx, pb)
    }
}
```

**Automatic Parameter Source Detection:**
```go
func analyzeParameters(methodType reflect.Type) []ParameterBinding {
    var bindings []ParameterBinding
    
    // Skip first parameter (always *neon.Context)
    for i := 1; i < methodType.NumIn(); i++ {
        paramType := methodType.In(i)
        binding := ParameterBinding{Type: paramType}
        
        switch {
        case isBasicType(paramType):
            // string, int, etc - assume path parameter
            binding.Source = SourcePath
            binding.PathParam = inferPathParamName(paramType)
        case hasQueryTags(paramType):
            // Struct with `query` tags
            binding.Source = SourceQuery
        case hasHeaderTags(paramType):
            // Struct with `header` tags
            binding.Source = SourceHeader
        default:
            // Default to request body binding
            binding.Source = SourceBody
        }
        
        bindings = append(bindings, binding)
    }
    
    return bindings
}
```

### Rich HTTP Error Types

**Built-in HTTP Error Types:**
```go
// Standard HTTP errors with auto-marshaling
func (s UserService) GetUser(ctx *neon.Context) (User, error) {
    userID := ctx.PathValue("id")
    
    if userID == "" {
        return User{}, neon.NewBadRequestError("User ID is required")
    }
    
    user, err := userRepo.FindByID(userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return User{}, neon.NewNotFoundError("User not found")
        }
        return User{}, neon.NewInternalServerError("Database error")
    }
    
    return user, nil
}
```

**Rich Error Structure:**
```go
type HTTPError struct {
    Status     int                    `json:"-"`
    Code       string                 `json:"code"`
    Message    string                 `json:"message"`
    Details    map[string]interface{} `json:"details,omitempty"`
    Timestamp  time.Time              `json:"timestamp"`
    RequestID  string                 `json:"request_id"`
}

// Built-in error constructors
func NewBadRequestError(message string) *HTTPError
func NewUnauthorizedError(message string) *HTTPError
func NewForbiddenError(message string) *HTTPError
func NewNotFoundError(message string) *HTTPError
func NewConflictError(message string) *HTTPError
func NewUnprocessableEntityError(message string, details map[string]interface{}) *HTTPError
func NewInternalServerError(message string) *HTTPError
func NewServiceUnavailableError(message string) *HTTPError
```

**Custom Error with Details:**
```go
func (s UserService) CreateUser(ctx *neon.Context) (User, error) {
    var user User
    if err := ctx.BindJSON(&user); err != nil {
        return User{}, neon.NewBadRequestError("Invalid JSON payload")
    }
    
    // Validation
    if user.Email == "" {
        return User{}, neon.NewUnprocessableEntityError("Validation failed", map[string]interface{}{
            "email": "Email is required",
        })
    }
    
    // Check for existing user
    if exists, _ := userRepo.ExistsByEmail(user.Email); exists {
        return User{}, neon.NewConflictError("User with this email already exists")
    }
    
    return userRepo.Create(user)
}
```

**Auto-Marshaled Error Response:**
```json
{
    "code": "VALIDATION_FAILED",
    "message": "Validation failed",
    "details": {
        "email": "Email is required"
    },
    "timestamp": "2025-08-16T13:45:30Z",
    "request_id": "req_123456789"
}
```

**Error Chaining Support:**
```go
func (s UserService) UpdateUser(ctx *neon.Context) (User, error) {
    // Wrap external errors with HTTP context
    user, err := userRepo.Update(userID, updateData)
    if err != nil {
        return User{}, neon.WrapError(err, 500, "USER_UPDATE_FAILED", "Failed to update user")
    }
    return user, nil
}
```

### Context Features
- **Path Parameters**: `ctx.PathValue("id")`
- **Query Parameters**: `ctx.Query("page")`, `ctx.QueryInt("limit")`
- **Request Body**: `ctx.BindJSON(&user)`, `ctx.BindXML(&user)`
- **Response Helpers**: `ctx.JSON(200, data)`, `ctx.XML(200, data)`, `ctx.String(200, "text")`
- **Headers**: `ctx.Header("Content-Type")`, `ctx.SetHeader("X-Custom", "value")`
- **Cookies**: `ctx.Cookie("session")`, `ctx.SetCookie(cookie)`

### Auto-Marshaling
Support multiple response formats with both handler types:

**Neon Context with Auto-Marshaling:**
```go
// Return struct - automatically marshaled to JSON
func (s UserService) GetUser(ctx *neon.Context) User {
    return User{ID: ctx.PathValue("id"), Name: "Alice"}
}

// Return with status code
func (s UserService) CreateUser(ctx *neon.Context) (User, int) {
    var user User
    ctx.BindJSON(&user)
    return user, 201
}

// Return error handling
func (s UserService) UpdateUser(ctx *neon.Context) (User, error) {
    var user User
    if err := ctx.BindJSON(&user); err != nil {
        return User{}, err
    }
    return userRepo.Save(user)
}
```

**Auto Request Body Binding (Concrete Parameters):**
```go
// Request body automatically bound to concrete parameter
func (s UserService) CreateUser(ctx *neon.Context, user CreateUserRequest) (User, error) {
    // 'user' parameter is automatically populated from request body
    // No need for manual binding!
    return userRepo.Create(user)
}

// Multiple parameters supported
func (s UserService) UpdateUser(ctx *neon.Context, userID string, updateReq UpdateUserRequest) (User, error) {
    // userID from path parameter, updateReq from request body
    return userRepo.Update(userID, updateReq)
}

// Query parameters also supported
func (s UserService) ListUsers(ctx *neon.Context, params ListUsersParams) ([]User, error) {
    // params automatically populated from query string
    return userRepo.List(params.Page, params.Limit, params.Sort)
}
```

**Parameter Binding Sources:**
```go
type ListUsersParams struct {
    Page  int    `query:"page" default:"1"`
    Limit int    `query:"limit" default:"10"`
    Sort  string `query:"sort" default:"name"`
}

type UpdateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"email"`
}

// Handler signature determines binding source
func (s UserService) ComplexHandler(
    ctx *neon.Context,           // Always first - framework context
    userID string,               // Path parameter (auto-detected from {userID})
    req UpdateUserRequest,       // Request body (struct = JSON binding)
    params ListUsersParams,      // Query parameters (has query tags)
) (User, error) {
    // All parameters automatically bound!
    return userRepo.UpdateWithParams(userID, req, params)
}
```

**Framework Parameter Detection:**
```go
// Framework automatically detects parameter sources:
// 1. *neon.Context - Framework context (always first)
// 2. string, int, etc with path parameter names - Path parameters
// 3. Structs with `json` tags - Request body binding
// 4. Structs with `query` tags - Query parameter binding
// 5. Structs with `header` tags - Header binding
```

**Traditional HTTP Handler (Manual Control):**
```go
func (s UserService) GetUserLegacy(w http.ResponseWriter, r *http.Request) {
    userID := r.PathValue("id")
    user := User{ID: userID, Name: "Alice"}
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

### Migration Strategy
- **Gradual Migration**: Convert handlers one by one from `http.Handler` to `neon.Context`
- **No Breaking Changes**: Existing code continues to work
- **Performance**: Both approaches have similar performance characteristics
- **Choice**: Developers can choose the style that fits their needs

## v0.3.0 - Request Validation & Error Handling

### Built-in Validation with Rich Errors
```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=18,max=120"`
}

// Option 1: Concrete parameter binding (automatic validation)
func (s UserService) CreateUser(ctx *neon.Context, req CreateUserRequest) (User, error) {
    // 'req' is automatically validated against struct tags
    // If validation fails, framework returns HTTPError automatically
    
    // Business logic validation
    if userRepo.ExistsByEmail(req.Email) {
        return User{}, neon.NewConflictError("Email already exists")
    }
    
    return userRepo.Create(req)
}

// Option 2: Manual binding with validation
func (s UserService) CreateUserManual(ctx *neon.Context) (User, error) {
    var req CreateUserRequest
    if err := ctx.BindJSON(&req); err != nil {
        // Auto-validates and returns structured error
        return User{}, err  // Framework converts validation errors to HTTPError
    }
    
    return userRepo.Create(req)
}
```

### Automatic Validation Error Formatting
When validation fails, the framework automatically returns:
```json
{
    "code": "VALIDATION_FAILED",
    "message": "Request validation failed",
    "details": {
        "name": "Name must be between 2 and 50 characters",
        "email": "Email format is invalid",
        "age": "Age must be at least 18"
    },
    "timestamp": "2025-08-16T13:45:30Z",
    "request_id": "req_123456789"
}
```

### Custom Error Handlers
```go
app.SetErrorHandler(func(err error, ctx *neon.Context) {
    switch e := err.(type) {
    case *neon.HTTPError:
        // Already formatted - just return
        ctx.JSON(e.Status, e)
    case *ValidationError:
        // Convert validation error to HTTPError
        httpErr := neon.NewUnprocessableEntityError("Validation failed", e.Details)
        ctx.JSON(httpErr.Status, httpErr)
    default:
        // Wrap unknown errors
        httpErr := neon.WrapError(err, 500, "INTERNAL_ERROR", "Internal server error")
        ctx.JSON(httpErr.Status, httpErr)
    }
})
```

### Error Middleware Integration
```go
// Automatic error recovery and logging
app.UseErrorRecovery(&neon.ErrorRecoveryConfig{
    LogErrors:        true,
    IncludeStackTrace: app.Env == neon.DevEnv,
    HideInternalErrors: app.Env == neon.ProdEnv,
})
```

## v0.4.0 - Advanced Routing

### Route Groups
```go
api := app.Group("/api/v1")
api.Use(authMiddleware)  // Apply to all routes in group

users := api.Group("/users")
users.AddService(&UserService{})

posts := api.Group("/posts")
posts.AddService(&PostService{})
```

### Route Constraints
```go
type UserService struct {
    neon.Module `base:"/users"`
    getUser     neon.Get `url:"/{id:int}"`     // Only numeric IDs
    getBySlug   neon.Get `url:"/{slug:alpha}"` // Only alphabetic
    getFiles    neon.Get `url:"/files/{path...}"` // Wildcard
}
```

### Subdomain Routing
```go
type APIService struct {
    neon.Module `subdomain:"api" base:"/v1"`
    getStatus   neon.Get `url:"/status"`
}
// Matches: api.example.com/v1/status
```

## v0.5.0 - Content Negotiation

### Accept Header Handling
```go
func (s UserService) GetUser(ctx *neon.Context) User {
    user := getUserFromDB(ctx.PathValue("id"))
    // Framework automatically responds with JSON, XML, or YAML
    // based on Accept header
    return user
}
```

### Multiple Response Formats
```go
type UserService struct {
    neon.Module `base:"/users"`
    getUser     neon.Get `url:"/{id}" produces:"application/json,application/xml,text/csv"`
}
```

### Custom Serializers
```go
app.RegisterSerializer("application/csv", &CSVSerializer{})
app.RegisterSerializer("application/protobuf", &ProtobufSerializer{})
```

## v0.6.0 - Observability & Monitoring

### Built-in Metrics
```go
app.EnableMetrics(&neon.MetricsConfig{
    Path: "/metrics",
    Prometheus: true,
})
// Exposes request count, duration, status codes
```

### Request Tracing
```go
app.EnableTracing(&neon.TracingConfig{
    ServiceName: "my-api",
    Endpoint:    "http://jaeger:14268/api/traces",
})
```

### Health Checks
```go
app.AddHealthCheck("database", func(ctx *neon.Context) error {
    return db.PingContext(ctx.Request().Context())
})

app.AddHealthCheck("redis", func(ctx *neon.Context) error {
    return redisClient.Ping(ctx.Request().Context()).Err()
})
// Exposes /health endpoint
```

## v0.7.0 - Security & Authentication

### Built-in Authentication
```go
app.UseJWT(&neon.JWTConfig{
    Secret:     "my-secret",
    TokenLookup: "header:Authorization,query:token,cookie:jwt",
})

type UserService struct {
    neon.Module `base:"/users"`
    getProfile  neon.Get `url:"/profile" auth:"required"`
    getPublic   neon.Get `url:"/public" auth:"optional"`
}
```

### Rate Limiting
```go
app.UseRateLimit(&neon.RateLimitConfig{
    Global: "1000/hour",
    PerIP:  "100/hour",
})

type UserService struct {
    getUser neon.Get `url:"/{id}" ratelimit:"50/minute"`
}
```

### CORS Configuration
```go
app.EnableCORS(&neon.CORSConfig{
    AllowedOrigins:   []string{"https://example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Authorization", "Content-Type"},
    AllowCredentials: true,
})
```

## v0.8.0 - Performance & Caching

### Response Caching
```go
type UserService struct {
    getUser  neon.Get `url:"/{id}" cache:"5m"`
    getUsers neon.Get `url:"/" cache:"1m" cache-key:"page,limit"`
}
```

### Compression
```go
app.EnableCompression(&neon.CompressionConfig{
    Level:     6,
    MinLength: 1024,
    Types:     []string{"application/json", "text/html"},
})
```

### Connection Pooling
```go
app.Configure(&neon.ServerConfig{
    ReadTimeout:       30 * time.Second,
    WriteTimeout:      30 * time.Second,
    IdleTimeout:       120 * time.Second,
    MaxHeaderBytes:    1 << 20,
    KeepAlivesEnabled: true,
})
```

## v0.9.0 - Developer Experience

### OpenAPI Generation
```go
type UserService struct {
    neon.Module `base:"/users" tags:"user-management"`
    getUser     neon.Get `url:"/{id}" summary:"Get user by ID" description:"Returns a single user"`
}

app.GenerateOpenAPI(&neon.OpenAPIConfig{
    Title:       "My API",
    Version:     "1.0.0",
    Description: "API for managing users",
    OutputPath:  "./docs/openapi.yaml",
})
```

### Development Tools
```go
if app.Env == neon.DevEnv {
    app.EnableHotReload()    // Auto-restart on file changes
    app.EnableDebugLogging() // Detailed request/response logs
    app.EnableProfiling()    // pprof endpoints
}
```

### Testing Utilities
```go
// In your tests
func TestUserAPI(t *testing.T) {
    app := neon.New()
    app.AddService(&UserService{})
    
    test := neon.NewTestClient(app)
    
    resp := test.GET("/users/123")
    assert.Equal(t, 200, resp.StatusCode)
    
    var user User
    resp.JSON(&user)
    assert.Equal(t, "123", user.ID)
}
```

## v1.0.0 - Advanced Features

### WebSocket Support
```go
type ChatService struct {
    neon.Module `base:"/chat"`
    connect     neon.WebSocket `url:"/ws"`
}

func (s ChatService) Connect(ws *neon.WebSocketContext) {
    for {
        var msg Message
        if err := ws.ReadJSON(&msg); err != nil {
            break
        }
        // Handle message
        ws.WriteJSON(response)
    }
}
```

### Server-Sent Events
```go
type EventService struct {
    events neon.SSE `url:"/events"`
}

func (s EventService) Events(sse *neon.SSEContext) {
    for event := range eventChannel {
        sse.Send("event", event)
    }
}
```

### Background Jobs
```go
app.AddJob("cleanup", neon.Daily, func(ctx *neon.JobContext) error {
    return cleanupOldData()
})

app.AddJob("sync", neon.Every(30*time.Minute), syncHandler)
```

## Implementation Priority

1. **v0.2.0** - Custom Context & Auto-Marshaling (High Priority)
2. **v0.3.0** - Validation & Error Handling (High Priority)
3. **v0.4.0** - Advanced Routing (Medium Priority)
4. **v0.5.0** - Content Negotiation (Medium Priority)
5. **v0.6.0** - Observability (Medium Priority)
6. **v0.7.0** - Security Features (High Priority)
7. **v0.8.0** - Performance (Low Priority)
8. **v0.9.0** - Developer Experience (Medium Priority)
9. **v1.0.0** - Advanced Features (Future Stable Release)

## Community Input

We welcome community feedback on this roadmap. Please open issues or discussions to:
- Suggest new features
- Propose changes to existing plans
- Vote on priority levels
- Contribute implementations
