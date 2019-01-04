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
	issueToken(w http.ResponseWriter, r *http.Request, params map[string]string)
}

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

	if clientID == "" {
		showError(w, r, 400, "Bad Request", "client_id is required")
	} else if clientID == serverConfig.AuthCodeCnfg.ClientID {
		presentAuthScreen(w, r)
	} else {
		showError(w, r, 401, "Unauthorized", "Invalid client_id")
	}
}

// issueToken checks for the existence of all parameters detailed in Section 4.1.3 of RFC 6749 (https://tools.ietf.org/html/rfc6749#section-4.1.3).
// If not present, an HTTP 400 response is sent.
// Else a new token is generated, added to the store, and returned to the user in a JSON response.
func (h *authCodeHandler) issueToken(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["client_id"] == "" || params["grant_type"] == "" ||
		params["redirect_uri"] == "" || params["code"] == "" {
		showJSONError(w, r, 400, requestError{Error: "invalid_request",
			Desc: "client_id, grant_type=authorization_code, code and redirect_uri are required."})
		return
	}

	token, err := store.NewAuthCodeToken(params["code"], params["client_id"])
	if err != nil {
		showJSONError(w, r, 400, requestError{Error: "invalid_request",
			Desc: "The code supplied was used previously. The access token issued with that code has been revoked."})
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	jsonBytes, err := json.Marshal(token)

	fmt.Fprintln(w, string(jsonBytes))
}
