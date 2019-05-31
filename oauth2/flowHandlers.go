package oauth2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RohitAwate/OAuth2Bin/oauth2/store"
)

// flowHandler handles the request for a specific OAuth 2.0 flow
type flowHandler interface {
	handleAuth(w http.ResponseWriter, r *http.Request)
}

// Enum for the OAuth 2.0 flows
const (
	AuthCode    = 1
	Implicit    = 2
	ROPC        = 3
	ClientCreds = 4
)

// Implementation of flowHandler for the Authorization Code Flow
type authCodeHandler struct {
}

// handleAuth checks for the existence of client_id in the query parameters.
// If not present, an HTTP 400 response is sent.
// If an unrecognized client_id is found, an HTTP 401 response is sent.
// Else, an authorization screen is presented to the user.
func (h *authCodeHandler) handleAuth(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	clientID := queryParams.Get("client_id")

	switch clientID {
	case "":
		showError(w, r, 400, "Bad Request", "client_id is required")
	case serverConfig.AuthCodeCnfg.ClientID:
		presentAuthScreen(w, r, AuthCode)
	default:
		showError(w, r, 401, "Unauthorized", "Invalid client_id")
	}
}

// issueToken checks for the existence of all parameters detailed in Section 4.1.3 of RFC 6749 (https://tools.ietf.org/html/rfc6749#section-4.1.3).
// If not present, an HTTP 400 response is sent.
// Else a new token is generated, added to the store, and returned to the user in a JSON response.
func handleAuthCodeToken(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["client_id"] == "" || params["grant_type"] == "" ||
		params["redirect_uri"] == "" || params["code"] == "" {
		showJSONError(w, r, 400, requestError{
			Error: "invalid_request",
			Desc:  "client_id, grant_type=authorization_code, code and redirect_uri are required.",
		})
		return
	}

	token, err := store.NewAuthCodeToken(params["code"], "")
	if err != nil {
		showJSONError(w, r, 400, requestError{
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
func handleRefresh(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["refresh_token"] == "" {
		showJSONError(w, r, 400, requestError{
			Error: "invalid_request",
			Desc:  "refresh_token required",
		})
		return
	}

	// If found, invalidate previously issued token
	if store.RefreshTokenExists(params["refresh_token"], true) {
		token, err := store.NewRefreshToken(params["refresh_token"])
		if err != nil {
			showJSONError(w, r, 500, requestError{
				Error: "could not generate token",
				Desc:  err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		jsonBytes, err := json.Marshal(token)

		fmt.Fprintln(w, string(jsonBytes))
	} else {
		showJSONError(w, r, 400, requestError{
			Error: "invalid refresh_token",
			Desc:  "expired or invalid refresh token",
		})
	}
}

// Implementation of flowHandler for the Implicit Grant Flow
type implicitHandler struct {
}

func (*implicitHandler) handleAuth(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	clientID := queryParams.Get("client_id")

	switch clientID {
	case "":
		showError(w, r, 400, "Bad Request", "client_id is required")
	case serverConfig.AuthCodeCnfg.ClientID:
		presentAuthScreen(w, r, Implicit)
	default:
		showError(w, r, 401, "Unauthorized", "Invalid client_id")
	}
}
