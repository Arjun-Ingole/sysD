// Redis Wrapper
// https://redis.io/docs/latest/develop/clients/go/
package cache

import (
	"sync"

	"github.com/redis/go-redis/v9"
)

// Ideally get these from env
const (
	redisAddress   string = "localhost:6379"
	redisPassword  string = ""
	defultDatabase int    = 0
	// This is the serialization protocol, how redis parses the data and handles data types, 2 is legacy and long used industry standard
	// 3 is newer, better standard, introduced in redis 7
	defualtProtocol int = 3
)

var (
	instance *redis.Client
	once     sync.Once // This is to ensure only one client is initiated and used throughout the app
)

func GetRedisClient() *redis.Client {
	once.Do(func() {
		instance = redis.NewClient(&redis.Options{
			Addr:     redisAddress,
			Password: redisPassword,
			DB:       defultDatabase,
			Protocol: defualtProtocol,
		})
	})
	return instance
}
