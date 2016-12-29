package gshhg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"github.com/sethgrid/multibar"
	redis "gopkg.in/redis.v4"
)

// Point is a point in polygons
type Point struct {
	Lon, Lat float64
}

// Polygon is the definition of a polygon
type Polygon struct {
	ID        int32
	Count     int32
	Flag      int32
	West      float64
	East      float64
	South     float64
	North     float64
	Area      float64
	AreaFull  float64
	Container int32
	Ancestor  int32
	Points    []Point
}

// IsInside check if a point is inside a polygon
func (polygon Polygon) IsInside(point Point) bool {
	var firstPoint Point
	intersection := 0
	// inside frame ?
	if point.Lon < polygon.West || point.Lon > polygon.East || point.Lat > polygon.North || point.Lat < polygon.South {
		return false
	}

	for index, secondPoint := range polygon.Points {
		if index > 0 && secondPoint.Lon != firstPoint.Lon {
			slope := (secondPoint.Lat - firstPoint.Lat) / (secondPoint.Lon - firstPoint.Lon)
			offset := (firstPoint.Lat*(secondPoint.Lon-firstPoint.Lon) - firstPoint.Lon*(secondPoint.Lat-firstPoint.Lat)) / (secondPoint.Lon - firstPoint.Lon)
			projectionLat := slope*point.Lon + offset
			if projectionLat < point.Lat {
				intersection++
			}
		}
		firstPoint = secondPoint
	}
	return intersection%2 == 1
}

func readPolygon(file io.Reader) (polygon Polygon, err error) {
	header := make([]byte, 44)

	if _, err = file.Read(header); err != nil {
		if err.Error() == "EOF" {
			return
		}
		log.Fatal("GSHHG header: file.Read failed\n", err)
	}

	// convert in an array of int16
	dataHeader := make([]int32, 11)
	dataHeaderBuf := bytes.NewReader(header)
	if err := binary.Read(dataHeaderBuf, binary.BigEndian, dataHeader); err != nil {
		log.Fatal("Byte to int16 failed\n", err)
	}

	polygon = Polygon{
		ID:        dataHeader[0],
		Count:     dataHeader[1],
		Flag:      dataHeader[2],
		West:      float64(dataHeader[3]) / 1000000,
		East:      float64(dataHeader[4]) / 1000000,
		South:     float64(dataHeader[5]) / 1000000,
		North:     float64(dataHeader[6]) / 1000000,
		Area:      float64(dataHeader[7]) / 10,
		AreaFull:  float64(dataHeader[8]) / 10,
		Container: dataHeader[9],
		Ancestor:  dataHeader[10],
		Points:    make([]Point, dataHeader[1]),
	}

	pointBytes := make([]byte, 8*polygon.Count)
	if _, err = file.Read(pointBytes); err != nil {
		log.Fatal("GSHHG points: file.Read failed\n", err)
	}

	dataPoints := make([]int32, 2*polygon.Count)
	dataPointsBuf := bytes.NewReader(pointBytes)
	if err := binary.Read(dataPointsBuf, binary.BigEndian, dataPoints); err != nil {
		log.Fatal("Byte to int16 failed\n", err)
	}

	for i := int32(0); i < polygon.Count; i++ {
		polygon.Points[i].Lon = float64(dataPoints[i*2]) / 1000000
		polygon.Points[i].Lat = float64(dataPoints[i*2+1]) / 1000000
	}

	return
}

func getOne(file io.Reader) error {
	if polygon, err := readPolygon(file); err == nil {
		fmt.Printf("%d => %d elements\tArea: %.01fkm²\n\tWest: %.03f\n\tEast: %.03f\n\tSouth: %.03f\n\tNorth: %.03f\n", polygon.ID, polygon.Count, polygon.Area, polygon.West, polygon.East, polygon.South, polygon.North)
	} else {
		return err
	}
	return nil
}

// Loader load GSHHG file
func Loader(file io.Reader, c *redis.Client, progressBar multibar.ProgressFunc, fake bool) (err error) {

	for i := 0; i < 1000000; i++ {
		if err = getOne(file); err != nil {
			if err.Error() == "EOF" {
				fmt.Printf("Done!\n")
				return
			}
			log.Fatal("Shit !", err)
		}

	}

	return err
}
