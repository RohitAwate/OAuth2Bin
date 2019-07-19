package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Driver for the following tests.
//
// Deliberately not setting VisualError to true since it would call utils.ShowError
// which reads the template from the public/ folder which is not accessible to this
// package while running tests.
func TestPostFormValidatorHandle(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pfv := PostFormValidator{
			Request:     r,
			VisualError: false,
		}

		pfv.Handle(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Testing PostFormValidator for textual output")
		})(w, r)
	})

	testGetRequest(t, handler)
	testPostRequest(t, handler)
	testNoContentType(t, handler)
	testContentType(t, handler)
}

// Checks if a HTTP 405 status code is received on a GET request
func testGetRequest(t *testing.T, handler http.HandlerFunc) {
	getRecorder := httptest.NewRecorder()

	getReq, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(getRecorder, getReq)
	if getRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatal("GET request accepted by PostFormValidator middleware")
	}
}

// Checks if a HTTP 405 status code is received on a POST request
func testPostRequest(t *testing.T, handler http.HandlerFunc) {
	postRecorder := httptest.NewRecorder()

	postReq, err := http.NewRequest(http.MethodPost, "", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(postRecorder, postReq)
	if postRecorder.Code == http.StatusMethodNotAllowed {
		t.Fatal("POST request not accepted by PostFormValidator middleware")
	}
}

// Checks if a HTTP 400 status code is received on a POST request
// with application/json request as the content-type
func testContentType(t *testing.T, handler http.HandlerFunc) {
	recorder := httptest.NewRecorder()

	postReq, err := http.NewRequest(http.MethodPost, "", nil)
	if err != nil {
		t.Fatal(err)
	}

	postReq.Header.Add("Content-Type", "application/json")

	handler.ServeHTTP(recorder, postReq)
	if recorder.Code != http.StatusBadRequest {
		t.Fatal("Non application/x-www-form-urlencoded request accepted by PostFormValidator middleware")
	}
}

// Checks if a HTTP 400 status code is received on a POST request
// with no content-type
func testNoContentType(t *testing.T, handler http.HandlerFunc) {
	recorder := httptest.NewRecorder()

	postReq, err := http.NewRequest(http.MethodPost, "", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, postReq)
	if recorder.Code != http.StatusBadRequest {
		t.Fatal("Request with no content-type accepted by PostFormValidator middleware")
	}
}
