package oauth2

import (
	"net/http"
)

// authorizationHandler handles the request for a specific OAuth 2.0 flow
type authorizationHandler interface {
	handleAuth(w http.ResponseWriter, r *http.Request)
}

// Implementation of authorizationHandler for the Authorization Code Flow
type authCodeHandler struct {
}

// authCodeHandler checks for the existence of client_id in the query parameters.
// If not present, an HTTP 400 response is sent.
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
