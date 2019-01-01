package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/RohitAwate/OAuth2Bin/oauth2"
)

func main() {
	fd, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}

	jsonBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatal(err)
	}

	fd.Close()

	var config oauth2.OA2Config
	json.Unmarshal(jsonBytes, &config)

	// Since Heroku allocates a port dynamically
	var port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := oauth2.NewOA2Server(port, config)
	server.Start()
}
