# Changelog

All notable changes to the Neon HTTP Framework will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-08-16

### Major Changes

#### Dependency Elimination
- **BREAKING**: Removed `github.com/go-chi/chi` dependency
- Implemented custom HTTP routing using Go's built-in `net/http` package
- Reduced external dependencies to only `fatih/color` and `go-logr/logr`

#### Modern Go Support
- **BREAKING**: Updated minimum Go version from 1.13 to 1.22
- Added support for `http.ServeMux` pattern matching with `{id}` syntax
- Implemented path parameter extraction using `r.PathValue()`

#### Enhanced Routing System
- Named path parameters support (e.g., `/users/{id}`, `/users/{id}/posts/{postId}`)
- Multiple HTTP methods on same path (GET and PUT on `/users/{id}`)
- Proper HTTP status codes (404 for non-existent routes, 405 for method not allowed)
- Fixed route pattern conflicts by implementing method-aware dispatching

### Added

- `logr.Logger` interface support for custom logging implementations
- `SetLogger()` method to configure logging
- Named middleware registration with `RegisterMiddleware()`
- Comprehensive unit test suite with 100% coverage
- Integration tests for middleware chain validation
- Internal route registry for better route management

### Fixed

- Route pattern conflicts in `http.ServeMux`
- Middleware execution order and counting
- Interface export naming (`moduler` → `Moduler`)
- Debug output contaminating response bodies
- Catch-all routes preventing proper 404 responses

### Removed

- **BREAKING**: Static file server (`/static/`)
- **BREAKING**: API documentation endpoint (`/docs`)
- **BREAKING**: Default welcome route (`/`)
- **BREAKING**: Chi-specific route patterns (`:id` → `{id}`)

### Migration Guide

#### Update Go Version
```go
// go.mod
go 1.22  // Previously: go 1.13
```

#### Update Route Patterns
```go
// Before (chi style)
getUser Get `url:"/:id"`

// After (http.ServeMux style)  
getUser Get `url:"/{id}"`
```

#### Extract Path Parameters
```go
// Before (chi)
userID := chi.URLParam(r, "id")

// After (built-in)
userID := r.PathValue("id")
```

---

## [1.x.x] - Previous Versions

Previous versions used chi router and are now deprecated. See git history for details.
