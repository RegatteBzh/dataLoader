package database

import (
	"github.com/spf13/viper"
	redis "gopkg.in/redis.v4"
)

// Open create a redis client
func Open() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis_host") + ":" + viper.GetString("redis_port"),
		Password: viper.GetString("redis_password"),
		DB:       0, // use default DB
	})
}
