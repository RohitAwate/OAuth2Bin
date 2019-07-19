package server

import (
	"net/http"

	"github.com/RohitAwate/OAuth2Bin/oauth2/config"
	"github.com/RohitAwate/OAuth2Bin/oauth2/utils"
)

func handleImplicitAuth(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	clientID := queryParams.Get("client_id")

	switch clientID {
	case "":
		utils.ShowError(w, r, 400, "Bad Request", "client_id is required")
	case serverConfig.AuthCodeCnfg.ClientID:
		utils.PresentAuthScreen(w, r, config.Implicit)
	default:
		utils.ShowError(w, r, 401, "Unauthorized", "Invalid client_id")
	}
}
