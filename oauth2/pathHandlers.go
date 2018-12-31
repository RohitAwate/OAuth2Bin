package oauth2

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Routes the request to a AuthorizationHandler based on the request_type
func handleAuth(w http.ResponseWriter, r *http.Request) {
	var handler authorizationHandler
	switch r.URL.Query().Get("response_type") {
	case "code":
		handler = &authCodeHandler{}
	default:
		showError(w, r, 400, "Bad Request", "response_type is required")
		return
	}

	handler.handleAuth(w, r)
}

// Invoked by the Authorization Grant screen when the user accepts the authorization request.
// Extracts the redirect_uri from the JSON body, attaches an authorization grant to it,
// and redirects the user-agent to that URI.
func handleAccepted(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		showError(w, r, 405, "Bad Request", r.Method+" not allowed.")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		showError(w, r, 400, "Bad Request", "Could not read body of request.")
		return
	}

	var msg map[string]string
	err = json.Unmarshal(body, &msg)
	if err != nil {
		showError(w, r, 400, "Bad Request", "Invalid JSON found in request body.")
		return
	}

	redirectURI := msg["redirect_uri"] + "?code=" + serverConfig.AuthCodeCnfg.AuthGrant
	http.Redirect(w, r, redirectURI, http.StatusSeeOther)
}

// Accepts only POST requests with application/x-www-form-urlencoded body.
// Parses the body, verifies if client_id, grant_type, code and redirect_uri parameters are present.
// Refer: RFC 6749 Section 4.1.3 (https://tools.ietf.org/html/rfc6749#section-4.1.3)

// Returns a JSON object containing the access_token, refresh_token, token_type, expires_in
func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		showJSONError(w, r, 405, r.Method+" not allowed.")
		return
	} else if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		showJSONError(w, r, 400, "Invalid Content-Type: "+r.Header.Get("Content-Type"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		showJSONError(w, r, 500, "An error occurred while processing your request")
		return
	}

	params, err := parseParams(string(body))
	if err != nil {
		showJSONError(w, r, 400, "Expected parameters not found. Refer RFC 6749 Section 4.1.3 (https://tools.ietf.org/html/rfc6749#section-4.1.3)")
		return
	}

	var handler tokenHandler
	switch params["grant_type"] {
	case "authorization_code":
		handler = &authCodeTokenHandler{}
	default:
		showJSONError(w, r, 400, "grant_type is required")
		return
	}

	handler.grant(w, r, params)
}
