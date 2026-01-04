package main

import (
	"arjuningole/ratelimiter/algorithms"
	cache "arjuningole/ratelimiter/redis"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const PORT = ":9080"

var tokenBucketOptions = algorithms.TokenBucketOptions{
	BUCKET_SIZE:                64,
	BUCKET_REFILL_AMOUNT:       2,
	BUCKET_REFILL_INTERVAL:     5,
	TOKEN_REQUIRED_FOR_REQUEST: 1,
	REQUEST_UNIQUE_IDENTIFIER:  "gettb",
}

func main() {
	context, cancel := context.WithCancel(context.Background()) // This can be further extended for graceful shutdowns
	defer cancel()
	// Mux is just a multiplexer who's job is to route the request to the correct handler
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", rootHandler)
	go algorithms.TokenGenerator(context, tokenBucketOptions)
	mux.Handle("GET /token-bucket", algorithms.TokenBucket(http.HandlerFunc(genericSuccessResponder), tokenBucketOptions))
	fmt.Println("Running on PORT", PORT)
	http.ListenAndServe(PORT, applyMiddleware(mux,
		loggingMiddleware,
	))
}

// This is application level middle ware which wraps the entire mux object
func applyMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request:", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

type Response struct {
	IsHealthy bool `json:"isHealthy"`
	Status    int  `json:"status"`
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	rdb := cache.GetRedisClient()
	ctx := context.Background()
	ping, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	response := Response{
		IsHealthy: ping == "PONG",
		Status:    getStatusHelper(ping),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(getStatusHelper(ping))
	json.NewEncoder(w).Encode(response)
}

func getStatusHelper(ping string) int {
	if ping == "PONG" {
		return 200
	} else {
		return 500
	}
}

type GenericResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func genericSuccessResponder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(GenericResponse{
		Message: "Success",
		Status:  200,
	})
}
