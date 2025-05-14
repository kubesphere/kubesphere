package util

import (
	"math"
	"slices"
	"unsafe"
)

// NewPtrSlice returns a slice of pointers to T with length n,
// with only 2 allocations performed no matter the size of n.
// See:
// https://gist.github.com/CAFxX/e96e8a5c3841d152f16d266a1fe7f8bd#slices-of-pointers
func NewPtrSlice[T any](n int) []*T {
	return GrowPtrSlice[T](nil, n)
}

// GrowPtrSlice appends n elements to the slice, each pointing to
// a newly-allocated T. The resulting slice has length equal to len(s)+n.
//
// It performs at most 2 allocations, regardless of n.
func GrowPtrSlice[T any](s []*T, n int) []*T {
	s = slices.Grow(s, n)
	p := make([]T, n)
	for i := range n {
		s = append(s, &p[i])
	}
	return s
}

// Allocation free conversion from []byte to string (unsafe)
// Note that the byte slice must not be modified after conversion
func ByteSliceToString(bs []byte) string {
	return unsafe.String(unsafe.SliceData(bs), len(bs))
}

// Allocation free conversion from ~string to []byte (unsafe)
// Note that the byte slice must not be modified after conversion
func StringToByteSlice[T ~string](s T) []byte {
	return unsafe.Slice(unsafe.StringData(string(s)), len(s))
}

// NumDigitsInt returns the number of digits in n.
// This is useful for pre-allocating buffers for string conversion.
func NumDigitsInt(n int) int {
	if n == 0 {
		return 1
	}

	if n < 0 {
		n = -n
	}

	return int(math.Log10(float64(n))) + 1
}

// NumDigitsUint returns the number of digits in n.
// This is useful for pre-allocating buffers for string conversion.
func NumDigitsUint(n uint64) int {
	if n == 0 {
		return 1
	}

	return int(math.Log10(float64(n))) + 1
}
