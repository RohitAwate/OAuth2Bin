package oauth2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
		showError(w, r, 405, "Method Not Allowed", r.Method+" not allowed.")
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
	case "refresh_token":
		handleRefresh(w, r, params)
		return
	default:
		showJSONError(w, r, 400, "grant_type is required")
		return
	}

	handler.issueToken(w, r, params)
}

type echoResponse struct {
	Method      string `json:"method"`
	HTTPVersion string `json:"httpVersion"`

	Body           string            `json:"body"`
	QueryParams    map[string]string `json:"queryParams"`
	URLEncodedForm map[string]string `json:"urlencodedForm"`
	MultipartForm  map[string]string `json:"mutipartForm"`

	Headers map[string]string `json:"headers"`
	Origin  string            `json:"origin"`
}

// [Auth Not Required] handleEcho echoes the request in the response body as JSON
func handleEcho(w http.ResponseWriter, r *http.Request) {
	// Generate response
	response := echoResponse{
		Method:      r.Method,
		HTTPVersion: fmt.Sprintf("%d.%d", r.ProtoMajor, r.ProtoMinor),
		Origin:      r.RemoteAddr,
	}

	response.Headers = make(map[string]string)
	for key, val := range r.Header {
		response.Headers[key] = val[0]
	}

	params := r.URL.Query()
	if len(params) > 0 {
		response.QueryParams = make(map[string]string)
		for key, val := range params {
			response.QueryParams[key] = val[0]
		}
	}

	// Parses application/x-www-form-urlencoded body
	// only for POST, PATCH and PUT requests
	// Since we are only interested in these request bodies,
	// we check if the length of the PostForm is greater than zero.
	r.ParseForm()
	if len(r.PostForm) > 0 {
		response.URLEncodedForm = make(map[string]string)
		for key, val := range r.PostForm {
			response.URLEncodedForm[key] = val[0]
		}
	}

	r.ParseMultipartForm(1024)
	if r.MultipartForm != nil {
		response.MultipartForm = make(map[string]string)
		// Add string key-value pairs
		for key, val := range r.MultipartForm.Value {
			response.MultipartForm[key] = val[0]
		}

		// Add the file key-value pairs. The name of the file is used as value
		for key, val := range r.MultipartForm.File {
			response.MultipartForm[key] = fmt.Sprintf("%s (%dB)", val[0].Filename, val[0].Size)
		}
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	} else {
		response.Body = string(body)
	}

	jsonStr, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		showJSONError(w, r, 500, struct {
			Error string `json:"error"`
		}{Error: "Error while processing request"})
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	fmt.Fprintln(w, string(jsonStr))
}