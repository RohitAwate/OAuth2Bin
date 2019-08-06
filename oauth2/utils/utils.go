package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// PresentAuthScreen shows the authorization screen to the user
func PresentAuthScreen(w http.ResponseWriter, r *http.Request, flow int) {
	scopeList := []string{
		"Fly to Mars",
		"Travel back in time",
		"Ride a dragon",
	}

	authScreenStruct := struct {
		ScopeList []string
		Flow      int
	}{
		ScopeList: scopeList,
		Flow:      flow,
	}

	tmpl, err := template.ParseFiles(
		"public/templates/authScreen.html",
		"public/templates/nav.html",
		"public/templates/footer.html",
	)
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.ExecuteTemplate(w, "auth", authScreenStruct)
	if err != nil {
		log.Fatal(err)
	}
}

// ShowError presents the error screen to the user
func ShowError(w http.ResponseWriter, r *http.Request, status int, title string, desc string) {
	tmpl, err := template.ParseFiles(
		"public/templates/error.html",
		"public/templates/nav.html",
		"public/templates/footer.html",
	)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(status)
	err = tmpl.ExecuteTemplate(w, "error", struct {
		Title string
		Desc  string
	}{Title: title, Desc: desc})
	if err != nil {
		log.Fatal(err)
	}
}

// RequestError is used as response for failed requests.
// Using the necessary structures mentioned in RFC 6749 Section 4.1.2.1 (https://tools.ietf.org/html/rfc6749#section-4.1.2.1)
// error_uri is ignored since this is not a real API and has no documentation.
// state is ignored because it is ignored by flowHandlers.
type RequestError struct {
	Error string `json:"error"`
	Desc  string `json:"error_description"`
}

// ShowJSONError presents the error to the user or application
// in the form of a JSON string
func ShowJSONError(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "An error occurred while processing your request.")
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, string(body))
}

// RenderTemplate renders the template with the given template, sets the status code for the response
func RenderTemplate(w http.ResponseWriter, r *http.Request, templateName string, status int, data interface{}) {
	template, err := template.ParseFiles(fmt.Sprintf("public/templates/%s.html", templateName))
	if err != nil {
		panic(err)
	}

	w.WriteHeader(status)
	template.Execute(w, data)
}

// ParseParams parses a URL string containing application/x-www-urlencoded
// parameters and returns a map of string key-value pairs of the same
func ParseParams(str string) (map[string]string, error) {
	str, err := url.QueryUnescape(str)
	if err != nil {
		return nil, err
	}

	if strings.Contains(str, "?") {
		str = strings.Split(str, "?")[1]
	}

	if !strings.Contains(str, "=") {
		return nil, fmt.Errorf("\"%s\" contains no key-value pairs", str)
	}

	pairs := make(map[string]string)
	for _, pair := range strings.Split(string(str), "&") {
		items := strings.Split(pair, "=")
		pairs[items[0]] = items[1]
	}

	return pairs, nil
}

// ParseBasicAuthHeader decodes the Basic Auth header.
// First checks if the string contains the substring "Basic"
// and strips it off if present.
// Returns the username:password pair
func ParseBasicAuthHeader(header string) (string, string) {
	// Trimming leading and trailing whitespace
	header = strings.TrimSpace(header)

	// Check if the entire header value was used as the argument
	// eg: Basic Y2xpZW50SUQ6Y2xpZW50U2VjcmV0
	// If yes, strip off "Basic "
	if strings.HasPrefix(header, "Basic ") {
		header = strings.Split(header, " ")[1]
	}

	bytes, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		log.Println(err)
		return "", ""
	}

	str := string(bytes)
	pair := strings.Split(str, ":")
	if len(pair) != 2 {
		return "", ""
	}

	return pair[0], pair[1]
}

// Backspace clears 'places' number of characters from the console
func Backspace(places int) {
	str := strings.Repeat("\r \r", places)
	fmt.Print(str)
}
