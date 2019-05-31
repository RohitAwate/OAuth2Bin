package oauth2

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// presentAuthScreen shows the authorization screen to the user
func presentAuthScreen(w http.ResponseWriter, r *http.Request, flow int) {
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
		"public/templates/base.html",
	)
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.ExecuteTemplate(w, "auth", authScreenStruct)
	if err != nil {
		log.Fatal(err)
	}
}

// showError presents the error screen to the user
func showError(w http.ResponseWriter, r *http.Request, status int, title string, desc string) {
	tmpl, err := template.ParseFiles(
		"public/templates/error.html",
		"public/templates/base.html",
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

	if !strings.Contains(url, "=") {
		return nil, fmt.Errorf("%s contains no key-value pairs", url)
	}

	pairs := make(map[string]string)
	for _, pair := range strings.Split(string(url), "&") {
		items := strings.Split(pair, "=")
		pairs[items[0]] = items[1]
	}

	return pairs, nil
}
