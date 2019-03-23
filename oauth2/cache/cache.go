package cache

import (
	"fmt"
	"os"

	"github.com/gomodule/redigo/redis"
)

var pool redis.Pool

// NewConn returns a Redis connection
// It is the responsibility of the receiver to close the connection.
func NewConn() redis.Conn {
	return pool.Get()
}

func init() {
	pool = redis.Pool{
		MaxActive: 30,
		MaxIdle:   10,
		Dial: func() (redis.Conn, error) {
			var conn redis.Conn
			var err error
			if os.Getenv("REDIS_HOST") == "" && os.Getenv("REDIS_PASS") == "" && os.Getenv("REDIS_PORT") == "" {
				// Defaults to a local Redis server
				conn, err = redis.Dial("tcp", ":6379")
			} else {
				addr := fmt.Sprintf("redis://:%s@%s:%s", os.Getenv("REDIS_PASS"),
					os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
				conn, err = redis.DialURL(addr)
			}

			if err != nil {
				// Panics if connection could not be established with a Redis server
				panic(err)
			}

			return conn, nil
		},
	}
}
