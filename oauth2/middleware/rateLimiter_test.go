package middleware

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestLimiterHandle(t *testing.T) {
	policies := make([]RatePolicy, 1)
	policies[0] = RatePolicy{
		Route:   "/",
		Limit:   50,
		Minutes: 1,
	}

	limiter := RateLimiter{Policies: policies}
	http.HandleFunc("/", limiter.Handle(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from OAuth 2.0 Bin!")
	}))
	go http.ListenAndServe(":8081", nil)

	time.Sleep(3 * time.Second)

	// Go creates a separate TCP connection for every HTTP request
	// if we use http.Get(url). This creates a problem in our case since
	// OA2B sees a different port number for every request thereby allowing
	// it to pass. Hence, to properly test out the rate-limiting middleware, we
	// use a custom single client for making requests which keeps the
	// TCP connection alive between requests.
	//
	// Reference: https://awmanoj.github.io/tech/2016/12/16/keep-alive-http-requests-in-golang/
	transport := &http.Transport{
		MaxIdleConnsPerHost: 1024,
		TLSHandshakeTimeout: 0 * time.Second,
	}
	client := &http.Client{Transport: transport}

	// We send policy limit - 1 requests in this loop, which should exhaust the limit.
	for i := 0; i < policies[0].Limit; i++ {
		res, err := client.Get("http://localhost:8081")
		if err != nil || res.StatusCode != 200 {
			t.Fatalf("request failed\n")
		}

		io.Copy(ioutil.Discard, res.Body)
		defer res.Body.Close()
		t.Logf("HTTP %d: request succeeded\n", res.StatusCode)
	}

	// This request is made beyond the prescribed limit and
	// must thus give HTTP 429 status code.
	res, err := client.Get("http://localhost:8081")
	if err != nil {
		t.Fatalf("request failed\n")
	}

	io.Copy(ioutil.Discard, res.Body)
	defer res.Body.Close()
	if res.StatusCode != 429 {
		t.Fatalf("HTTP %d: request allowed beyond policy limit\n", res.StatusCode)
	}
}
