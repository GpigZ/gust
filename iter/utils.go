package iter

import (
	"math"

	"github.com/andeya/gust"
)

func saturatingAdd(a, b uint) uint {
	if a < math.MaxUint-b {
		return a + b
	}
	return math.MaxUint
}

func checkedAdd(a, b uint) gust.Option[uint] {
	if a <= math.MaxUint-b {
		return gust.Some(a + b)
	}
	return gust.None[uint]()
}

func uintCheckedMul(a, b uint) gust.Option[uint] {
	if a <= math.MaxUint/b {
		return gust.Some(a * b)
	}
	return gust.None[uint]()
}
