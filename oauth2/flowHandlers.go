package oauth2

import (
	"fmt"
	"net/http"

	"github.com/RohitAwate/OAuth2Bin/oauth2/store"
)

// flowHandler handles the request for a specific OAuth 2.0 flow
type flowHandler interface {
	handleAuth(w http.ResponseWriter, r *http.Request)
	grant(w http.ResponseWriter, r *http.Request, params map[string]string)
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

func (h *authCodeHandler) grant(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["client_id"] == "" && params["grant_type"] == "" &&
		params["redirect_uri"] == "" && params["code"] == "" {
		showJSONError(w, r, 400, "client_id, grant_type=authorization_code, code and redirect_uri are required.")
		return
	}

	token, err := store.NewAuthCodeToken(params["code"])
	if err != nil {
		showJSONError(w, r, 400, "The code supplied was used previously. The access token issued with that code has been revoked.")
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	fmt.Fprintln(w, token)
}
