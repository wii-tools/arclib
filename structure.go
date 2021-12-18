package arclib

import (
	"encoding/binary"
)

type arcHeader struct {
	Magic          uint32
	RootNodeOffset uint32
	HeaderSize     uint32
	DataOffset     uint32
	// Padding
	_ [16]byte
}

type arcNode struct {
	Type       ARCType
	NameOffset [3]byte
	DataOffset uint32
	Size       uint32
}

type ARCType uint8

const (
	File ARCType = iota
	Directory
)

// name returns the name as specified with offsets specified by the node.
// It reads from the specified string table.
func (a *arcNode) name(table []byte) string {
	// Nintendo gives us 3 bytes, a "uint24" if you will.
	// Alright then... we need to normalize this.

	// The Wii uses big-endian throughout every possible corner.
	// This means that we can safely convert this to a uint32
	// by inserting a null byte in front.
	offset := a.NameOffset
	posByte := []byte{0x00, offset[0], offset[1], offset[2]}
	pos := binary.BigEndian.Uint32(posByte)

	// Now that we have this position, we can read in from this
	// position to the first null byte.
	// I believe Nintendo uses strcpy to achieve this in C.
	var tmp []byte
	for {
		current := table[pos : pos+1][0]
		if current == 0x00 {
			// We completed our string!
			break
		}

		// Add this current byte to our array.
		tmp = append(tmp, current)
		pos++
	}

	return string(tmp)
}
