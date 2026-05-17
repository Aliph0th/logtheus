package utils

import "math"

func NormalizeFloat(value float64) interface{} {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return nil
	}

	return value
}
