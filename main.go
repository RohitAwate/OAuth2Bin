package main

import (
	"os"

	"github.com/RohitAwate/OAuth2Bin/oauth2/server"
)

func main() {
	// Since Heroku allocates a port dynamically
	var port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := server.NewOA2Server(port, "config/flowParams.json", "config/ratePolicies.json")
	server.Start()
}
