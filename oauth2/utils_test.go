package oauth2

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseParams(t *testing.T) {
	queries := "https://cloud.digitalocean.com/v1/oauth/token?client_id=CLIENT_ID&client_secret=CLIENT_SECRET&grant_type=authorization_code&code=AUTHORIZATION_CODE&redirect_uri=CALLBACK_URL"

	result, err := parseParams(queries)
	if err != nil {
		panic(err)
	}

	expected := map[string]string{
		"client_id":     "CLIENT_ID",
		"client_secret": "CLIENT_SECRET",
		"grant_type":    "authorization_code",
		"code":          "AUTHORIZATION_CODE",
		"redirect_uri":  "CALLBACK_URL",
	}

	if !reflect.DeepEqual(result, expected) {
		fmt.Println("Result: ")
		for k, v := range result {
			fmt.Printf("%s: %s\n", k, v)
		}

		fmt.Println("\nExpected: ")
		for k, v := range expected {
			fmt.Printf("%s: %s\n", k, v)
		}

		t.Errorf("Something went wrong!\n")
	}
}
