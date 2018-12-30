package main

import (
	"github.com/RohitAwate/OAuth2Bin/oauth2"
)

func main() {
	acc := oauth2.AuthCodeConfig{
		AuthURL:      "https://oauth2bin.herokuapp.com/authorize",
		TokenURL:     "https://oauth2bin.herokuapp.com/token",
		ClientID:     "clientID",
		ClientSecret: "clientSecret",
		AccessToken:  "accessToken",
	}

	ic := oauth2.ImplicitConfig{
		AuthURL:     "https://oauth2bin.herokuapp.com/authorize",
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
