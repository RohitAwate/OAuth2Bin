package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/RohitAwate/OAuth2Bin/oauth2/cache"
	"github.com/RohitAwate/OAuth2Bin/oauth2/config"
	"github.com/RohitAwate/OAuth2Bin/oauth2/utils"
)

// Routes the request to a AuthorizationHandler based on the request_type
func handleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ShowError(w, r, 405, "Method Not Allowed", r.Method+" not allowed.")
		return
	}

	params := r.URL.Query()

	// Perform empty checks on the following parameters:
	// - response_type
	// - client_id
	if params.Get("response_type") == "" || params.Get("client_id") == "" {
		utils.ShowError(w, r, 400, "Bad Request", "response_type and client_id are required.")
		return
	}

	switch r.URL.Query().Get("response_type") {
	case "code":
		handleAuthCodeAuth(w, r)
	case "token":
		handleImplicitAuth(w, r)
	default:
		utils.ShowError(w, r, http.StatusBadRequest, "Authorization Flow Error", "Unknown response_type: "+r.URL.Query().Get("response_type"))
	}
}

// Invoked by the Authorization Grant screen when the user accepts the authorization request.
// Extracts the redirect_uri from the JSON body, attaches an authorization grant to it,
// and redirects the user-agent to that URI.
func handleResponse(w http.ResponseWriter, r *http.Request) {
	flow, err := strconv.Atoi(r.FormValue("flow"))
	if err != nil {
		utils.ShowError(w, r, 400, "OAuth 2.0 Flow Error", "Unrecognized flow")
		return
	}

	response := r.FormValue("response")
	redirectURI, err := url.QueryUnescape(r.FormValue("redirectURI"))
	if err != nil {
		utils.ShowError(w, r, http.StatusBadRequest, "Bad Request", "Invalid redirect_uri")
		return
	}

	if response == "ACCEPT" {
		switch flow {
		case config.AuthCode:
			redirectURI += "?code=" + cache.NewAuthCodeGrant(redirectURI)
		case config.Implicit:
			token, err := cache.NewImplicitToken()
			if err != nil {
				utils.ShowError(w, r, 500, "Internal Server Error", "Token generation failed. Please try again.")
				return
			}

			redirectURI += fmt.Sprintf("#access_token=%s&token_type=bearer&expires_in=%d", token.AccessToken, token.ExpiresIn)
		}

	} else if response == "CANCEL" {
		redirectURI += "?error=access_denied"
	}

	http.Redirect(w, r, redirectURI, http.StatusSeeOther)
}

// Redirects the request to the appropriate flowHandler by checking the 'grant_type' parameter.
// Refer RFC 6749 Section 4.1.3 (https://tools.ietf.org/html/rfc6749#section-4.1.3)
// Accepts only POST requests with application/x-www-form-urlencoded body.
func handleToken(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.ShowJSONError(w, r, 500, "An error occurred while processing your request")
		return
	}

	params, err := utils.ParseParams(string(body))
	if err != nil {
		log.Println(err)
		utils.ShowJSONError(w, r, 400, "Expected parameters not found.")
		return
	}

	if params["client_id"] == "" && params["client_secret"] == "" {
		clientID, clientSecret := utils.ParseBasicAuthHeader(r.Header.Get("Authorization"))
		params["client_id"] = clientID
		params["client_secret"] = clientSecret
	}

	switch params["grant_type"] {
	case "authorization_code":
		handleAuthCodeToken(w, r, params)
	case "password":
		handleROPCToken(w, r, params)
	case "client_credentials":
		handleClientCredsToken(w, r, params)
	case "refresh_token":
		if len(params["refresh_token"]) != 72 {
			utils.ShowJSONError(w, r, 400, utils.RequestError{
				Error: "invalid_request",
				Desc:  "refresh_token missing or invalid",
			})
			return
		}

		if strings.HasPrefix(params["refresh_token"], cache.AuthCodeFlowID) {
			handleAuthCodeRefresh(w, r, params)
		} else if strings.HasPrefix(params["refresh_token"], cache.ROPCFlowID) {
			handleROPCRefresh(w, r, params)
		}
	default:
		utils.ShowJSONError(w, r, 400, "grant_type absent or invalid")
	}
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
		utils.ShowJSONError(w, r, 500, struct {
			Error string `json:"error"`
		}{Error: "Error while processing request"})
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	fmt.Fprintln(w, string(jsonStr))
}
