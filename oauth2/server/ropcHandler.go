package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/RohitAwate/OAuth2Bin/oauth2/store"
	"github.com/RohitAwate/OAuth2Bin/oauth2/utils"
)

// Checks if the values for username, password, client_id and client_secret match the server presets.
// If yes, an access token is issued.
// Refer: https://tools.ietf.org/html/rfc6749#section-4.3.2
func handleROPCToken(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["username"] != serverConfig.ROPCCnfg.Username ||
		params["password"] != serverConfig.ROPCCnfg.Password ||
		params["client_id"] != serverConfig.ROPCCnfg.ClientID ||
		params["client_secret"] != serverConfig.ROPCCnfg.ClientSecret {
		utils.ShowJSONError(w, r, http.StatusBadRequest, utils.RequestError{
			Error: "invalid_request",
			Desc:  "username, password, client_id and client_secret are missing or invalid",
		})
		return
	}

	// If everything checks out, issue the token
	token, err := store.NewROPCToken("")
	if err != nil {
		log.Println(err)
		if err != nil {
			utils.ShowJSONError(w, r, http.StatusInternalServerError, utils.RequestError{
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

func handleROPCRefresh(w http.ResponseWriter, r *http.Request, params map[string]string) {
	// Invalidate previously issued token
	if store.ROPCRefreshTokenExists(params["refresh_token"], true) {
		token, err := store.NewROPCRefreshToken(params["refresh_token"])
		if err != nil {
			utils.ShowJSONError(w, r, http.StatusInternalServerError, utils.RequestError{
				Error: "Internal Server Error",
				Desc:  "Token generation failed. Please try again.",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		jsonBytes, err := json.Marshal(token)

		fmt.Fprintln(w, string(jsonBytes))
	} else {
		utils.ShowJSONError(w, r, http.StatusBadRequest, utils.RequestError{
			Error: "invalid_refresh_token",
			Desc:  "expired or invalid refresh token",
		})
	}
}
