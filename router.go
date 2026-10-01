package golek

import (
	"context"
	"net/http"
	"reflect"
	"strings"
)

// Mux is the HTTP request multiplexer / router.
// It dispatches requests to registered handlers using a radix tree
// and supports middleware chaining, route grouping, and sub-routing.
type Mux struct {
	tree        map[string]*node
	middlewares []func(http.Handler) http.Handler

	// Route metadata for OpenAPI generation
	routes []RouteInfo

	// Inline mux (for Group)
	parent  *Mux
	inline  bool
	pattern string

	notFound         http.HandlerFunc
	methodNotAllowed http.HandlerFunc
}

// RouteInfo stores metadata about a registered route for documentation/OpenAPI.
type RouteInfo struct {
	Method       string
	Pattern      string
	HandlerName  string
	Summary      string
	Description  string
	Tags         []string
	Deprecated   bool
	RequestType  reflect.Type
	ResponseType reflect.Type
}

// NewMux creates a new Mux router.
func NewMux() *Mux {
	return &Mux{
		tree:        make(map[string]*node),
		middlewares: make([]func(http.Handler) http.Handler, 0),
		routes:      make([]RouteInfo, 0),
	}
}

// ServeHTTP dispatches the request to the handler whose pattern matches.
func (mx *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path

	root, ok := mx.tree[method]
	var handler http.Handler

	if ok {
		var routeParams Params
		extractParam := func(key, value string) {
			routeParams = append(routeParams, Param{Key: key, Value: value})
		}

		if h, found := root.findRoute(path, extractParam); found {
			handler = h
			// Inject params into context if they exist
			if len(routeParams) > 0 {
				ctx := context.WithValue(r.Context(), ParamsCtxKey, routeParams)
				r = r.WithContext(ctx)
			}
		}
	}

	if handler == nil {
		// Check if it's a 405 Method Not Allowed
		allowed := mx.allowedMethods(path)
		if len(allowed) > 0 {
			w.Header().Set("Allow", strings.Join(allowed, ", "))
			if mx.methodNotAllowed != nil {
				handler = mx.methodNotAllowed
			} else {
				handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				})
			}
		} else {
			// 404 Not Found
			if mx.notFound != nil {
				handler = mx.notFound
			} else {
				handler = http.NotFoundHandler()
			}
		}
	}

	// Apply middlewares in reverse order
	for i := len(mx.middlewares) - 1; i >= 0; i-- {
		handler = mx.middlewares[i](handler)
	}

	handler.ServeHTTP(w, r)
}

func (mx *Mux) allowedMethods(path string) []string {
	var allowed []string
	for method, root := range mx.tree {
		if method == http.MethodOptions {
			continue
		}
		if _, found := root.findRoute(path, func(key, value string) {}); found {
			allowed = append(allowed, method)
		}
	}
	if len(allowed) > 0 {
		allowed = append(allowed, http.MethodOptions)
	}
	return allowed
}

// Use appends one or more middlewares to the middleware stack.
func (mx *Mux) Use(middlewares ...func(http.Handler) http.Handler) {
	mx.middlewares = append(mx.middlewares, middlewares...)
}

// With adds inline middlewares and returns a new inline Mux.
func (mx *Mux) With(middlewares ...func(http.Handler) http.Handler) *Mux {
	m := &Mux{
		tree:             mx.tree,
		middlewares:      append(mx.middlewares[:len(mx.middlewares):len(mx.middlewares)], middlewares...),
		parent:           mx,
		inline:           true,
		pattern:          mx.pattern,
		notFound:         mx.notFound,
		methodNotAllowed: mx.methodNotAllowed,
	}
	return m
}

// Group creates a new inline Mux with a fresh middleware stack.
func (mx *Mux) Group(fn func(r *Mux)) *Mux {
	m := mx.With()
	if fn != nil {
		fn(m)
	}
	return m
}

// Route creates a sub-router with the given pattern prefix.
func (mx *Mux) Route(pattern string, fn func(r *Mux)) *Mux {
	m := &Mux{
		tree:             mx.tree,
		middlewares:      mx.middlewares[:len(mx.middlewares):len(mx.middlewares)],
		parent:           mx,
		inline:           true,
		pattern:          mx.pattern + pattern,
		notFound:         mx.notFound,
		methodNotAllowed: mx.methodNotAllowed,
	}
	if fn != nil {
		fn(m)
	}
	return m
}

