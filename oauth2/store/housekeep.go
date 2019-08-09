package store

import (
	"log"
	"sync"
	"time"

	"github.com/RohitAwate/OAuth2Bin/oauth2/cache"
	"github.com/RohitAwate/OAuth2Bin/oauth2/utils"
	"github.com/gomodule/redigo/redis"
)

// Array of housekeeping service functions
var housekeepingFuncs = [...]func(redis.Conn){
	authCodeTokenHousekeep, authCodeGrantHousekeep,
	implicitTokenHousekeep, ropcTokenHousekeep,
	clientCredsTokenHousekeep,
}

func init() {
	// Background goroutine that fires the housekeeping functions every
	// 5 minutes for cleaning up expired grants and tokens.
	go func() {
		log.Println("Housekeeping service has started")
		conn := cache.NewConn()
		wg := sync.WaitGroup{}

		for {
			for _, hkFunc := range housekeepingFuncs {
				wg.Add(1)
				hkFunc(conn)
				wg.Done()
			}

			utils.Sleep(5 * time.Second)
		}
	}()
}
