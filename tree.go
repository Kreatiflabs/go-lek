package golek

import (
	"net/http"
	"strings"
)

// nodeType represents the type of a radix tree node
type nodeType uint8

const (
	ntStatic   nodeType = iota // /users
	ntParam                    // {id}
	ntCatchAll                 // *
)

// node represents a node in the radix tree
type node struct {
	prefix    string
	children  []*node
	handler   http.Handler
	nType     nodeType
	paramName string // for ntParam: the parameter name
	wildChild bool   // has a param/catchall child
	indices   string // first bytes of children prefixes for fast lookup
}

// addRoute adds a route pattern and handler to the tree.
// It supports static routes, parameterized routes ({param}), and catch-all routes (*).
func (n *node) addRoute(pattern string, handler http.Handler) {
	// A simple segment-based insertion adapted to the radix tree structure for robustness.
	// In a full production system this would be a heavily optimized byte-level radix tree.

	if len(pattern) == 0 {
		n.handler = handler
		return
	}

	// Handle Catch-All
	if pattern[0] == '*' {
		n.insertChild(pattern, handler, ntCatchAll)
		return
	}

	// Handle Parameter
	if pattern[0] == '{' {
		end := strings.IndexByte(pattern, '}')
		if end == -1 {
			panic("invalid route pattern: missing '}'")
		}
		paramName := pattern[1:end]
		rem := pattern[end+1:]

		child := n.insertChild(pattern[:end+1], nil, ntParam)
		child.paramName = paramName
		if len(rem) > 0 {
			child.addRoute(rem, handler)
		} else {
			child.handler = handler
		}
		return
	}

	// Handle Static
	// Find the next dynamic part
	nextDynamic := -1
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '{' || pattern[i] == '*' {
			nextDynamic = i
			break
		}
	}

	if nextDynamic == -1 {
		// Purely static
		n.insertChild(pattern, handler, ntStatic)
		return
	}

	// Mix of static and dynamic
	staticPart := pattern[:nextDynamic]
	rem := pattern[nextDynamic:]

	child := n.insertChild(staticPart, nil, ntStatic)
	child.addRoute(rem, handler)
}

// insertChild creates or finds an existing child node for the given prefix and type.
// It simplifies tree construction.
func (n *node) insertChild(prefix string, handler http.Handler, typ nodeType) *node {
	// Check for exact match in existing children
	for _, child := range n.children {
		if child.prefix == prefix && child.nType == typ {
			if handler != nil {
				child.handler = handler
			}
			return child
		}
	}

	// Create new child
	child := &node{
		prefix:  prefix,
		nType:   typ,
		handler: handler,
	}

	if typ == ntParam || typ == ntCatchAll {
		n.wildChild = true
	} else if len(prefix) > 0 {
		n.indices += string(prefix[0])
	}

	n.children = append(n.children, child)
	return child
}

// findRoute finds a handler matching the given path, returns handler and params
func (n *node) findRoute(path string, params func(key, value string)) (http.Handler, bool) {
	if len(path) == 0 {
		if n.handler != nil {
			return n.handler, true
		}
		return nil, false
	}

	// Fast path for static children using indices
	for _, child := range n.children {
		switch child.nType {
		case ntStatic:
			if strings.HasPrefix(path, child.prefix) {
				rem := path[len(child.prefix):]
				if handler, found := child.findRoute(rem, params); found {
					return handler, true
				}
			}
		case ntParam:
			// Parameter matching
			end := strings.IndexByte(path, '/')
			if end == -1 {
				end = len(path)
			}
			paramValue := path[:end]
			rem := path[end:]

			if handler, found := child.findRoute(rem, params); found {
				params(child.paramName, paramValue)
				return handler, true
			}
		case ntCatchAll:
			// Catch-all matching
			params("*", path) // Standardize catch-all param name, or let it be empty
			if child.handler != nil {
				return child.handler, true
			}
		}
	}

	// Check if the current node is a leaf matching an empty remainder
	if len(path) == 0 && n.handler != nil {
		return n.handler, true
	}

	return nil, false
}
