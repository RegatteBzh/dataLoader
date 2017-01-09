package polar

import (
	"strconv"

	redis "gopkg.in/redis.v4"
)

func (sailChar SailCharacteristic) pushWindsRedis(c *redis.Client, redisName string) (err error) {
	c.Del(redisName)
	args := make([]interface{}, len(sailChar.Winds))
	for i, value := range sailChar.Winds {
		args[i] = interface{}(value)
	}
	cmd := c.LPush(redisName, args...)
	return cmd.Err()
}

func (sailChar SailCharacteristic) pushAnglesRedis(c *redis.Client, redisName string) (err error) {
	c.Del(redisName)
	args := make([]interface{}, len(sailChar.Polars))
	for i, polar := range sailChar.Polars {
		args[i] = interface{}(polar.Angle)
	}
	cmd := c.LPush(redisName, args...)
	return cmd.Err()
}

func (sailChar SailCharacteristic) pushSpeedsRedis(c *redis.Client, redisName string) (err error) {
	c.Del(redisName)
	args := make([]interface{}, len(sailChar.Polars))
	for i, polar := range sailChar.Polars {
		table := "["
		for j, value := range polar.Speed {
			if j > 0 {
				table = table + ","
			}
			table = table + strconv.FormatFloat(value, 'f', -1, 64)
		}
		table = table + "]"
		args[i] = interface{}(table)
	}
	cmd := c.LPush(redisName, args...)
	return cmd.Err()
}
