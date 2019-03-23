package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/RohitAwate/OAuth2Bin/oauth2/middleware"

	"github.com/RohitAwate/OAuth2Bin/oauth2"
)

func main() {
	// Since Heroku allocates a port dynamically
	var port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := oauth2.NewOA2Server(port, *getServerConfig())
	policies := getRatePolicies()
	if policies != nil {
		server.SetRateLimiter(policies)
	}

	server.Start()
}

func getServerConfig() *oauth2.OA2Config {
	fd, err := os.Open("config/server.json")
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

	var config oauth2.OA2Config
	err = json.Unmarshal(jsonBytes, &config)
	if err != nil {
		log.Fatal(err)
	}

	return &config
}

func getRatePolicies() []middleware.Policy {
	fd, err := os.Open("config/policy.json")
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
