package utils

import (
	"reflect"
	"testing"
)

func testParseParamsFunc(url string) func(*testing.T) {
	return func(t *testing.T) {
		result, err := ParseParams(url)
		if err != nil {
			t.Log("Returned an error.")
			return
		}

		expected := map[string]string{
			"client_id":     "CLIENT_ID",
			"client_secret": "CLIENT_SECRET",
			"grant_type":    "authorization_code",
			"code":          "AUTHORIZATION_CODE",
			"redirect_uri":  "CALLBACK_URL",
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%s failed: Maps not equal", t.Name())
		}
	}
}

func TestParseParams(t *testing.T) {
	t.Run("Empty string", testParseParamsFunc(""))
	t.Run("Full URL", testParseParamsFunc("https://cloud.digitalocean.com/v1/oauth/token?client_id=CLIENT_ID&client_secret=CLIENT_SECRET&grant_type=authorization_code&code=AUTHORIZATION_CODE&redirect_uri=CALLBACK_URL"))

	t.Run("Only Params", testParseParamsFunc("client_id=CLIENT_ID&client_secret=CLIENT_SECRET&grant_type=authorization_code&code=AUTHORIZATION_CODE&redirect_uri=CALLBACK_URL"))
	t.Run("Only Params with leading ?", testParseParamsFunc("client_id=CLIENT_ID&client_secret=CLIENT_SECRET&grant_type=authorization_code&code=AUTHORIZATION_CODE&redirect_uri=CALLBACK_URL"))

	t.Run("No queries", testParseParamsFunc("https://cloud.digitalocean.com/v1/oauth/token"))
	t.Run("No queries with leading ?", testParseParamsFunc("https://cloud.digitalocean.com/v1/oauth/token?"))
}
