package algorithms

import (
	cache "arjuningole/ratelimiter/redis"
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

// TODO: Handle concurrent requests

// Two things are configuarable here
//  1. Bucket Size (Maximum allow request in a burst)
// 	2. Bucket Refilling Rate (How many tokens to add and at what rate)
//  We can also determine how many tokens each request takes (ex: For heavy compute we can make it take more tokens)
//  So we can make it configuarable on route level, ratehr than a middleware as well

// Keep it as short as possible to save memory
// tkb: takes less memory than tokenbucket:
const TokenBucketRedisKey = "tkb:"

type TokenBucketOptions struct {
	BUCKET_SIZE                int
	BUCKET_REFILL_AMOUNT       int
	BUCKET_REFILL_INTERVAL     int // this should be in seconds
	TOKEN_REQUIRED_FOR_REQUEST int
	REQUEST_UNIQUE_IDENTIFIER  string
}

type response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func TokenBucket(next http.Handler, options TokenBucketOptions) http.Handler {
	rdb := cache.GetRedisClient()
	ctx := context.Background()
	key := TokenBucketRedisKey + options.REQUEST_UNIQUE_IDENTIFIER // this is to set on roiute level else it would use the defualt bucket
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result, err := rdb.Get(ctx, key).Result()
		if err != nil {
			log.Println("Unable to Find the current Size")
			result = "0"
		}
		currentSize, err := strconv.Atoi(result)
		if currentSize-options.TOKEN_REQUIRED_FOR_REQUEST < 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(429)
			json.NewEncoder(w).Encode(response{
				Message: "Too Many Request",
				Status:  429,
			})
		} else {
			// Consume the tokens
			newSize := math.Max(float64(0), (float64(currentSize) - float64(options.TOKEN_REQUIRED_FOR_REQUEST)))
			rdb.Set(ctx, key, newSize, 0)
			next.ServeHTTP(w, r)
		}
	})
}

// Responsible for adding the tokens in the bucket
func TokenGenerator(context context.Context, options TokenBucketOptions) {
	rdb := cache.GetRedisClient()
	key := TokenBucketRedisKey + options.REQUEST_UNIQUE_IDENTIFIER
	// This will run periodically every BUCKET_REFILL_INTERVAL seconds
	// https://stackoverflow.com/questions/16466320/is-there-a-way-to-do-repetitive-tasks-at-intervals
	ticker := time.NewTicker(time.Second * time.Duration(options.BUCKET_REFILL_INTERVAL))
	defer ticker.Stop()

	// Set the value initially if not found
	exists, _ := rdb.Exists(context, key).Result()
	if exists == 0 {
		rdb.Set(context, key, options.BUCKET_SIZE, 0)
	}
	for {
		select {
		case <-context.Done(): // Basically Error
			return
		case <-ticker.C:
			result, err := rdb.Get(context, key).Result()
			if err != nil {
				log.Println("Unable to Find the current Size")
			}
			currentSize, err := strconv.Atoi(result)
			if currentSize < options.BUCKET_SIZE {
				newSize := math.Min(float64(options.BUCKET_SIZE), (float64(currentSize) + float64(options.BUCKET_REFILL_AMOUNT)))
				rdb.Set(context, key, newSize, 0)
			}
		}
	}
}
