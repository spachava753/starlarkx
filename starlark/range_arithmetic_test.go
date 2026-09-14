package starlark

import (
	"math/big"
	"testing"
)

func TestRangeLenBoundaries(t *testing.T) {
	hi := int(^uint(0) >> 1)
	lo := -hi - 1
	bounds := []int{lo, lo + 1, -2, -1, 0, 1, 2, hi - 1, hi}
	for _, start := range bounds {
		for _, stop := range bounds {
			for _, step := range bounds {
				if step == 0 {
					continue
				}
				want := new(big.Int)
				if step > 0 && start < stop || step < 0 && start > stop {
					distance := new(big.Int).Sub(big.NewInt(int64(stop)), big.NewInt(int64(start)))
					distance.Abs(distance)
					stride := new(big.Int).Abs(big.NewInt(int64(step)))
					rem := new(big.Int)
					want.QuoRem(distance, stride, rem)
					if rem.Sign() != 0 {
						want.Add(want, big.NewInt(1))
					}
				}
				if got := rangeLen(start, stop, step); uint64(got) != want.Uint64() {
					t.Errorf("rangeLen(%d, %d, %d) = %d, want %s", start, stop, step, got, want)
				}
			}
		}
	}
}
