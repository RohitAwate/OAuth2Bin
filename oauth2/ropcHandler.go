package oauth2

import (
	"net/http"
)

// Checks if the values for username, password, client_id and client_secret match the
// server presets.
// If yes, an access token is issued.
//
// Refer: https://tools.ietf.org/html/rfc6749#section-4.3.2
func handleROPCToken(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if params["username"] != serverConfig.ROPCCnfg.Username ||
		params["password"] != serverConfig.ROPCCnfg.Password ||
		params["client_id"] != serverConfig.ROPCCnfg.ClientID ||
		params["client_secret"] != serverConfig.ROPCCnfg.ClientSecret {
		showJSONError(w, r, 400, requestError{
			Error: "missing_or_invalid_parameters",
			Desc:  "username, password, client_id and client_secret are missing or invalid",
		})
		return
	}

	// issue token
}
