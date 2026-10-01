package golek

import (
	"context"
	"net/http"
	"sync"
)

// contextKey is a private type for context keys
type contextKey struct{ name string }

var (
	// ParamsCtxKey is used to store URL parameters in the request context.
	ParamsCtxKey = &contextKey{"GolekParams"}
	// engineCtxKey is used to store the golek Context in the request context.
	engineCtxKey = &contextKey{"GolekContext"}
)

// ErrorHandlerFunc defines the error handler signature.
type ErrorHandlerFunc func(c *Context, err error)

// Engine is the main go-lek framework instance.
// It embeds *Mux for routing and adds binding, validation,
// error handling, and other framework features.
type Engine struct {
	*Mux

	// Binder is the request binder (default: DefaultBinder)
	Binder Binder

	// Validator is the request validator (default: DefaultValidator)
	Validator Validator

	// ErrorHandler handles errors returned from typed handlers.
	ErrorHandler ErrorHandlerFunc

	// contextPool for reusing Context objects
	pool sync.Pool
}

// New creates a new Engine with default configuration.
func New() *Engine {
	engine := &Engine{
		Binder:       &DefaultBinder{},
		Validator:    &DefaultValidator{},
		ErrorHandler: defaultErrorHandler,
	}
	engine.Mux = NewMux()

	engine.pool.New = func() any {
		return &Context{
			engine: engine,
		}
	}

	return engine
}

// ServeHTTP implements http.Handler. Creates/reuses Context for each request.
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get Context from pool
	c := e.pool.Get().(*Context)

	// Reset Context state with wrapped ResponseWriter
	c.reset(w, r)

	// Store the Context in the request's context for GetContext()
	ctx := context.WithValue(r.Context(), engineCtxKey, c)
	c.Request = r.WithContext(ctx)

	// Serve the request using the embedded Mux
	// The Mux will set params in the request context; we need to intercept that.
	e.Mux.ServeHTTP(c.Writer, c.Request)

	// Return to pool
	e.pool.Put(c)
}

// URLParam returns the URL parameter from a request's context.
// This is the primary way to access route parameters from standard http.HandlerFunc.
func URLParam(r *http.Request, key string) string {
	if params, ok := r.Context().Value(ParamsCtxKey).(Params); ok {
		return params.Get(key)
	}
	return ""
}

// GetContext returns the golek Context from the request.
// It retrieves the Context that was stored by Engine.ServeHTTP.
func GetContext(r *http.Request) *Context {
	if c, ok := r.Context().Value(engineCtxKey).(*Context); ok {
		// Sync the request in case middleware modified it
		c.Request = r
		return c
	}
	return nil
}
