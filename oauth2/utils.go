package oauth2

import (
	"fmt"
	"html/template"
	"net/http"
)

// PresentAuthScreen shows the authorization screen to the user
func PresentAuthScreen(w http.ResponseWriter, r *http.Request) {
	scopeList := []string{
		"Go to Mars",
		"Travel back in time",
	}

	authScreenStruct := struct {
		ScopeList []string
	}{
		ScopeList: scopeList,
	}

	RenderTemplate(w, r, "auth_grant", 200, authScreenStruct)
}

// NotFound presents the 404 screen to the user
func NotFound(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, r, "404", 404, nil)
}

// ErrorTemplate defines the set of parameters required to show an error page
type ErrorTemplate struct {
	Title string
	Desc  string
}

// Error presents the error screen to the user
func Error(w http.ResponseWriter, r *http.Request, status int, errTmpl ErrorTemplate) {
	RenderTemplate(w, r, "error", status, errTmpl)
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
