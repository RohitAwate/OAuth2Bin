package main

import (
	"github.com/RohitAwate/OAuth2Bin/oauth2"
)

func main() {
	acc := oauth2.AuthCodeConfig{
		AuthURL:      "http://localhost:8080/authorize",
		TokenURL:     "http://localhost:8080/token",
		ClientID:     "clientID",
		ClientSecret: "clientSecret",
		AccessToken:  "accessToken",
	}

	ic := oauth2.ImplicitConfig{
		AuthURL:     "https://localhost:8080/authorize",
		ClientID:    "clientID",
		AccessToken: "accessToken",
	}

	config := oauth2.OA2Config{
		AuthCodeCnfg: acc,
		ImplicitCnfg: ic,
	}

	server := oauth2.NewOA2Server(8080, config)
	server.Start()
}
