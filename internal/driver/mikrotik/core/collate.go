package core

// Printable ASCII collation order matching standard display sorting.
const asciiOrder = " _-,;:!?.'\"()[]{}@*/\\&#%`^+<=>|~$0123456789abcdefghijklmnopqrstuvwxyz"

var primaryWeight [128]int

func init() {
	for i := range primaryWeight {
		primaryWeight[i] = -1
	}
	for i, r := range asciiOrder {
		primaryWeight[r] = i + 1
		if r >= 'a' && r <= 'z' {
			primaryWeight[r-32] = i + 1
		}
	}
}

func primary(r rune) int {
	if r < 128 && primaryWeight[r] >= 0 {
		return primaryWeight[r]
	}
	return len(asciiOrder) + 1 + int(r)
}

func upper(r rune) int {
	if r >= 'A' && r <= 'Z' {
		return 1
	}
	return 0
}

// Collate compares string a and b in natural collation order (-1, 0, or 1).
func Collate(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	n := len(ra)
	if len(rb) < n {
		n = len(rb)
	}
	for i := 0; i < n; i++ {
		wa, wb := primary(ra[i]), primary(rb[i])
		if wa != wb {
			if wa < wb {
				return -1
			}
			return 1
		}
	}
	if len(ra) != len(rb) {
		if len(ra) < len(rb) {
			return -1
		}
		return 1
	}
	for i := 0; i < n; i++ {
		ca, cb := upper(ra[i]), upper(rb[i])
		if ca != cb {
			if ca < cb {
				return -1
			}
			return 1
		}
	}
	return 0
}
