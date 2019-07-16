package oauth2

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/RohitAwate/OAuth2Bin/oauth2/store"
)

func handleClientCredsToken(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["client_id"] != serverConfig.ROPCCnfg.ClientID ||
		params["client_secret"] != serverConfig.ROPCCnfg.ClientSecret {
		showJSONError(w, r, 400, requestError{
			Error: "invalid_request",
			Desc:  "client_id and client_secret are missing or invalid",
		})
		return
	}

	// If everything checks out, issue the token
	token, err := store.NewClientCredsToken()
	if err != nil {
		log.Println(err)
		if err != nil {
			showJSONError(w, r, 500, requestError{
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
