package oauth2

import (
	"net/http"
)

// FlowHandler handles the request for a specific OAuth 2.0 flow
type FlowHandler interface {
	HandleGrant(w http.ResponseWriter, r *http.Request)
}

//------------------------------- Implementations -------------------------------

// AuthCodeHandler handles the Authorization Code flow
type AuthCodeHandler struct {
	Config *AuthCodeConfig
}

// HandleGrant extracts the query parameters, presents an authorization screen
// and redirects to the redirect URL
func (h *AuthCodeHandler) HandleGrant(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	if queryParams.Get("redirect_uri") == "" && queryParams.Get("client_id") == "" {
		Error(w, r, 400, ErrorTemplate{Title: "Bad Request", Desc: "redirect_uri and client_id are required"})
	} else {
		PresentAuthScreen(w, r)
	}
}
