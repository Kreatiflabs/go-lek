package main

import (
	"fmt"
	"log"
	"net/http"

	golek "github.com/kreatiflabs/go-lek"
	"github.com/kreatiflabs/go-lek/middleware"
)

// ============================================================================
// Request & Response types
// ============================================================================

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age"   validate:"gte=0,lte=150"`
}

// UserResponse is the response for user endpoints.
type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// MessageResponse is a generic message response.
type MessageResponse struct {
	Message string `json:"message"`
}

// GetUserRequest binds path parameter.
type GetUserRequest struct {
	ID int `param:"id"`
}

// ============================================================================
// Handlers
// ============================================================================

// createUser handles POST /api/v1/users
// Demonstrates: auto-bind JSON body + auto-validate + typed response
func createUser(c *golek.Context, req *CreateUserRequest) (*UserResponse, error) {
	// req is already bound and validated automatically!
	return &UserResponse{
		ID:    1,
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}, nil
}

// getUser handles GET /api/v1/users/{id}
// Demonstrates: path parameter binding + error handling
func getUser(c *golek.Context, req *GetUserRequest) (*UserResponse, error) {
	if req.ID == 999 {
		return nil, golek.NotFound("user not found")
	}
	return &UserResponse{
		ID:    req.ID,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}, nil
}

// listUsers handles GET /api/v1/users
// Demonstrates: HNoReq (no request body for GET)
func listUsers(c *golek.Context) (*[]UserResponse, error) {
	users := []UserResponse{
		{ID: 1, Name: "John Doe", Email: "john@example.com", Age: 30},
		{ID: 2, Name: "Jane Doe", Email: "jane@example.com", Age: 25},
	}
	return &users, nil
}

func main() {
	// Create a new go-lek engine
	r := golek.New()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Golek-style handler with Wrap
	r.Get("/ping", golek.Wrap(func(c *golek.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"message":   "pong",
			"framework": "go-lek",
		})
	}))

	// API v1 routes with grouping
	r.Route("/api/v1", func(r *golek.Mux) {
		// Users CRUD
		r.Get("/users", golek.HNoReq(listUsers))
		r.Post("/users", golek.H(createUser))
		r.Get("/users/{id}", golek.H(getUser))
	})

	// OpenAPI spec endpoint
	r.Get("/openapi.json", golek.OpenAPIHandler(r, golek.OpenAPIInfo{
		Title:       "Go-Lek Example API",
		Description: "A sample API built with the go-lek framework",
		Version:     "1.0.0",
	}))

	// Start server with graceful shutdown
	addr := ":3000"
	fmt.Println("🔍 Go-Lek Example Server")
	fmt.Println("========================")
	fmt.Printf("Server:    http://localhost%s\n", addr)
	fmt.Printf("Health:    http://localhost%s/health\n", addr)
	fmt.Printf("API:       http://localhost%s/api/v1/users\n", addr)
	fmt.Printf("OpenAPI:   http://localhost%s/openapi.json\n", addr)
	fmt.Println()

	log.Fatal(r.ListenAndServe(addr))
}
