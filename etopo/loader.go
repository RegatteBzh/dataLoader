package etopo

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"strconv"

	"github.com/sethgrid/multibar"

	"gopkg.in/redis.v3"
)

// DATASIZE is the size of one datum
const DATASIZE = 2 // sizeof int16

// Loader load Etopo data to Redis
func Loader(file io.Reader, width int, height int, c *redis.Client, redisName string, threshold int16, progressBar multibar.ProgressFunc) (err error) {
	fileLine := make([]byte, width*DATASIZE)

	c.Del(redisName)

	progressBar(0)

	for line := 0; line < height; line++ {

		// read line in file
		if _, err = file.Read(fileLine); err != nil {
			log.Fatal("etopo1: file.Read failed\n", err)
		}

		// convert in an array of int16
		data := make([]int16, width)
		dataBuf := bytes.NewReader(fileLine)
		if err := binary.Read(dataBuf, binary.LittleEndian, data); err != nil {
			log.Fatal("Byte to int16 failed\n", err)
		}

		for column, altitude := range data {
			if altitude > threshold {
				lat := 90 - float64(180)*float64(column)/float64(height-1)
				lon := float64(360)*float64(line)/float64(width-1) - 180
				loc := redis.GeoLocation{
					Longitude: lon,
					Latitude:  lat,
					Name:      strconv.Itoa(int(altitude)),
				}
				c.GeoAdd(redisName, &loc)
			}
		}
		progressBar(line)
	}

	return
}
