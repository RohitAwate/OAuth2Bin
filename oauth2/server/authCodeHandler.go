package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RohitAwate/OAuth2Bin/oauth2/config"
	"github.com/RohitAwate/OAuth2Bin/oauth2/store"
	"github.com/RohitAwate/OAuth2Bin/oauth2/utils"
)

// handleAuth checks for the existence of client_id in the query parameters.
// If not present, an HTTP 400 response is sent.
// If an unrecognized client_id is found, an HTTP 401 response is sent.
// Else, an authorization screen is presented to the user.
func handleAuthCodeAuth(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	clientID := queryParams.Get("client_id")

	switch clientID {
	case "":
		utils.ShowError(w, r, 400, "Bad Request", "client_id is required")
	case serverConfig.AuthCodeCnfg.ClientID:
		utils.PresentAuthScreen(w, r, config.AuthCode)
	default:
		utils.ShowError(w, r, 401, "Unauthorized", "Invalid client_id")
	}
}

// issueToken checks for the existence of all parameters detailed in Section 4.1.3 of RFC 6749 (https://tools.ietf.org/html/rfc6749#section-4.1.3).
// If not present, an HTTP 400 response is sent.
// Else a new token is generated, added to the store, and returned to the user in a JSON response.
func handleAuthCodeToken(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["client_id"] == "" || params["grant_type"] == "" ||
		params["redirect_uri"] == "" || params["code"] == "" {
		utils.ShowJSONError(w, r, 400, utils.RequestError{
			Error: "invalid_request",
			Desc:  "client_id, grant_type=authorization_code, code and redirect_uri are required",
		})
		return
	}

	token, err := store.NewAuthCodeToken(params["code"], "")
	if err != nil {
		utils.ShowJSONError(w, r, 400, utils.RequestError{
			Error: "invalid_request",
			Desc:  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	jsonBytes, err := json.Marshal(token)

	fmt.Fprintln(w, string(jsonBytes))
}

// Refer RFC 6749 Section 6 (https://tools.ietf.org/html/rfc6749#section-6)
func handleAuthCodeRefresh(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["refresh_token"] == "" {
		utils.ShowJSONError(w, r, 400, utils.RequestError{
			Error: "invalid_request",
			Desc:  "refresh_token required",
		})
		return
	}

	// If found, invalidate previously issued token
	if store.AuthCodeRefreshTokenExists(params["refresh_token"], true) {
		token, err := store.NewAuthCodeRefreshToken(params["refresh_token"])
		if err != nil {
			utils.ShowJSONError(w, r, 500, utils.RequestError{
				Error: "Internal Server Error",
				Desc:  "Token generation failed. Please try again.",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		jsonBytes, err := json.Marshal(token)

		fmt.Fprintln(w, string(jsonBytes))
	} else {
		utils.ShowJSONError(w, r, 400, utils.RequestError{
			Error: "invalid_refresh_token",
			Desc:  "expired or invalid refresh token",
		})
	}
}
