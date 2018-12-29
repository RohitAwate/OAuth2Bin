package oauth2

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// OA2Server implements an OAuth 2.0 server
type OA2Server struct {
	Port   int
	Config OA2Config
}

// NewOA2Server returns a new OAuth 2.0 server which runs
// on the specified port with the specified configuration
func NewOA2Server(port int) *OA2Server {
	return &OA2Server{Port: port}
}

// Start listening for requests
func (s *OA2Server) Start() {
	public := http.FileServer(http.Dir("public/"))
	http.Handle("/public/", http.StripPrefix("/public/", public))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		file, err := ioutil.ReadFile("public/index.html")
		if err != nil {
			log.Fatal(err)
		}

		w.Write(file)
	})
	http.HandleFunc("/auth", handleAuth)
	http.ListenAndServe(":8080", nil)
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!")
}
