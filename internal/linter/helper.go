package linter

// ToPerc computes the percentage from one number over another.
func ToPerc(v1, v2 float64) float64 {
	if v2 == 0 {
		return 0
	}
	return (v1 / v2) * 100
}
