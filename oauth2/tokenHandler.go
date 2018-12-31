package oauth2

import (
	"fmt"
	"net/http"

	"github.com/RohitAwate/OAuth2Bin/oauth2/store"
)

// TokenHandler handles the granting, refreshing and revoking of tokens
// for a specific OAuth 2.0 flow
type tokenHandler interface {
	grant(w http.ResponseWriter, r *http.Request, params map[string]string)
}

//------------------------------- Implementations -------------------------------

type authCodeTokenHandler struct {
	Params map[string]string
}

func (h *authCodeTokenHandler) grant(w http.ResponseWriter, r *http.Request, params map[string]string) {
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
