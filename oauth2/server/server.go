package server

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/RohitAwate/OAuth2Bin/oauth2/config"
	"github.com/RohitAwate/OAuth2Bin/oauth2/middleware"
)

// OA2Server implements an OAuth 2.0 server
type OA2Server struct {
	Port    string
	Config  config.OA2Config
	Limiter middleware.RateLimiter
}

var serverConfig config.OA2Config

// NewOA2Server returns a new OAuth 2.0 server which runs
// on the specified port with the specified configuration
func NewOA2Server(port string, serverConfigPath string, ratePoliciesPath string) *OA2Server {
	serverConfig = *getServerConfig(serverConfigPath)
	return &OA2Server{
		Port:   port,
		Config: serverConfig,
		Limiter: middleware.RateLimiter{
			Policies: getRatePolicies(ratePoliciesPath),
		},
	}
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

	s.route("/", s.handleHome)
	s.route("/authorize", handleAuth)
	s.route("/response", func(w http.ResponseWriter, r *http.Request) {
		pfv := middleware.PostFormValidator{
			Request:     r,
			VisualError: true,
		}

		pfv.Handle(handleResponse)(w, r)
	})
	s.route("/token", func(w http.ResponseWriter, r *http.Request) {
		pfv := middleware.PostFormValidator{
			Request:     r,
			VisualError: false,
		}

		pfv.Handle(handleToken)(w, r)
	})
	s.route("/echo", handleEcho)

	log.Printf("OAuth 2.0 Server has started on port %s.\n", s.Port)
	http.ListenAndServe("0.0.0.0:"+s.Port, nil)
}

// Registers a callback for the specified URL pattern.
// The callback first examines if request route matches the pattern.
// If not, the 404 logic is triggered. This is the purpose behind the existence
// of this method.
//
// However, if the paths match, the IP limiter is triggered which returns another callback.
// This callback is called by passing it the RequestWriter and the Request pointer.
func (s *OA2Server) route(pattern string, handler http.HandlerFunc) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != pattern {
			handleNotFound(w, r)
			return
		}

		s.Limiter.Handle(handler)(w, r)
	})
}

// Serves the home page
func (s *OA2Server) handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"public/templates/index.html",
		"public/templates/nav.html",
		"public/templates/cards.html",
		"public/templates/footer.html",
	)
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.ExecuteTemplate(w, "home", s.Config)
	if err != nil {
		log.Fatal(err)
	}
}

// Serves the 404 page
func handleNotFound(w http.ResponseWriter, r *http.Request) {
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
}

func getServerConfig(serverConfigPath string) *config.OA2Config {
	fd, err := os.Open(serverConfigPath)
	if err != nil {
		return nil
	}
	defer fd.Close()

	jsonBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatal(err)
	}

	if len(jsonBytes) <= 0 {
		return nil
	}

	var config config.OA2Config
	err = json.Unmarshal(jsonBytes, &config)
	if err != nil {
		log.Fatal(err)
	}

	return &config
}

func getRatePolicies(ratePoliciesPath string) []middleware.Policy {
	fd, err := os.Open(ratePoliciesPath)
	if err != nil {
		return nil
	}
	defer fd.Close()

	jsonBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatal(err)
	}

	if len(jsonBytes) <= 0 {
		return nil
	}

	var policies []middleware.Policy
	err = json.Unmarshal(jsonBytes, &policies)
	if err != nil {
		log.Fatal(err)
	}

	return policies
}
