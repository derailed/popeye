package linter

func toPerc(v1, v2 float64) float64 {
	if v2 == 0 {
		return 0
	}
	return (v1 / v2) * 100
}
