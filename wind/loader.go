package wind

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"

	"github.com/sethgrid/multibar"
	redis "gopkg.in/redis.v3"
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
func Loader(file io.Reader, c *redis.Client, redisName string, progressBar multibar.ProgressFunc) (err error) {
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

		if lon >= 180 {
			// lon[180 - 359] => x=[0-179]
			loc := redis.GeoLocation{
				Longitude: float64(lon) - 180,
				Latitude:  float64(lat),
				Name:      string(jsonSpeed),
			}
			c.GeoAdd(redisName, &loc)

		} else {
			// lon[0 - 179] => x=[180-359]
			loc := redis.GeoLocation{
				Longitude: float64(lon) + 180,
				Latitude:  float64(lat),
				Name:      string(jsonSpeed),
			}
			c.GeoAdd(redisName, &loc)
		}

		lon = lon + 1

		if lon >= 360 {
			lat--
			lon = 0
			progress++
			progressBar(progress)
		}
	}

	return err
}
