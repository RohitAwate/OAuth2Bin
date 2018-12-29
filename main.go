package main

import (
	"github.com/RohitAwate/OAuth2Bin/oauth2"
)

func main() {
	server := oauth2.NewOA2Server(8080)
	server.Start()
}
