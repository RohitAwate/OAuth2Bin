package oauth2

import (
	"html/template"
	"log"
	"net/http"

	"github.com/RohitAwate/OAuth2Bin/oauth2/middleware"
)

// OA2Server implements an OAuth 2.0 server
type OA2Server struct {
	Port    string
	Config  OA2Config
	Limiter middleware.RateLimiter
}

var serverConfig OA2Config

// NewOA2Server returns a new OAuth 2.0 server which runs
// on the specified port with the specified configuration
func NewOA2Server(port string, config OA2Config) *OA2Server {
	serverConfig = config
	return &OA2Server{Port: port, Config: config}
}

// SetRateLimiter creates a new RateLimiter which enforces
// the policies passed.
func (s *OA2Server) SetRateLimiter(policies []middleware.Policy) {
	s.Limiter = middleware.RateLimiter{Policies: policies}
}

// Start sets up the static file server, handling routes and then starts listening for requests
func (s *OA2Server) Start() {
	public := http.FileServer(http.Dir("public/"))
	http.Handle("/public/", http.StripPrefix("/public/", public))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(
			"public/templates/index.html",
			"public/templates/base.html",
			"public/templates/cards.html",
		)
		if err != nil {
			log.Fatal(err)
		}

		err = tmpl.ExecuteTemplate(w, "home", s.Config)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/authorize", s.Limiter.CheckLimit(handleAuth))
	http.HandleFunc("/response", s.Limiter.CheckLimit(handleResponse))
	http.HandleFunc("/token", s.Limiter.CheckLimit(handleToken))
	log.Printf("OAuth 2.0 Server has started on port %s.\n", s.Port)
	http.ListenAndServe(":"+s.Port, nil)
}
