package golek

import (
	"net/http"
)

// H wraps a typed handler function into a standard http.HandlerFunc.
// It automatically:
//   1. Binds the request body/params/query to the Req type
//   2. Validates the bound request struct
//   3. Calls the handler function
//   4. Serializes the response as JSON
//   5. Handles errors via the engine's ErrorHandler
func H[Req any, Res any](fn func(c *Context, req *Req) (*Res, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := GetContext(r)
		var req Req
		if err := c.Bind(&req); err != nil {
			c.engine.ErrorHandler(c, err)
			return
		}
		if err := c.Validate(&req); err != nil {
			c.engine.ErrorHandler(c, err)
			return
		}

		res, err := fn(c, &req)
		if err != nil {
			c.engine.ErrorHandler(c, err)
			return
		}

		if res != nil && !c.Writer.Written() {
			c.JSON(200, res)
		}
	}
}

// HNoReq wraps a handler that takes no request body but returns a typed response.
// Useful for GET handlers.
func HNoReq[Res any](fn func(c *Context) (*Res, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := GetContext(r)

		res, err := fn(c)
		if err != nil {
			c.engine.ErrorHandler(c, err)
			return
		}

		if res != nil && !c.Writer.Written() {
			c.JSON(200, res)
		}
	}
}

// HNoRes wraps a handler that takes a request but returns no response body.
// Returns 204 No Content on success.
func HNoRes[Req any](fn func(c *Context, req *Req) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := GetContext(r)
		var req Req
		if err := c.Bind(&req); err != nil {
			c.engine.ErrorHandler(c, err)
			return
		}
		if err := c.Validate(&req); err != nil {
			c.engine.ErrorHandler(c, err)
			return
		}

		if err := fn(c, &req); err != nil {
			c.engine.ErrorHandler(c, err)
			return
		}

		if !c.Writer.Written() {
			c.NoContent()
		}
	}
}

// Wrap converts a golek-style handler (func(*Context)) into http.HandlerFunc.
func Wrap(fn func(c *Context)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := GetContext(r)
		fn(c)
	}
}
