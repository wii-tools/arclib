package arclib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"os"
)

var (
	ErrDirectory       = errors.New("requested path is a directory")
	ErrInvalidRootNode = errors.New("root node was not a directory")
	ErrInvalidMagic    = errors.New("invalid ARC magic")
)

type ARC struct {
	contents []ARCRecord
}

type ARCRecord struct {
	Path string
	Data []byte
	Size int
	Type ARCType
}

// LoadFromFile reads a file and passes its contents to Load.
func (a *ARC) LoadFromFile(path string) error {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return a.Load(contents)
}

// Load takes the given ARC and breaks it down into a more easily dealt with format.
func (a *ARC) Load(contents []byte) error {
	m := miniRead{
		contents: contents,
	}

	// Read the header so we know what's where.
	// The header takes up the first 0x20/32 bytes of the file.
	var header arcHeader
	err := readToType(m.readLen(32), &header)
	if err != nil {
		return err
	}

	// Simple sanity check.
	if header.Magic != 0x55AA382D {
		return ErrInvalidMagic
	}

	// Read the root node. It resides 32 bytes in and has a length of 12.
	// It should also be a directory.
	var rootNode arcNode
	err = readToType(m.readLen(12), &rootNode)
	if rootNode.Type != Directory {
		return ErrInvalidRootNode
	}

	// We now need to calculate the string offset.

	// We know the size of the header and that of the root node.
	// The root node specifies the amount of nodes it extends to,
	// so we know there will be up to that number of nodes afterwards.
	// As each node is 12 bytes, 32 + 12x will be the string table's offset.
	stringOffset := 32 + (12 * rootNode.Size)

	// To readers: The first byte of the string table is a
	// null byte! This is as the root node has no name,
	// and all names are read until the first null byte.

	// This is a generous length - it may be at max 0x40 more than we need,
	// but hopefully that's not an issue given how relatively small it is.
	stringTable := contents[stringOffset:header.DataOffset]

	// We'll store all directories we encounter.
	var directories []arcNode

	for size := 0; size != int(rootNode.Size); {
		if size == 0 {
			// This is the root node. We are not going to handle it.
			size++
			continue
		}
		size++

		// Read the current node.
		var currentNode arcNode
		err = readToType(m.readLen(12), &currentNode)
		if err != nil {
			return err
		}

		// Ensure it is a type we know.
		switch currentNode.Type {
		case Directory:
		case File:
			break
		default:
			return errors.New("unknown node type")
		}

		// Iterate through all tracked directories to build hierarchy for its path.
		path := ""
		for _, dir := range directories {
			path += dir.name(stringTable) + "/"
		}
		path += currentNode.name(stringTable)

		// If this is a directory, we need to keep track of it for path purposes.
		if currentNode.Type == Directory {
			directories = append(directories, currentNode)
		}

		// Evaluate if we need to remove any directories by size.
		// If the current size is equivalent to their size, they will contain no other contents.
		// We loop in reverse to properly remove contents.
		for index := len(directories) - 1; index >= 0; index-- {
			dir := directories[index]
			if int(dir.Size) == size {
				directories = remove(directories, index)
			}
		}

		// Finally, add this recorded type to our own format.
		record := ARCRecord{
			// We'll add its data as nil due to the fact records can include directories.
			Data: nil,
			Size: int(currentNode.Size),
			Type: currentNode.Type,
			Path: path,
		}

		// Add data if it is a file.
		if currentNode.Type == File {
			contentOffset := currentNode.DataOffset
			record.Data = contents[contentOffset : contentOffset+currentNode.Size]
		}

		a.contents = append(a.contents, record)
	}

	return nil
}

// Contents returns a list of paths of files within the ARC.
func (a *ARC) Contents() []string {
	var names []string
	for _, node := range a.contents {
		// We don't want to handle directories.
		if node.Type == Directory {
			continue
		}

		names = append(names, node.Path)
	}
	return names
}

// Read returns the contents of a file at the given path.
func (a *ARC) Read(path string) ([]byte, error) {
	for _, node := range a.contents {
		if path != node.Path {
			continue
		}

		// We don't want to handle directories.
		if node.Type == Directory {
			return nil, ErrDirectory
		}

		return node.Data, nil
	}
	return nil, os.ErrNotExist
}

// remove removes an element from our directory slice.
func remove(dirs []arcNode, pos int) []arcNode {
	return append(dirs[:pos], dirs[pos+1:]...)
}

// readToType takes a source and reads its bytes to a specified interface.
func readToType(src []byte, dst interface{}) error {
	tmp := bytes.NewBuffer(src)
	err := binary.Read(tmp, binary.BigEndian, dst)
	if err != nil {
		return err
	}
	return nil
}

type miniRead struct {
	contents []byte
	position int
}

func (m *miniRead) readLen(len int) []byte {
	contents := m.contents[m.position : m.position+len]
	m.position += len
	return contents
}
