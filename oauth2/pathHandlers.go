package oauth2

import (
	"io/ioutil"
	"net/http"

	"github.com/RohitAwate/OAuth2Bin/oauth2/store"
)

// Routes the request to a AuthorizationHandler based on the request_type
func handleAuth(w http.ResponseWriter, r *http.Request) {
	var handler flowHandler
	params := r.URL.Query()

	if params.Get("response_type") == "" || params.Get("client_id") == "" || params.Get("redirect_uri") == "" {
		showError(w, r, 400, "Bad Request", "response_type, client_id and redirect_uri are required.")
		return
	}

	switch r.URL.Query().Get("response_type") {
	case "code":
		handler = &authCodeHandler{}
	}

	handler.handleAuth(w, r)
}

// Invoked by the Authorization Grant screen when the user accepts the authorization request.
// Extracts the redirect_uri from the JSON body, attaches an authorization grant to it,
// and redirects the user-agent to that URI.
func handleResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		showError(w, r, 405, "Bad Request", r.Method+" not allowed.")
		return
	}

	response := r.FormValue("response")
	redirectURI := r.FormValue("redirectURI")

	if response == "ACCEPT" {
		redirectURI += "?code=" + store.NewAuthCodeGrant()
	} else if response == "CANCEL" {
		redirectURI += "?error=access_denied"
	}

	http.Redirect(w, r, redirectURI, http.StatusSeeOther)
}

// Redirects the request to the appropriate flowHandler by checking the 'grant_type' parameter.
// Refer RFC 6749 Section 4.1.3 (https://tools.ietf.org/html/rfc6749#section-4.1.3)
// Accepts only POST requests with application/x-www-form-urlencoded body.
func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		showJSONError(w, r, 405, struct {
			Error string `json:"error"`
		}{Error: r.Method + " not allowed."})
		return
	} else if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		showJSONError(w, r, 400, struct {
			Error string `json:"error"`
		}{Error: "Invalid Content-Type: " + r.Header.Get("Content-Type")})
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

	var handler flowHandler
	switch params["grant_type"] {
	case "authorization_code":
		handler = &authCodeHandler{}
	default:
		showJSONError(w, r, 400, "grant_type is required")
		return
	}

	handler.issueToken(w, r, params)
}
