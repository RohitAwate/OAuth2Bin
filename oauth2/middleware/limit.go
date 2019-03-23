package middleware

import (
	"fmt"
	"net/http"
)

// Policy represents the rate limiting policy
// for a specific route.
//
// Route: the server route to apply the policy to
// Limit: the number of API calls allowed
// Period: the duration over which the Limit is imposed
type Policy struct {
	Route     string `json:"route"`
	Limit     int    `json:"limit"`
	PeriodMin int    `json:"period"`
}

// RateLimiter is an implementation of Middleware.
// It holds a list of policies that are checked
// when the CheckLimit method is invoked.
type RateLimiter struct {
	Policies []Policy
}

// CheckLimit checks if the client is within the limits enforced by the policies
// and returns the appropriate boolean value.
func (rl *RateLimiter) CheckLimit(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		policy := rl.getPolicy(r.URL.Path)
		if policy == nil {
			// letting this request pass since no policies are set
			handler.ServeHTTP(w, r)
			return
		}

		hits, err := setHit(policy, r.RemoteAddr)
		if err != nil {
			// letting this request pass since there may be an issue with Redis
			handler.ServeHTTP(w, r)
			return
		}

		if hits > policy.Limit {
			showError(w, r)
		} else {
			handler.ServeHTTP(w, r)
		}
	}
}

func (rl *RateLimiter) getPolicy(route string) *Policy {
	for _, policy := range rl.Policies {
		if route == policy.Route {
			return &policy
		}
	}

	return nil
}

func showError(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "You have exceeded the rate limit on this route")
}
