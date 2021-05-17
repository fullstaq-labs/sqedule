package lib

const MaxUint64 = ^uint64(0)
const MaxInt64 = int64(MaxUint64 >> 1)

// AddAndClamp calculates `a + b` in an overflow-proof manner, and ensures that the result
// is between `[min, max]`.
func AddAndClamp(a uint64, b int64, min uint64, max uint64) uint64 {
	var result uint64

	if b < 0 {
		if a < uint64(-b) {
			result = 0
		} else {
			result = a - uint64(-b)
		}
	} else {
		result = a + uint64(b)
		if result < a {
			result = MaxUint64
		}
	}

	if result < min {
		result = min
	} else if result > max {
		result = max
	}

	return result
}
