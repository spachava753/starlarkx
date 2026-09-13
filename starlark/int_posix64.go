//go:build (linux || darwin || dragonfly || freebsd || netbsd || solaris) && (amd64 || arm64 || mips64x || ppc64 || ppc64le || loong64 || s390x)

package starlark

// This file defines an optimized Int implementation for 64-bit machines
// running POSIX. It reserves a 4GB portion of the address space using
// mmap and represents int32 values as addresses within that range. This
// disambiguates int32 values from *big.Int pointers, letting all Int
// values be represented as an unsafe.Pointer, so that Int-to-Value
// interface conversion need not allocate.

// Although iOS (which, like macOS, appears as darwin/arm64) is
// POSIX-compliant, it limits each process to about 700MB of virtual
// address space, which defeats the optimization.  Similarly,
// OpenBSD's default ulimit for virtual memory is a measly GB or so.
// On both those platforms the attempted optimization will fail and
// fall back to the slow implementation.

// An alternative approach to this optimization would be to embed the
// int32 values in pointers using odd values, which can be distinguished
// from (even) *big.Int pointers. However, the Go runtime does not allow
// user programs to manufacture pointers to arbitrary locations such as
// within the zero page, or non-span, non-mmap, non-stack locations,
// and it may panic if it encounters them; see Issue #382.

import (
	"log"
	"math"
	"math/big"
	"unsafe"

	"golang.org/x/sys/unix"
)

// intImpl represents a union of (int32, *big.Int) in a single pointer,
// so that Int-to-Value conversions need not allocate.
//
// The pointer is either a *big.Int, if the value is big, or a pointer into a
// reserved portion of the address space (smallints), if the value is small
// and the address space allocation succeeded.
//
// See int_generic.go for the basic representation concepts.
//
// This optimization is retained deliberately. In a Go 1.27.0 darwin/arm64
// audit on Apple M4 (three 100ms runs, rounded medians), replacing it with
// int_generic.go's safe two-field union changed these existing benchmarks:
//
//	                      current                 safe union
//	bench_int             23.8 us,    1 alloc      38.7 us, 2002 allocs
//	bench_range_iteration  1.86 us,   3 allocs      3.34 us,  204 allocs
//	bench_mix             34.1 us, 483 allocs      45.8 us, 1808 allocs
//
// Each bench_int operation performs 1000 increments; range iteration visits
// 200 elements. These are local measurements, not universal slowdown factors.
// The safe union grows Int from 8 to 16 bytes on this target and allocates when
// escaping Int values are boxed as Value. A safe pointer-backed representation
// with a bounded cache also passed the suite, but uncached values allocate and
// Go-level equality of those Int values needs separate review.
//
// Safety depends on retaining the mmap region for the process lifetime, never
// dereferencing marker pointers, and keeping marker offsets in the int32 range.
// The 4 GiB mapping reserves virtual address space, not eagerly populated RAM;
// it is not a Go heap allocation that the GC can move or reclaim. Real big.Int
// pointers remain pointers visible to the GC. Do not unmap or reuse the region.
//
// The stored-uintptr conversion in makeSmallInt intentionally triggers go vet.
// It depends on Go accepting non-heap mmap pointers, outside ordinary safe-Go
// guarantees. Targeted arithmetic and checkptr tests found no corruption in the
// audit, but do not prove those runtime assumptions valid on every platform.
type intImpl unsafe.Pointer

// get returns the (small, big) arms of the union.
func (i Int) get() (int64, *big.Int) {
	if smallints == 0 {
		// optimization disabled
		if x := (*big.Int)(i.impl); isSmall(x) {
			return x.Int64(), nil
		} else {
			return 0, x
		}
	}

	if ptr := uintptr(i.impl); ptr >= smallints && ptr < smallints+1<<32 {
		return math.MinInt32 + int64(ptr-smallints), nil
	}
	return 0, (*big.Int)(i.impl)
}

// Precondition: math.MinInt32 <= x && x <= math.MaxInt32
func makeSmallInt(x int64) Int {
	if smallints == 0 {
		// optimization disabled
		return Int{intImpl(big.NewInt(x))}
	}

	return Int{intImpl(uintptr(x-math.MinInt32) + smallints)}
}

// Precondition: x cannot be represented as int32.
func makeBigInt(x *big.Int) Int { return Int{intImpl(x)} }

// smallints is the base address of a 2^32 byte memory region.
// Pointers to addresses in this region represent int32 values.
// We assume smallints is not at the very top of the address space.
//
// Zero means the optimization is disabled and all Ints allocate a big.Int.
var smallints = reserveAddresses(1 << 32)

func reserveAddresses(len int) uintptr {
	b, err := unix.Mmap(-1, 0, len, unix.PROT_READ, unix.MAP_PRIVATE|unix.MAP_ANON)
	if err != nil {
		log.Printf("Starlark failed to allocate 4GB address space: %v. Integer performance may suffer.", err)
		return 0 // optimization disabled
	}
	return uintptr(unsafe.Pointer(&b[0]))
}
