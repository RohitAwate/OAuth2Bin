package oauth2

import (
	"html/template"
	"log"
	"net/http"
)

// OA2Server implements an OAuth 2.0 server
type OA2Server struct {
	Port   string
	Config OA2Config
}

var serverConfig OA2Config

// NewOA2Server returns a new OAuth 2.0 server which runs
// on the specified port with the specified configuration
func NewOA2Server(port string, config OA2Config) *OA2Server {
	serverConfig = config
	return &OA2Server{Port: port, Config: config}
}

// Start sets up the static file server, handling routes and then starts listening for requests
func (s *OA2Server) Start() {
	public := http.FileServer(http.Dir("public/"))
	http.Handle("/public/", http.StripPrefix("/public/", public))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		template, err := template.ParseFiles("public/templates/index.html")
		if err != nil {
			panic(err)
		}

		template.Execute(w, s.Config)
	})

	http.HandleFunc("/authorize", handleAuth)
	http.HandleFunc("/response", handleResponse)
	http.HandleFunc("/token", handleToken)
	log.Printf("OAuth 2.0 Server has started on port %s.\n", s.Port)
	http.ListenAndServe(":"+s.Port, nil)
}
