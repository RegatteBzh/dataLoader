package polar

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/sethgrid/multibar"
	redis "gopkg.in/redis.v4"
)

// Polar is a polar for a given sail and a given angle
type Polar struct {
	Angle float64
	Speed []float64
}

// SailCharacteristic is the characteristic of a sail
type SailCharacteristic struct {
	Name   string
	Winds  []float64
	Polars []Polar
}

func knotToMeter(knot float64) float64 {
	return knot * float64(0.514444)
}

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
	for i, value := range sailChar.Polars {
		args[i] = interface{}(value.Angle)
	}
	cmd := c.LPush(redisName, args...)
	return cmd.Err()
}

func loadSail(csvFile io.Reader, c *redis.Client, redisName string, progressBar multibar.ProgressFunc, fake bool) error {
	sail, err := loader(csvFile, redisName, progressBar, fake)
	if err != nil {
		log.Fatal(err)
	}

	if !fake {
		if err := sail.pushWindsRedis(c, redisName+"_winds"); err != nil {
			log.Fatal(err)
		}
		if err := sail.pushAnglesRedis(c, redisName+"_angles"); err != nil {
			log.Fatal(err)
		}
		/* TODO : store all polars */
	} else {
		fmt.Printf("%v+\n", sail)
	}

	return err
}

func loader(csvFile io.Reader, redisName string, progressBar multibar.ProgressFunc, fake bool) (sailChar SailCharacteristic, err error) {

	sailChar = SailCharacteristic{
		Name: redisName,
	}

	reader := csv.NewReader(csvFile)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1

	csvData, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	sailChar.Winds = make([]float64, len(csvData[0]))
	for windIndex, wind := range csvData[0][1:] {
		windLevel, err := strconv.ParseFloat(wind, 32)
		if err != nil {
			log.Fatal("Error parsing wind data")
		}
		// knot to m/s conversion
		sailChar.Winds[windIndex] = knotToMeter(windLevel)
	}

	sailChar.Polars = make([]Polar, len(csvData)-1) // ignore first line
	for angleIndex, polarSample := range csvData {
		//skip the firt record
		if angleIndex == 0 {
			continue
		}

		angle, err := strconv.ParseFloat(polarSample[0], 32)
		if err != nil {
			log.Fatal("Error parsing wind angle")
		}

		newPolar := Polar{
			Angle: angle,
			Speed: make([]float64, len(polarSample)-1), // ignore first column
		}

		for speedIndex, speed := range polarSample {
			if speedIndex > 0 {
				newPolarVal, err := strconv.ParseFloat(speed, 32)
				if err != nil {
					log.Fatal("Error parsing wind speed")
				}
				// knot to m/s conversion
				newPolar.Speed[speedIndex-1] = knotToMeter(newPolarVal)
			}
		}
		sailChar.Polars[angleIndex-1] = newPolar
		progressBar(int(angle))
	}
	progressBar(180)

	return
}
