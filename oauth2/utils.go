package oauth2

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// presentAuthScreen shows the authorization screen to the user
func presentAuthScreen(w http.ResponseWriter, r *http.Request) {
	scopeList := []string{
		"Go to Mars",
		"Travel back in time",
	}

	authScreenStruct := struct {
		ScopeList []string
	}{
		ScopeList: scopeList,
	}

	renderTemplate(w, r, "authGrant", 200, authScreenStruct)
}

// showNotFound presents the 404 screen to the user
func showNotFound(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "404", 404, nil)
}

// showError presents the error screen to the user
func showError(w http.ResponseWriter, r *http.Request, status int, title string, desc string) {
	renderTemplate(w, r, "error", status, struct {
		Title string
		Desc  string
	}{Title: title, Desc: desc})
}

func showJSONError(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
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

// renderTemplate renders the template with the given template, sets the status code for the response
func renderTemplate(w http.ResponseWriter, r *http.Request, templateName string, status int, data interface{}) {
	template, err := template.ParseFiles(fmt.Sprintf("public/templates/%s.html", templateName))
	if err != nil {
		panic(err)
	}

	w.WriteHeader(status)
	template.Execute(w, data)
}

// parseParams parses a URL string containing application/x-www-urlencoded
// parameters and returns a map of string key-value pairs of the same
func parseParams(url string) (map[string]string, error) {
	if strings.Contains(url, "?") {
		url = strings.Split(url, "?")[1]
	}

	if !strings.Contains(url, "&") {
		return nil, fmt.Errorf("%s contains no key-value pairs", url)
	}

	pairs := make(map[string]string)
	for _, pair := range strings.Split(string(url), "&") {
		items := strings.Split(pair, "=")
		pairs[items[0]] = items[1]
	}

	return pairs, nil
}
