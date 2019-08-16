package server

import (
	"encoding/csv"
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/RohitAwate/OAuth2Bin/oauth2/cache"
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
	setupRoutes(s)
	setupGracefulShutdown()

	log.Printf("OAuth 2.0 Server has started on port %s\n", s.Port)
	err := http.ListenAndServe(":"+s.Port, nil)
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server on port %s\n", s.Port)
	}
}

var osSignal chan os.Signal

func setupGracefulShutdown() {
	osSignal = make(chan os.Signal)
	signal.Notify(osSignal, syscall.SIGTERM)
	signal.Notify(osSignal, syscall.SIGINT)
	go onStopServer()
}

func onStopServer() {
	<-osSignal
	cache.ClosePool()
	os.Exit(0)
}

func setupRoutes(s *OA2Server) {
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

// Reads the server config from the specified path and returns it
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

	// Remove trailing "/" in the URL, if any
	if strings.HasSuffix(config.BaseURL, "/") {
		config.BaseURL = config.BaseURL[:len(config.BaseURL)-1]
	}

	return &config
}

// Reads the IP rate limiting policies from the specified file and
// returns them as an array, returns nil in case something goes wrong
func getRatePolicies(ratePoliciesPath string) []middleware.Policy {
	// Opens the file, reads the contents.
	fd, err := os.Open(ratePoliciesPath)
	if err != nil {
		log.Println("Could not read rate policies.")
		return nil
	}
	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Println("Could not read rate policies.")
		return nil
	}

	if len(data) <= 0 {
		return nil
	}

	// First, attempts to parse those contents as JSON.
	// If it works, returns the policies.
	policies, err := parseJSONPolicies(data)
	if err == nil {
		return policies
	}

	// Rewinding the file read pointer since the file may
	// already be consumed till the end by parseJSONPolicies.
	fd.Seek(0, io.SeekStart)

	// Attempts to parse the file as CSV
	policies, err = parseCSVPolicies(fd)
	if err != nil {
		log.Println("Unknown format for rate policies. JSON or CSV supported.")
		return nil
	}

	return policies
}

// Tries to parse the given data into an array of policies assuming that the format is JSON
func parseJSONPolicies(data []byte) ([]middleware.Policy, error) {
	var policies []middleware.Policy
	err := json.Unmarshal(data, &policies)
	if err != nil {
		return nil, err
	}

	return policies, nil
}

// Tries to parse the given data into an array of policies assuming that the format is CSV
func parseCSVPolicies(fd *os.File) ([]middleware.Policy, error) {
	lines, err := csv.NewReader(fd).ReadAll()
	if err != nil {
		return nil, err
	}

	policies := make([]middleware.Policy, len(lines))
	for i, line := range lines {
		limit, err := strconv.Atoi(strings.TrimSpace(line[1]))
		if err != nil {
			log.Fatalf("Expect integer value for policy rate limit: %s" + err.Error())
		}

		minutes, err := strconv.Atoi(strings.TrimSpace(line[2]))
		if err != nil {
			log.Fatalf("Expect integer value for policy time limit: %s" + err.Error())
		}

		policies[i] = middleware.Policy{
			Route:   strings.TrimSpace(line[0]),
			Limit:   limit,
			Minutes: minutes,
		}
	}

	return policies, nil
}
