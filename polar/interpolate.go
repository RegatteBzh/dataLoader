package polar

func (first Polar) interpolate(second Polar) (result []Polar) {
	result = append(result, first)
	for angle := first.Angle + 1; angle < second.Angle; angle++ {
		percent := (angle-first.Angle)/second.Angle - first.Angle
		current := Polar{
			Angle: angle,
			Speed: make([]float64, len(first.Speed)),
		}
		for i, firstSpeed := range first.Speed {
			current.Speed[i] = firstSpeed + percent*(second.Speed[i]-firstSpeed)
		}
		result = append(result, current)
	}
	return
}