// Mount attaches an http.Handler at the given pattern.
func (mx *Mux) Mount(pattern string, handler http.Handler) {
	// Simple approach: register it as a catch-all if it ends with /
	// Strip prefix logic should ideally wrap the handler
	if !strings.HasSuffix(pattern, "/*") {
		pattern += "/*"
	}

	h := http.StripPrefix(strings.TrimSuffix(pattern, "/*"), handler)
	mx.Handle(pattern, h)
}

// Handle registers a handler for the given pattern (all methods).
func (mx *Mux) Handle(pattern string, handler http.Handler) {
	methods := []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete,
		http.MethodPatch, http.MethodHead, http.MethodOptions,
	}
	for _, method := range methods {
		mx.Method(method, pattern, handler)
	}
}

// Method registers a handler for a specific HTTP method and pattern.
func (mx *Mux) Method(method, pattern string, handler http.Handler) {
	if mx.tree[method] == nil {
		mx.tree[method] = &node{}
	}

	fullPattern := mx.pattern + pattern

	// Pre-wrap with inline middlewares if any
	if mx.inline && len(mx.middlewares) > 0 {
		h := handler
		// We only want to apply the difference in middlewares from the root if we were maintaining a separate tree,
		// but since we share the tree and compile at ServeHTTP time, actually inline middleware needs to be baked into the handler.
		// Wait, ServeHTTP applies mx.middlewares. But we are inserting into a shared tree!
		// So we MUST bake the current Mux's middlewares into the handler, because ServeHTTP uses the root Mux's middlewares.
		// Let's reconsider: The root Mux's ServeHTTP will apply root middlewares.
		// We should bake ALL current middlewares into the handler, and during root ServeHTTP it applies... wait.
		// If we bake them here, we shouldn't apply them again in root.
		// Actually, in Chi, handlers in the tree are fully wrapped.
		// So we bake all middlewares here.
		for i := len(mx.middlewares) - 1; i >= 0; i-- {
			h = mx.middlewares[i](h)
		}
		handler = h
	}

	mx.tree[method].addRoute(fullPattern, handler)

	// Save route info
	// We need to resolve the root mux to save the route info globally if needed, or just locally.
	// We'll save it locally.
	mx.routes = append(mx.routes, RouteInfo{
		Method:  method,
		Pattern: fullPattern,
	})

	if mx.parent != nil {
		mx.parent.routes = append(mx.parent.routes, mx.routes[len(mx.routes)-1])
	}
}

// Get registers a GET route.
func (mx *Mux) Get(pattern string, handlerFn http.HandlerFunc) {
	mx.Method(http.MethodGet, pattern, handlerFn)
}

// Post registers a POST route.
func (mx *Mux) Post(pattern string, handlerFn http.HandlerFunc) {
	mx.Method(http.MethodPost, pattern, handlerFn)
}

// Put registers a PUT route.
func (mx *Mux) Put(pattern string, handlerFn http.HandlerFunc) {
	mx.Method(http.MethodPut, pattern, handlerFn)
}

// Delete registers a DELETE route.
func (mx *Mux) Delete(pattern string, handlerFn http.HandlerFunc) {
	mx.Method(http.MethodDelete, pattern, handlerFn)
}

// Patch registers a PATCH route.
func (mx *Mux) Patch(pattern string, handlerFn http.HandlerFunc) {
	mx.Method(http.MethodPatch, pattern, handlerFn)
}

// Head registers a HEAD route.
func (mx *Mux) Head(pattern string, handlerFn http.HandlerFunc) {
	mx.Method(http.MethodHead, pattern, handlerFn)
}

// Options registers an OPTIONS route.
func (mx *Mux) Options(pattern string, handlerFn http.HandlerFunc) {
	mx.Method(http.MethodOptions, pattern, handlerFn)
}

// NotFound sets a custom handler for 404 responses.
func (mx *Mux) NotFound(handlerFn http.HandlerFunc) {
	mx.notFound = handlerFn
}

// MethodNotAllowed sets a custom handler for 405 responses.
func (mx *Mux) MethodNotAllowed(handlerFn http.HandlerFunc) {
	mx.methodNotAllowed = handlerFn
}

// Routes returns all registered route info for documentation.
func (mx *Mux) Routes() []RouteInfo {
	return mx.routes
}

// Note: Param type is defined in context.go
