package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/RohitAwate/OAuth2Bin/oauth2/cache"
	"github.com/RohitAwate/OAuth2Bin/oauth2/utils"
)

func handleClientCredsToken(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["client_id"] != serverConfig.ClientCredsCnfg.ClientID ||
		params["client_secret"] != serverConfig.ClientCredsCnfg.ClientSecret {
		utils.ShowJSONError(w, r, 400, utils.RequestError{
			Error: "invalid_request",
			Desc:  "client_id and client_secret are missing or invalid",
		})
		return
	}

	// If everything checks out, issue the token
	token, err := cache.NewClientCredsToken()
	if err != nil {
		log.Println(err)
		if err != nil {
			utils.ShowJSONError(w, r, 500, utils.RequestError{
				Error: "Internal Server Error",
				Desc:  "Token generation failed. Please try again.",
			})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	jsonBytes, err := json.Marshal(token)

	fmt.Fprintln(w, string(jsonBytes))
}
