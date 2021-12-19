package arclib

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
)

// miniMuxer assists with serialization of separate components of an ARC.
type miniMuxer struct {
	records []arcNode
	strings []byte
	data    []byte

	recordCount uint32
}

// addString adds the given string to the strings table.
// It returns the offset within for usage within node tracking.
func (m *miniMuxer) addString(value string) [3]byte {
	pos := len(m.strings)
	m.strings = append(m.strings, []byte(value)...)
	// Null terminate the given string.
	m.strings = append(m.strings, 0x00)

	// The Wii assumes big-endian on all types within.
	bytePos := make([]byte, 4)
	binary.BigEndian.PutUint32(bytePos, uint32(pos))

	// We can simply remove the leading byte (0x00, presumably).
	var result [3]byte
	copy(result[:], bytePos[1:4])
	return result
}

// addData adds the given string to the data table.
func (m *miniMuxer) addData(contents []byte) uint32 {
	pos := len(m.data)
	m.data = append(m.data, contents...)
	return uint32(pos)
}

// addRecord tracks a given record for muxing later.
func (m *miniMuxer) addRecord(record arcNode) {
	m.recordCount += 1
	m.records = append(m.records, record)
}

// addFile tracks a file as a record.
func (m *miniMuxer) addFile(file ARCFile) {
	dataPos := m.addData(file.Data)
	strPos := m.addString(file.Filename)

	m.addRecord(arcNode{
		Type:       File,
		NameOffset: strPos,
		DataOffset: dataPos,
		Size:       uint32(file.Size()),
	})
}

// addDir tracks a dir as a record.
func (m *miniMuxer) addDir(dir ARCDir) {
	// The size of a directory is the count of all children it will contain.
	size := dir.RecursiveCount()

	pos := m.addString(dir.Filename)
	m.addRecord(arcNode{
		Type:       Directory,
		NameOffset: pos,
		Size:       m.recordCount + uint32(size),
	})
}

// recordsToBytes serializes all tracked records into bytes, or returns an error.
func (m *miniMuxer) recordsToBytes() ([]byte, error) {
	var working []byte

	for _, record := range m.records {
		// Serialize to bytes.
		result, err := writeToBytes(record)
		if err != nil {
			return nil, err
		}

		working = append(working, result...)
	}

	return working, nil
}

// recurseDir iterates through the given directory and creates records for everything within.
func (m *miniMuxer) recurseDir(dir ARCDir) {
	// Add this directory's record.
	m.addDir(dir)

	// Handle sub-directories first.
	for _, subDir := range dir.Subdirs {
		m.recurseDir(subDir)
	}

	// Lastly, add files.
	for _, file := range dir.Files {
		m.addFile(file)
	}
}

// Save converts the given ARC hierarchy to a usable binary file.
func (a *ARC) Save() ([]byte, error) {
	m := new(miniMuxer)

	// Iterate through our hierarchy and record binary types.
	m.recurseDir(a.RootRecord)

	// Records are 12 bytes each. The count of all records * 12 represents their length.
	recordLen := 12 * len(m.records)
	headerSize := recordLen + len(m.strings)

	// Our data offset is the header size (32) + size of records and strings.
	dataOffset := 32 + headerSize

	header := arcHeader{
		Magic: ARCHeader,
		// Our root node 32 bytes.
		RootNodeOffset: 32,
		HeaderSize:     uint32(headerSize),
		DataOffset:     uint32(dataOffset),
	}

	// Update all data offsets now that we have calculated the proper offset.
	// Data offsets are the offset from the very start of the archive.
	for idx, record := range m.records {
		if record.Type == File {
			record.DataOffset += uint32(dataOffset)
		}

		m.records[idx] = record
	}
	records, err := m.recordsToBytes()
	if err != nil {
		return nil, err
	}

	// Serialize our header for writing.
	headerBytes, err := writeToBytes(header)
	if err != nil {
		return nil, err
	}

	var endFile []byte
	endFile = append(endFile, headerBytes...)
	endFile = append(endFile, records...)
	endFile = append(endFile, m.strings...)
	endFile = append(endFile, m.data...)

	return endFile, nil
}

// SaveToFile writes a serialized ARC to the specified path.
func (a *ARC) SaveToFile(path string) error {
	contents, err := a.Save()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, contents, 0755)
}

// writeToBytes takes a source and returns its bytes, or an error.
func writeToBytes(src interface{}) ([]byte, error) {
	w := bytes.NewBuffer([]byte{})
	err := binary.Write(w, binary.BigEndian, src)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}
