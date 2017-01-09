package polar

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
