package wind

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"

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

	//windMap := mapper.New(image.Rect(0, 0, 360, 181), 60, 60)

	progress := 1
	lat := 90
	lon := 0
	for i := uint32(0); i < count; i++ {
		speed := Speed{
			SpeedU: windsU[i],
			SpeedV: windsV[i],
		}

		jsonSpeed, err := json.Marshal(speed)
		if err != nil {
			log.Fatal("Cannot convert wind speed to json")
		}

		var loc redis.GeoLocation
		if lat >= -85 && lat <= 85 {
			if lon >= 180 {
				// lon[180 - 359] => x=[0-179]
				loc = redis.GeoLocation{
					Longitude: float64(lon) - 360,
					Latitude:  float64(lat),
					Name:      string(jsonSpeed),
				}

			} else {
				// lon[0 - 179] => x=[180-359]
				loc = redis.GeoLocation{
					Longitude: float64(lon),
					Latitude:  float64(lat),
					Name:      string(jsonSpeed),
				}
			}

			if !fake {
				c.GeoAdd(redisName, &loc)
			} else {
				fmt.Printf("[%.03f, %.03f] %s\n", loc.Longitude, loc.Latitude, loc.Name)
			}
		}

		lon++

		if lon >= 360 {
			lat--
			lon = 0
			progress++
			if !fake {
				progressBar(progress)
			}
		}
	}

	return err
}
