package middleware

import (
	"log"
	"net/http"
	"text/template"
)

// NotFoundMiddleware checks if the
type NotFoundMiddleware struct {
	URLPattern string
}

// NewNotFoundMiddleware returns a new instance of NotFoundMiddleware
func NewNotFoundMiddleware(pattern string) NotFoundMiddleware {
	return NotFoundMiddleware{URLPattern: pattern}
}

// Handle checks if the request's path matches URLPattern
func (nfm NotFoundMiddleware) Handle(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != nfm.URLPattern {
			// Serve the 404 page
			tmpl, err := template.ParseFiles(
				"public/templates/404.html",
				"public/templates/nav.html",
				"public/templates/footer.html",
			)
			if err != nil {
				log.Fatal(err)
			}

			err = tmpl.ExecuteTemplate(w, "404", nil)
			if err != nil {
				log.Fatal(err)
			}

			return
		}

		handler.ServeHTTP(w, r)
	}
}
