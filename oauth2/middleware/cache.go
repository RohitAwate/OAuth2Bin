package middleware

import (
	"fmt"

	"github.com/gomodule/redigo/redis"

	"github.com/RohitAwate/OAuth2Bin/oauth2/cache"
)

func setHit(policy *Policy, ip string) (int, error) {
	conn := cache.NewConn()
	defer conn.Close()

	key := fmt.Sprintf("%s:%s", policy.Route, ip)
	res, err := redis.String(conn.Do("GET", key))

	// if key exists, increment
	if err == nil {
		return redis.Int(conn.Do("INCR", key))
		// else, set key with value 1 and set TTL according to policy
	} else if res == "" && err == redis.ErrNil {
		res, err = redis.String(conn.Do("SET", key, 1, "EX", policy.PeriodMin*60))
		if res == "OK" {
			return 1, nil
		}

		return -1, nil
	} else {
		return -1, nil
	}
}
