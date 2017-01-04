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
	Winds  []float64
	Polars []Polar
}

func knotToMeter(knot float64) float64 {
	return knot * float64(0.514444)
}

func loader(csvFile io.Reader, c *redis.Client, redisName string, progressBar multibar.ProgressFunc, fake bool) error {

	fmt.Printf("%s\n", redisName)

	reader := csv.NewReader(csvFile)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1

	csvData, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("WindLevel:\n")
	for _, wind := range csvData[0][1:] {
		windLevel, err := strconv.ParseFloat(wind, 32)
		if err != nil {
			log.Fatal("Error parsing wind data")
		}
		// knot to m/s conversion
		windLevel = knotToMeter(windLevel)

		fmt.Printf("%f ", windLevel)
	}

	for index, polarSample := range csvData {
		//skip the firt record
		if index == 0 {
			continue
		}

		angle, err := strconv.ParseFloat(polarSample[0], 32)
		if err != nil {
			log.Fatal("Error parsing wind angle")
		}
		fmt.Printf("Angle: %f : ", angle)

		for i, speed := range polarSample {
			if i > 0 {
				newPolar, err := strconv.ParseFloat(speed, 32)
				if err != nil {
					log.Fatal("Error parsing wind speed")
				}
				// knot to m/s conversion
				newPolar = knotToMeter(newPolar)

				fmt.Printf("%f ", newPolar)
			}
		}
	}

	return nil
}
