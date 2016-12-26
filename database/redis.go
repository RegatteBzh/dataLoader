package database

import redis "gopkg.in/redis.v3"

// Open create a redis client
func Open() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}
