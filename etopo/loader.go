package etopo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/sethgrid/multibar"

	"gopkg.in/redis.v4"
)

// DATASIZE is the size of one datum
const DATASIZE = 2 // sizeof int16

// Loader load Etopo data to Redis
func Loader(file io.Reader, width int, height int, c *redis.Client, redisName string, threshold int16, progressBar multibar.ProgressFunc, fake bool) (err error) {
	fileLine := make([]byte, width*DATASIZE)

	if !fake {
		c.Del(redisName)
		progressBar(0)
	}

	for line := 0; line < height; line++ {

		// read line in file
		if _, err = file.Read(fileLine); err != nil {
			log.Fatal(redisName+": file.Read failed\n", err)
		}

		// convert in an array of int16
		data := make([]int16, width)
		dataBuf := bytes.NewReader(fileLine)
		if err := binary.Read(dataBuf, binary.LittleEndian, data); err != nil {
			log.Fatal("Byte to int16 failed\n", err)
		}

		lat := 90 - float64(180)*float64(line)/float64(height-1)

		for column, altitude := range data {
			if altitude > threshold && lat >= -85 && lat <= 85 {
				lon := float64(360)*float64(column)/float64(width-1) - 180
				loc := redis.GeoLocation{
					Longitude: lon,
					Latitude:  lat,
					Name:      strconv.Itoa(int(altitude)),
				}
				if !fake {
					c.GeoAdd(redisName, &loc)
				} else {
					fmt.Printf("[%.03f, %.03f] %s\n", loc.Longitude, loc.Latitude, loc.Name)
				}
			}
		}

		if !fake {
			progressBar(line)
		}
	}

	return
}
