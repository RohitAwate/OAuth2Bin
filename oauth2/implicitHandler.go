package oauth2

import "net/http"

func handleImplicitAuth(w http.ResponseWriter, r *http.Request) {
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
