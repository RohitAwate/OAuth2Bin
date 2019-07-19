package middleware

import (
	"net/http"

	"github.com/RohitAwate/OAuth2Bin/oauth2/utils"
)

// PostFormValidator verifies if a request has method POST and content-type
// "application/x-www-form-urlencoded". This is turned into a middleware since
// it is a common task in OA2B.
//
// Request: the request on which the middleware is to be applied
// VisualError: boolean which determines whether to present a visual (HTML)
// or textual (JSON) error in case the request doesn't satisfy the above conditions
type PostFormValidator struct {
	Request     *http.Request
	VisualError bool
}

// Handle implements the Middleware interface
// and performs the above mentioned job
func (pfv *PostFormValidator) Handle(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			title := "Method Not Allowed"
			desc := r.Method + " not allowed"
			pfv.presentError(w, r, http.StatusMethodNotAllowed, title, desc)
			return
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/x-www-form-urlencoded" {
			title := "Bad Request"

			var desc string
			if contentType == "" {
				desc = "Expecting content type"
			} else {
				desc = "Content type not allowed: " + contentType
			}

			pfv.presentError(w, r, http.StatusBadRequest, title, desc)
			return
		}

		handler.ServeHTTP(w, r)
	}
}

func (pfv *PostFormValidator) presentError(w http.ResponseWriter, r *http.Request, status int, title, desc string) {
	if pfv.VisualError {
		utils.ShowError(w, r, status, title, desc)
	} else {
		utils.ShowJSONError(w, r, status, utils.RequestError{
			Error: title,
			Desc:  desc,
		})
	}
}
