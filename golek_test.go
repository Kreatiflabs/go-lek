package golek

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestUser struct {
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=18"`
}

type TestResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestBasicRouting(t *testing.T) {
	app := New()

	app.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("world"))
	})

	req := httptest.NewRequest("GET", "/hello", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if rec.Body.String() != "world" {
		t.Fatalf("expected body 'world', got %s", rec.Body.String())
	}
}

func TestRouteParams(t *testing.T) {
	app := New()

	app.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := URLParam(r, "id")
		w.Write([]byte("user:" + id))
	})

	req := httptest.NewRequest("GET", "/users/42", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if rec.Body.String() != "user:42" {
		t.Fatalf("expected body 'user:42', got %s", rec.Body.String())
	}
}

func TestTypeSafeHandler_Success(t *testing.T) {
	app := New()

	app.Post("/users", H(func(c *Context, req *TestUser) (*TestResponse, error) {
		return &TestResponse{
			ID:   100,
			Name: req.Name,
		}, nil
	}))

	payload, _ := json.Marshal(TestUser{
		Name:  "Golek Dev",
		Email: "dev@golek.io",
		Age:   25,
	})

	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var res TestResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.ID != 100 || res.Name != "Golek Dev" {
		t.Fatalf("unexpected response data: %+v", res)
	}
}

func TestTypeSafeHandler_ValidationError(t *testing.T) {
	app := New()

	app.Post("/users", H(func(c *Context, req *TestUser) (*TestResponse, error) {
		return &TestResponse{ID: 1, Name: req.Name}, nil
	}))

	// Invalid email and age < 18
	payload, _ := json.Marshal(TestUser{
		Name:  "Go",
		Email: "invalid-email",
		Age:   10,
	})

	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", rec.Code)
	}
}

func TestRouteGrouping(t *testing.T) {
	app := New()

	app.Route("/api/v1", func(r *Mux) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("pong v1"))
		})
	})

	req := httptest.NewRequest("GET", "/api/v1/ping", nil)
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "pong v1" {
		t.Fatalf("expected 'pong v1', got %s", rec.Body.String())
	}
}

func BenchmarkRouterStatic(b *testing.B) {
	app := New()
	app.Get("/hello/world/benchmark", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/hello/world/benchmark", nil)
	rec := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app.ServeHTTP(rec, req)
	}
}

func BenchmarkRouterParam(b *testing.B) {
	app := New()
	app.Get("/users/{id}/profile", func(w http.ResponseWriter, r *http.Request) {
		_ = URLParam(r, "id")
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/users/12345/profile", nil)
	rec := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app.ServeHTTP(rec, req)
	}
}
