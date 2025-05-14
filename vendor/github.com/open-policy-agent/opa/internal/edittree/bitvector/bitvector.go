// Package bitvector provides the implementation of a variable sized compact vector of bits
// which supports lookups, sets, appends, insertions, and deletions.
package bitvector

// A BitVector is a variable sized vector of bits. It supports
// lookups, sets, appends, insertions, and deletions.
//
// This class is not thread safe.
type BitVector struct {
	data   []byte
	length int
}

// NewBitVector creates and initializes a new bit vector with length
// elements, using data as its initial contents.
func NewBitVector(data []byte, length int) *BitVector {
	return &BitVector{
		data:   data,
		length: length,
	}
}

// Bytes returns a slice of the contents of the bit vector. If the caller changes the returned slice,
// the contents of the bit vector may change.
func (vector *BitVector) Bytes() []byte {
	return vector.data
}

// Length returns the current number of elements in the bit vector.
func (vector *BitVector) Length() int {
	return vector.length
}

// This function shifts a byte slice one bit lower (less significant).
// bit (either 1 or 0) contains the bit to put in the most significant
// position of the last byte in the slice.
// This returns the bit that was shifted off of the last byte.
func shiftLower(bit byte, b []byte) byte {
	bit <<= 7
	for i := len(b) - 1; i >= 0; i-- {
		newByte := b[i] >> 1
		newByte |= bit
		bit = (b[i] & 1) << 7
		b[i] = newByte
	}
	return bit >> 7
}

// This function shifts a byte slice one bit higher (more significant).
// bit (either 1 or 0) contains the bit to put in the least significant
// position of the first byte in the slice.
// This returns the bit that was shifted off the last byte.
func shiftHigher(bit byte, b []byte) byte {
	for i := 0; i < len(b); i++ {
		newByte := b[i] << 1
		newByte |= bit
		bit = (b[i] & 0x80) >> 7
		b[i] = newByte
	}
	return bit
}

// Returns the minimum number of bytes needed for storing the bit vector.
func (vector *BitVector) bytesLength() int {
	lastBitIndex := vector.length - 1
	lastByteIndex := lastBitIndex >> 3
	return lastByteIndex + 1
}

// Panics if the given index is not within the bounds of the bit vector.
func (vector *BitVector) indexAssert(i int) {
	if i < 0 || i >= vector.length {
		panic("Attempted to access element outside buffer")
	}
}

// Append adds a bit to the end of a bit vector.
func (vector *BitVector) Append(bit byte) {
	index := uint32(vector.length)
	vector.length++

	if vector.bytesLength() > len(vector.data) {
		vector.data = append(vector.data, 0)
	}

	byteIndex := index >> 3
	byteOffset := index % 8
	oldByte := vector.data[byteIndex]
	var newByte byte
	if bit == 1 {
		newByte = oldByte | 1<<byteOffset
	} else {
		// Set all bits except the byteOffset
		mask := byte(^(1 << byteOffset))
		newByte = oldByte & mask
	}

	vector.data[byteIndex] = newByte
}

// Element returns the bit in the ith index of the bit vector.
// Returned value is either 1 or 0.
func (vector *BitVector) Element(i int) byte {
	vector.indexAssert(i)
	byteIndex := i >> 3
	byteOffset := uint32(i % 8)
	b := vector.data[byteIndex]
	// Check the offset bit
	return (b >> byteOffset) & 1
}

// Set changes the bit in the ith index of the bit vector to the value specified in
// bit.
func (vector *BitVector) Set(bit byte, index int) {
	vector.indexAssert(index)
	byteIndex := uint32(index >> 3)
	byteOffset := uint32(index % 8)

	oldByte := vector.data[byteIndex]

	var newByte byte
	if bit == 1 {
		// turn on the byteOffset'th bit
		newByte = oldByte | 1<<byteOffset
	} else {
		// turn off the byteOffset'th bit
		removeMask := byte(^(1 << byteOffset))
		newByte = oldByte & removeMask
	}
	vector.data[byteIndex] = newByte
}

// Insert inserts bit into the supplied index of the bit vector. All
// bits in positions greater than or equal to index before the call will
// be shifted up by one.
func (vector *BitVector) Insert(bit byte, index int) {
	vector.indexAssert(index)
	vector.length++

	// Append an additional byte if necessary.
	if vector.bytesLength() > len(vector.data) {
		vector.data = append(vector.data, 0)
	}

	byteIndex := uint32(index >> 3)
	byteOffset := uint32(index % 8)
	var bitToInsert byte
	if bit == 1 {
		bitToInsert = 1 << byteOffset
	}

	oldByte := vector.data[byteIndex]
	// This bit will need to be shifted into the next byte
	leftoverBit := (oldByte & 0x80) >> 7
	// Make masks to pull off the bits below and above byteOffset
	// This mask has the byteOffset lowest bits set.
	bottomMask := byte((1 << byteOffset) - 1)
	// This mask has the 8 - byteOffset top bits set.
	topMask := ^bottomMask
	top := (oldByte & topMask) << 1
	newByte := bitToInsert | (oldByte & bottomMask) | top

	vector.data[byteIndex] = newByte
	// Shift the rest of the bytes in the slice one higher, append
	// the leftoverBit obtained above.
	shiftHigher(leftoverBit, vector.data[byteIndex+1:])
}

// Delete removes the bit in the supplied index of the bit vector. All
// bits in positions greater than or equal to index before the call will
// be shifted down by one.
func (vector *BitVector) Delete(index int) {
	vector.indexAssert(index)
	vector.length--
	byteIndex := uint32(index >> 3)
	byteOffset := uint32(index % 8)

	oldByte := vector.data[byteIndex]

	// Shift all the bytes above the byte we're modifying, return the
	// leftover bit to include in the byte we're modifying.
	bit := shiftLower(0, vector.data[byteIndex+1:])

	// Modify oldByte.
	// At a high level, we want to select the bits above byteOffset,
	// and shift them down by one, removing the bit at byteOffset.

	// This selects the bottom bits
	bottomMask := byte((1 << byteOffset) - 1)
	// This selects the top (8 - byteOffset - 1) bits
	topMask := byte(^((1 << (byteOffset + 1)) - 1))
	// newTop is the top bits, shifted down one, combined with the leftover bit from shifting
	// the other bytes.
	newTop := (oldByte&topMask)>>1 | (bit << 7)
	// newByte takes the bottom bits and combines with the new top.
	newByte := (bottomMask & oldByte) | newTop
	vector.data[byteIndex] = newByte

	// The desired length is the byte index of the last element plus one,
	// where the byte index of the last element is the bit index of the last
	// element divided by 8.
	byteLength := vector.bytesLength()
	if byteLength < len(vector.data) {
		vector.data = vector.data[:byteLength]
	}
}
