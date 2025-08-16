# Neon HTTP Framework

[![Go Report Card](https://goreportcard.com/badge/github.com/sri-shubham/neon)](https://goreportcard.com/report/github.com/sri-shubham/neon)
[![GitHub issues](https://img.shields.io/github/issues/sri-shubham/neon)](https://github.com/sri-shubham/neon/issues)
[![GitHub stars](https://img.shields.io/github/stars/sri-shubham/neon)](https://github.com/sri-shubham/neon/stargazers)
[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org/dl/)
[![Version](https://img.shields.io/badge/version-0.1.0-green.svg)](CHANGELOG.md)

A lightweight, zero-dependency REST framework for Go that simplifies API development through struct tags and middleware composition.

> **⚠️ Development Status**: This framework is in active development (v0.1.0) and is **NOT production-ready**. The API may change significantly between versions until v1.0.0. Use at your own risk in production environments. The main branch may be unstable and is subject to force pushes and breaking changes without notice.

> **⚠️ Development Status**: This framework is currently in active development (v0.1.0). While functional, it is **not production-ready**. The main branch may contain breaking changes without notice as we work toward a stable v1.0.0 release. Use at your own risk in production environments.

## What's New in v0.1

- **Zero External Dependencies**: Removed chi router, now uses Go's built-in `net/http`
- **Named Path Parameters**: Support for `/users/{id}` and `/users/{id}/posts/{postId}`
- **Modern Go**: Requires Go 1.22+ with latest `http.ServeMux` features
- **Multi-Method Routes**: Same path supports multiple HTTP methods
- **Proper HTTP Status**: Real 404s and 405 Method Not Allowed responses
- **Logger Interface**: Bring your own logger with `logr.Logger` interface
- **Comprehensive Testing**: 100% test coverage with unit and integration tests

## Key Features

**Struct Tag Configuration:**
Define HTTP handlers using clean, readable struct tags that eliminate boilerplate code.

**Three-Level Middleware System:**
- **Global**: Applied to all endpoints
- **Service**: Applied to all endpoints in a service  
- **Endpoint**: Applied to specific endpoints only

**Named Path Parameters:**
Extract URL parameters easily using Go 1.22's `r.PathValue()`:
```go
userID := r.PathValue("id")        // From /users/{id}
postID := r.PathValue("postId")    // From /users/{id}/posts/{postId}
```

## Getting Started

### Prerequisites
- Go 1.22 or higher

### Installation
```bash
go get -u github.com/sri-shubham/neon
```

### Quick Start
```go
package main

import (
    "fmt"
    "net/http"
    "github.com/sri-shubham/neon"
)

type UserService struct {
    neon.Module `base:"/users" v:"1"`
    getUser     neon.Get `url:"/{id}"`
    createUser  neon.Post `url:"/"`
}

func (s UserService) GetUser(w http.ResponseWriter, r *http.Request) {
    userID := r.PathValue("id")
    w.Write([]byte(fmt.Sprintf("User ID: %s", userID)))
}

func (s UserService) CreateUser(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("User created"))
}

func main() {
    app := neon.New()
    app.AddService(&UserService{})
    
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

### Test Your API
```bash
curl http://localhost:8080/users/alice    # Returns: User ID: alice
curl -X POST http://localhost:8080/users/ # Returns: User created
```
       neon.Module `base:"/user" v:"1"`
       getUser     neon.Get
   }

   func (s UserService) GetUser(w http.ResponseWriter, r *http.Request) {
       fmt.Fprint(w, "Hello, Neon Magic!")
   }
   ```

4. **Run Your Magic Server:**
   ```bash
   go run test/main.go
   ```

5. **Explore the Glow:**
   Neon adds a glow to your routes:
   - Base route for UserService: `/user`
   - Version: `v1`
   - Route for `getUser`: `/user/v1`

## Join the Coding Wizardry! 🧙‍♂️

Ready to illuminate your API development with Neon? Join the coding wizardry and explore the full potential of Neon on [GitHub](https://github.com/sri-shubham/neon). Let the coding magic begin! 🚀✨
