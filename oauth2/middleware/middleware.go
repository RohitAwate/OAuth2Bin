package middleware

import "net/http"

// Middleware represents middleware that may help
// in filtering HTTP requests.
type Middleware interface {
	Handle(handler http.HandlerFunc) http.HandlerFunc
}

// Chain makes it easy to chain a series of middleware to a handler.
// Reference: https://medium.com/@chrisgregory_83433/chaining-middleware-in-go-918cfbc5644d
func Chain(handler http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	if len(m) < 1 {
		return handler
	}

	wrapped := handler

	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i].Handle(wrapped)
	}

	return wrapped
}
