package util

func ContainsString(l []string, s string) bool {
	for _, n := range l {
		if s == n {
			return true
		}
	}

	return false
}
