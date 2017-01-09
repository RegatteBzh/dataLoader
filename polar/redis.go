package polar

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/sethgrid/multibar"

	redis "gopkg.in/redis.v4"
)

// LoadRedis loads sails from Redis
func LoadRedis(client *redis.Client, redisName string) (sail SailCharacteristic, err error) {

	windSlice, err := client.LRange(redisName+"_winds", 0, 100).Result()
	if err != nil {
		log.Fatal(err)
	}
	sail.Winds = make([]float64, len(windSlice))
	for i, wind := range windSlice {
		sail.Winds[i], err = strconv.ParseFloat(wind, 64)
		if err != nil {
			log.Fatal(err)
		}
	}

	angleSlice, err := client.LRange(redisName+"_angles", 0, 190).Result()
	if err != nil {
		log.Fatal(err)
	}
	speedSlice, err := client.LRange(redisName+"_speeds", 0, 190).Result()
	if err != nil {
		log.Fatal(err)
	}
	sail.Polars = make([]Polar, len(angleSlice))
	for j, angle := range angleSlice {
		sail.Polars[j].Angle, err = strconv.ParseFloat(angle, 64)
		if err != nil {
			log.Fatal(err)
		}
		speeds := strings.Split(speedSlice[j], ",")
		sail.Polars[j].Speed = make([]float64, len(speeds))
		for k, speed := range speeds {
			sail.Polars[j].Speed[k], err = strconv.ParseFloat(speed, 64)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return
}

func (sail SailCharacteristic) saveRedis(c *redis.Client, redisName string, progressBar multibar.ProgressFunc, fake bool) (err error) {

	interpolatedSails := SailCharacteristic{
		Name:  sail.Name,
		Winds: sail.Winds,
	}

	for i, polar := range sail.Polars {
		if i > 0 {
			interpolation := sail.Polars[i-1].interpolate(polar)
			interpolatedSails.Polars = append(interpolatedSails.Polars, interpolation...)
		}
	}
	interpolatedSails.Polars = append(interpolatedSails.Polars, sail.Polars[len(sail.Polars)-1])

	if !fake {
		if err = interpolatedSails.pushWindsRedis(c, redisName+"_winds"); err != nil {
			log.Fatal(err)
		}
		if err = interpolatedSails.pushAnglesRedis(c, redisName+"_angles"); err != nil {
			log.Fatal(err)
		}

		if err = interpolatedSails.pushSpeedsRedis(c, redisName+"_speeds"); err != nil {
			log.Fatal(err)
		}
		/* TODO : store all polars */
	} else {
		fmt.Printf("%v+\n", interpolatedSails)
	}

	return err
}

func (sail SailCharacteristic) pushWindsRedis(c *redis.Client, redisName string) (err error) {
	c.Del(redisName)
	args := make([]interface{}, len(sail.Winds))
	for i, value := range sail.Winds {
		args[i] = interface{}(value)
	}
	cmd := c.LPush(redisName, args...)
	return cmd.Err()
}

func (sail SailCharacteristic) pushAnglesRedis(c *redis.Client, redisName string) (err error) {
	c.Del(redisName)
	args := make([]interface{}, len(sail.Polars))
	for i, polar := range sail.Polars {
		args[i] = interface{}(polar.Angle)
	}
	cmd := c.LPush(redisName, args...)
	return cmd.Err()
}

func (sail SailCharacteristic) pushSpeedsRedis(c *redis.Client, redisName string) (err error) {
	c.Del(redisName)
	args := make([]interface{}, len(sail.Polars))
	for i, polar := range sail.Polars {
		table := ""
		for j, value := range polar.Speed {
			if j > 0 {
				table = table + ","
			}
			table = table + strconv.FormatFloat(value, 'f', -1, 64)
		}
		args[i] = interface{}(table)
	}
	cmd := c.LPush(redisName, args...)
	return cmd.Err()
}
