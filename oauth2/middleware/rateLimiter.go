package middleware

import (
	"fmt"
	"net/http"

	"github.com/RohitAwate/OAuth2Bin/oauth2/cache"
	"github.com/gomodule/redigo/redis"
)

// RatePolicy represents the rate limiting policy
// for a specific route.
//
// Route: the server route to apply the policy to
// Limit: the number of API calls allowed
// Minutes: the duration in minutes over which 'Limit' is imposed
type RatePolicy struct {
	Route   string `json:"route"`
	Limit   int    `json:"limit"`
	Minutes int    `json:"minutes"`
}

// RateLimiter is an implementation of Middleware.
// It holds a list of policies that are checked
// when the CheckLimit method is invoked.
type RateLimiter struct {
	Policies []RatePolicy
}

// Handle checks if the client is within the limits enforced by the policies
// and returns the appropriate boolean value.
func (rl RateLimiter) Handle(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		policy := rl.getRatePolicy(r.URL.Path)
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
			showError(policy, w, r)
		} else {
			handler.ServeHTTP(w, r)
		}
	}
}

// Searches the policies based on the route
func (rl RateLimiter) getRatePolicy(route string) *RatePolicy {
	for _, policy := range rl.Policies {
		if route == policy.Route {
			return &policy
		}
	}

	return nil
}

// TODO: try to use goroutines for Redis calls
// Registers a new hit for the route from the IP in Redis.
// Returns the current hit count or an error.
func setHit(policy *RatePolicy, ip string) (int, error) {
	conn := cache.NewConn()
	defer cache.CloseConn(conn)

	key := fmt.Sprintf("%s:%s", policy.Route, ip)
	res, err := redis.String(conn.Do("GET", key))

	// if key exists, increment
	if err == nil {
		return redis.Int(conn.Do("INCR", key))
		// else, set key with value 1 and set TTL according to policy
	} else if res == "" && err == redis.ErrNil {
		res, err = redis.String(conn.Do("SET", key, 1, "EX", policy.Minutes*60))
		if res == "OK" {
			return 1, nil
		}
	}

	return -1, nil
}

func showError(policy *RatePolicy, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
	fmt.Fprintf(w, "You have exceeded the rate limit of %d requests per %d minute(s) on this route.\n", policy.Limit, policy.Minutes)
}
