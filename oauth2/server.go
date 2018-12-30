package oauth2

import (
	"html/template"
	"net/http"
)

// OA2Server implements an OAuth 2.0 server
type OA2Server struct {
	Port   int
	Config OA2Config
}

var serverConfig OA2Config

// NewOA2Server returns a new OAuth 2.0 server which runs
// on the specified port with the specified configuration
func NewOA2Server(port int, config OA2Config) *OA2Server {
	serverConfig = config
	return &OA2Server{Port: port, Config: config}
}

// Start listening for requests
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
	http.ListenAndServe(":8080", nil)
}

// Routes the request to a FlowHandler based on the request_type
func handleAuth(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var handler FlowHandler
	switch queryParams.Get("response_type") {
	case "code":
		handler = &AuthCodeHandler{Config: &(serverConfig.AuthCodeCnfg)}
	default:
		Error(w, r, 400, ErrorTemplate{Title: "Bad Request", Desc: "response_type is required"})
		return
	}

	handler.Handle(w, r)
}
