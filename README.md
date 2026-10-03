<p align="center">
  <img src="logo.jpg" alt="Go-Lek Logo" width="200">
</p>

<h1 align="center">Go-Lek</h1>

<p align="center">
  <strong>Go-Lek</strong> (Jawa: <em>golek</em> = mencari) — A modern, type-safe Go HTTP framework.
</p>
<p align="center">

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![CI](https://github.com/Kreatiflabs/go-lek/actions/workflows/ci.yml/badge.svg)](https://github.com/Kreatiflabs/go-lek/actions)
[![Zero Dependencies](https://img.shields.io/badge/dependencies-zero-brightgreen)]()
[![License](https://img.shields.io/badge/license-MIT-blue)]()

</p>

> **Better than Chi.** Type-safe handlers, auto-binding, auto-validation, OpenAPI generation — all with zero dependencies.

## ✨ Features

| Feature | Chi | Go-Lek |
|---|:---:|:---:|
| Radix tree router | ✅ | ✅ |
| Middleware chain | ✅ | ✅ |
| Route grouping | ✅ | ✅ |
| `net/http` compatible | ✅ | ✅ |
| **Type-safe generic handlers** | ❌ | ✅ |
| **Auto request binding** | ❌ | ✅ |
| **Auto struct validation** | ❌ | ✅ |
| **Auto OpenAPI 3.0 generation** | ❌ | ✅ |
| **Structured error handling** | ❌ | ✅ |
| **Built-in response helpers** | ❌ | ✅ |
| **Built-in graceful shutdown** | ❌ | ✅ |
| **Built-in middleware suite** | Partial | ✅ |

## 🚀 Quick Start

```bash
go get github.com/kreatiflabs/golek
```

```go
package main

import (
    "github.com/kreatiflabs/golek"
    "github.com/kreatiflabs/golek/middleware"
)

type CreateUserRequest struct {
    Name  string `json:"name"  validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

type UserResponse struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    r := golek.New()

    r.Use(middleware.Logger)
    r.Use(middleware.Recovery)

    // Type-safe handler — auto bind, validate, serialize!
    r.Post("/users", golek.H(func(c *golek.Context, req *CreateUserRequest) (*UserResponse, error) {
        return &UserResponse{ID: 1, Name: req.Name, Email: req.Email}, nil
    }))

    r.ListenAndServe(":3000")
}
```

## 📖 Handler Types

### Classic `net/http` (Chi-compatible)
```go
r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("pong"))
})
```

### Go-Lek Context Handler
```go
r.Get("/ping", golek.Wrap(func(c *golek.Context) {
    c.JSON(200, map[string]string{"message": "pong"})
}))
```

### 🔥 Type-Safe Generic Handler
```go
// Auto-bind request body + validate + return typed response
r.Post("/users", golek.H(func(c *golek.Context, req *CreateUserRequest) (*UserResponse, error) {
    // req is already bound and validated!
    user := createUser(req)
    return &UserResponse{ID: user.ID, Name: user.Name}, nil
}))

// GET handler (no request body)
r.Get("/users/{id}", golek.HNoReq(func(c *golek.Context) (*UserResponse, error) {
    id := c.Param("id")
    user, err := findUser(id)
    if err != nil {
        return nil, golek.NotFound("user not found")
    }
    return user, nil
}))

// DELETE handler (no response body → returns 204)
r.Delete("/users/{id}", golek.HNoRes(func(c *golek.Context, req *DeleteReq) error {
    return deleteUser(req.ID)
}))
```

## 📦 Auto Request Binding

Bind path params, query params, headers, and JSON body with struct tags:

```go
type Request struct {
    ID    int    `param:"id"`            // from /users/{id}
    Page  int    `query:"page"`          // from ?page=1
    Token string `header:"Authorization"` // from header
    Name  string `json:"name"`           // from JSON body
}
```

## ✅ Validation

Validate with struct tags:

```go
type CreateUserRequest struct {
    Name  string `json:"name"  validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age"   validate:"gte=0,lte=150"`
    Role  string `json:"role"  validate:"oneof=admin user guest"`
}
```

Supported rules: `required`, `email`, `min`, `max`, `gte`, `lte`, `len`, `oneof`

Invalid requests return structured errors:
```json
{
  "code": "VALIDATION_ERROR",
  "message": "request validation failed",
  "details": [
    {"field": "Email", "tag": "required", "message": "field is required"},
    {"field": "Age", "tag": "gte", "value": -1, "message": "value must be >= 0"}
  ]
}
```

## 🚨 Error Handling

```go
// Structured errors with HTTP status codes
return nil, golek.NotFound("user not found")
return nil, golek.BadRequest("invalid input")
return nil, golek.Unauthorized("invalid token")
return nil, golek.Forbidden("access denied")

// Chainable error configuration
return nil, golek.NewHTTPError(409, "email already exists").
    WithCode("DUPLICATE_EMAIL").
    WithDetails(map[string]string{"email": email})
```

## 🛡️ Middleware

Built-in middleware suite:

```go
r.Use(middleware.Logger)          // Request logging with colors
r.Use(middleware.Recovery)        // Panic recovery
r.Use(middleware.RequestID)       // X-Request-ID injection
r.Use(middleware.RealIP)          // Real IP detection
r.Use(middleware.CORS())          // CORS support
r.Use(middleware.RateLimit(100))  // 100 req/sec per IP
r.Use(middleware.Timeout(30*time.Second)) // Request timeout
r.Use(middleware.Compress(5))     // Gzip compression
r.Use(middleware.BasicAuth("api", creds))  // Basic auth
r.Use(middleware.BearerAuth(validateToken)) // Bearer auth
```

## 🛤️ Routing

```go
r := golek.New()

// Route grouping
r.Route("/api/v1", func(r *golek.Mux) {
    r.Use(middleware.RateLimit(100))

    r.Route("/users", func(r *golek.Mux) {
        r.Get("/", golek.HNoReq(listUsers))
        r.Post("/", golek.H(createUser))
        r.Get("/{id}", golek.H(getUser))
        r.Delete("/{id}", golek.HNoRes(deleteUser))
    })
})

// Mount sub-handler
r.Mount("/static", http.FileServer(http.Dir("./public")))
```

## 📄 OpenAPI Generation

```go
r.Get("/openapi.json", golek.OpenAPIHandler(r, golek.OpenAPIInfo{
    Title:   "My API",
    Version: "1.0.0",
}))
```

## 🏁 Graceful Shutdown

```go
// Handles SIGINT/SIGTERM automatically with 30s drain timeout
r.ListenAndServe(":3000")
r.ListenAndServeTLS(":443", "cert.pem", "key.pem")

// Or package-level for any http.Handler
golek.ListenAndServe(":3000", myHandler)
```

## 📁 Project Structure

```
golek/
├── golek.go          # Engine, New(), GetContext()
├── router.go         # Radix tree router (Mux)
├── tree.go           # Radix tree implementation
├── context.go        # Context with helpers
├── response.go       # ResponseWriter wrapper
├── handler.go        # H(), HNoReq(), HNoRes(), Wrap()
├── binder.go         # Auto request binding
├── validator.go      # Struct validation
├── errors.go         # HTTPError, ValidationError
├── openapi.go        # OpenAPI 3.0 generation
├── graceful.go       # Graceful shutdown
└── middleware/
    ├── logger.go     # Request logging
    ├── recovery.go   # Panic recovery
    ├── cors.go       # CORS
    ├── ratelimit.go  # Rate limiting
    ├── requestid.go  # Request ID
    ├── timeout.go    # Timeout
    ├── compress.go   # Gzip compression
    ├── auth.go       # Basic/Bearer auth
    └── realip.go     # Real IP detection
```

## 🤝 Credits

Built with ❤️ by [KreatifLabs](https://github.com/kreatiflabs)

## 📜 License

MIT License
