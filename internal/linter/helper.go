package linter

// ToPerc computes the percentage from one number over another.
func ToPerc(v1, v2 float64) float64 {
	if v2 == 0 {
		return 0
	}
	return (v1 / v2) * 100
}

// In checks if a string is in a list of strings.
func in(ll []string, s string) bool {
	for _, l := range ll {
		if l == s {
			return true
		}
	}

	return false
}
