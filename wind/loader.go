package wind

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"

	"errors"

	"github.com/sethgrid/multibar"
	redis "gopkg.in/redis.v4"
)

// Speed is the speed of the wind (m/sec)
type Speed struct {
	SpeedU float32 `json:"speedU"`
	SpeedV float32 `json:"speedV"`
}

func readBlock(file io.Reader) (count uint32, winds []float32, err error) {
	size := make([]byte, 4)

	// read data block
	if _, err = file.Read(size); err != nil {
		log.Fatal("file.Read failed (ReadBlock)\n", err)
		return
	}
	countBuf := bytes.NewReader(size)
	if err = binary.Read(countBuf, binary.LittleEndian, &count); err != nil {
		log.Fatal("Byte to uint32 failed\n", err)
		return
	}

	// read data
	windb := make([]byte, count)
	if _, err = file.Read(windb); err != nil {
		log.Fatal("file.Read failed (ReadBlock)\n", err)
		return
	}

	// read the size again
	if _, err = file.Read(size); err != nil {
		log.Fatal("file.Read failed (ReadBlock)\n", err)
		return
	}

	count = count / 4

	winds = make([]float32, count)
	windsBuf := bytes.NewReader(windb)
	if err = binary.Read(windsBuf, binary.LittleEndian, &winds); err != nil {
		log.Fatal("Byte to float32 failed\n", err)
		return
	}

	return
}

// Loader reads wind from binary file and store to redis
func Loader(file io.Reader, c *redis.Client, redisName string, progressBar multibar.ProgressFunc, fake bool) (err error) {
	countU, windsU, err := readBlock(file)
	countV, windsV, err := readBlock(file)

	count := countU
	if count > countV {
		count = countV
	}

	for i := uint32(0); i < count; i++ {
		speed := Speed{
			SpeedU: windsU[i],
			SpeedV: windsV[i],
		}

		jsonSpeed, err := json.Marshal(speed)
		if err != nil {
			log.Fatal("Cannot convert wind speed to json")
		}

		if loc, err := makePoint(string(jsonSpeed), i); err != nil {
			if !fake {
				c.GeoAdd(redisName, &loc)
			} else {
				fmt.Printf("[%.03f, %.03f] %s\n", loc.Longitude, loc.Latitude, loc.Name)
			}
		}

		if i%360 == 0 {
			progressBar(int(i) / 360)
		}
	}

	return err
}

func makePoint(name string, index uint32) (redis.GeoLocation, error) {
	lon := index % 360
	lat := 90 - int(index)/360

	if lat >= -85 && lat <= 85 {
		if lon >= 180 {
			// lon[180 - 359] => x=[0-179]
			return redis.GeoLocation{
				Longitude: float64(lon) - 360,
				Latitude:  float64(lat),
				Name:      name,
			}, nil

		}
		// lon[0 - 179] => x=[180-359]
		return redis.GeoLocation{
			Longitude: float64(lon),
			Latitude:  float64(lat),
			Name:      name,
		}, nil
	}
	return redis.GeoLocation{}, errors.New("out of bounds")
}

// LoadAll loads all weather forecast
func LoadAll(files map[string]*fileLoader, client *redis.Client, redisName string, progressBar multibar.ProgressFunc, fake bool) (err error) {
	var count uint32

	if !fake {
		client.Del(redisName)
		progressBar(0)
	}

	for _, descriptor := range files {
		countU, windsU, errU := readBlock(descriptor.Handler)
		if errU != nil {
			log.Fatal(errU)
		}
		_, windsV, errV := readBlock(descriptor.Handler)
		if errV != nil {
			log.Fatal(errV)
		}
		count = countU
		descriptor.WindsU = windsU
		descriptor.WindsV = windsV
	}

	for i := uint32(0); i < count; i++ {

		speeds := make(map[int]Speed)

		for index, descriptor := range files {
			indexInt, err := strconv.Atoi(index)
			if err != nil {
				log.Fatal(err)
			}
			speeds[indexInt] = Speed{
				SpeedU: descriptor.WindsU[i],
				SpeedV: descriptor.WindsV[i],
			}
		}

		jsonSpeed, err := json.Marshal(speeds)
		if err != nil {
			log.Fatal("Cannot convert wind speed to json")
		}

		if loc, err := makePoint(string(jsonSpeed), i); err == nil {
			if !fake {
				client.GeoAdd(redisName, &loc)
			} else {
				fmt.Printf("[%.03f, %.03f] %s\n", loc.Longitude, loc.Latitude, loc.Name)
			}
		}

		if i%360 == 0 && !fake {
			progressBar(int(i) / 360)
		}
	}

	return
}
