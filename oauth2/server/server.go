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
func (s *OA2Server) SetRateLimiter(policies []middleware.RatePolicy) {
	s.Limiter = middleware.RateLimiter{Policies: policies}
}

// Start sets up the static file server, handling routes and then starts listening for requests
func (s *OA2Server) Start() {
	s.setupRoutes()
	setupGracefulShutdown()

	log.Printf("OAuth 2.0 Server has started on port %s\n", s.Port)
	err := http.ListenAndServe(":"+s.Port, nil)
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server on port %s\n", s.Port)
	}
}

func (s *OA2Server) chainCommonMiddleware(pattern string, handler http.HandlerFunc, extras ...middleware.Middleware) {
	middlewareSlice := []middleware.Middleware{s.Limiter, middleware.NewNotFoundMiddleware(pattern)}
	middlewareSlice = append(middlewareSlice, extras...)
	chain := middleware.Chain(handler, middlewareSlice...)
	http.HandleFunc(pattern, chain)
}

func (s *OA2Server) setupRoutes() {
	public := http.FileServer(http.Dir("public/"))
	http.Handle("/public/", http.StripPrefix("/public/", public))

	s.chainCommonMiddleware("/", s.handleHome)
	s.chainCommonMiddleware("/authorize", handleAuth)
	s.chainCommonMiddleware("/response", handleResponse, middleware.NewPostFormValidator(true))
	s.chainCommonMiddleware("/token", handleToken, middleware.NewPostFormValidator(false))
	s.chainCommonMiddleware("/echo", handleEcho)
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

// Channel over which we receive signals from the operating system
var osSignal chan os.Signal

// Configures what OS signals we wish to be notified about.
// Fires up a goroutine which listens for these signals and
// runs the shutdown logic when they're received.
func setupGracefulShutdown() {
	osSignal = make(chan os.Signal)
	signal.Notify(osSignal, syscall.SIGTERM)
	signal.Notify(osSignal, syscall.SIGINT)
	go onStopServer()
}

// Listens for OS signals and executes the shutdown logic
func onStopServer() {
	<-osSignal
	cache.ClosePool()
	os.Exit(0)
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
func getRatePolicies(ratePoliciesPath string) []middleware.RatePolicy {
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
func parseJSONPolicies(data []byte) ([]middleware.RatePolicy, error) {
	var policies []middleware.RatePolicy
	err := json.Unmarshal(data, &policies)
	if err != nil {
		return nil, err
	}

	return policies, nil
}

// Tries to parse the given data into an array of policies assuming that the format is CSV
func parseCSVPolicies(fd *os.File) ([]middleware.RatePolicy, error) {
	lines, err := csv.NewReader(fd).ReadAll()
	if err != nil {
		return nil, err
	}

	policies := make([]middleware.RatePolicy, len(lines))
	for i, line := range lines {
		limit, err := strconv.Atoi(strings.TrimSpace(line[1]))
		if err != nil {
			log.Fatalf("Expect integer value for policy rate limit: %s" + err.Error())
		}

		minutes, err := strconv.Atoi(strings.TrimSpace(line[2]))
		if err != nil {
			log.Fatalf("Expect integer value for policy time limit: %s" + err.Error())
		}

		policies[i] = middleware.RatePolicy{
			Route:   strings.TrimSpace(line[0]),
			Limit:   limit,
			Minutes: minutes,
		}
	}

	return policies, nil
}
