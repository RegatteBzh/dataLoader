package polar

import (
	"encoding/csv"
	"io"
	"log"
	"strconv"

	"github.com/sethgrid/multibar"
)

func csvLoader(csvFile io.Reader, redisName string, progressBar multibar.ProgressFunc, fake bool) (sailChar SailCharacteristic, err error) {

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
