package middleware

import "net/http"

// Middleware represents middleware that may help
// in filtering HTTP requests.
type Middleware interface {
	Handle(handler http.HandlerFunc) http.HandlerFunc
}
