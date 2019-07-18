package store

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/RohitAwate/OAuth2Bin/oauth2/cache"
	"github.com/gomodule/redigo/redis"
)

type authCodeTokenStruct struct {
	Token AuthCodeToken     `json:"token"`
	Meta  authCodeTokenMeta `json:"meta"`
}

type ropcTokenStruct struct {
	Token ROPCToken     `json:"token"`
	Meta  ropcTokenMeta `json:"meta"`
}

func tokenHousekeep(wg *sync.WaitGroup) {
	defer wg.Done()

	var token authCodeTokenStruct
	var err error
	var diff time.Duration

	conn := cache.NewConn()
	defer conn.Close()

	items, err := redis.ByteSlices(conn.Do("HGETALL", authCodeTokensSet))
	if err != nil {
		log.Println(err)
		return
	}

	for i := 1; i < len(items); i += 2 {
		err = json.Unmarshal(items[i], &token)
		if err != nil {
			log.Println(err)
			break
		}

		diff = time.Now().Sub(token.Meta.CreationTime)
		if diff >= time.Hour {
			conn.Do("HDEL", authCodeTokensSet, items[i-1])
		}
	}
}

func grantHousekeep(wg *sync.WaitGroup) {
	defer wg.Done()

	var intTime int64
	var issueTime time.Time

	conn := cache.NewConn()
	defer conn.Close()

	grants, err := redis.Strings(conn.Do("HGETALL", authCodeGrantSet))
	if err != nil {
		log.Println(err)
		return
	}

	for i := 1; i < len(grants); i += 2 {
		intTime, _ = strconv.ParseInt(grants[i], 10, 64)
		issueTime = time.Unix(intTime, 0)

		if time.Now().Sub(issueTime) >= time.Minute*10 {
			conn.Do("HDEL", authCodeGrantSet, grants[i-1])
		}
	}
}

func init() {
	// Background goroutine that fires the housekeeping function every
	// 5 minutes for cleaning up expired grants and tokens.
	go func() {
		log.Println("Housekeeping service has started.")
		timer := time.NewTimer(5 * time.Minute)
		wg := sync.WaitGroup{}
		for {
			wg.Add(2)
			go tokenHousekeep(&wg)
			go grantHousekeep(&wg)
			wg.Wait()
			<-timer.C
		}
	}()
}
