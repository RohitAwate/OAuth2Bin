package oauth2

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
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
	http.HandleFunc("/accepted", handleAccepted)
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

	handler.HandleGrant(w, r)
}

func handleAccepted(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		Error(w, r, 400, ErrorTemplate{Title: "Bad Request", Desc: r.Method + " not allowed."})
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		Error(w, r, 400, ErrorTemplate{Title: "Bad Request", Desc: "Could not read body of request."})
		return
	}

	var msg map[string]string
	err = json.Unmarshal(body, &msg)
	if err != nil {
		Error(w, r, 400, ErrorTemplate{Title: "Bad Request", Desc: "Invalid JSON found in request body."})
		return
	}

	redirectURI := msg["redirect_uri"] + "?code=" + serverConfig.AuthCodeCnfg.AuthGrant
	http.Redirect(w, r, redirectURI, http.StatusSeeOther)
}
