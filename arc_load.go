package arclib

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
)

// LoadFromFile reads a file and passes its contents to Load.
func LoadFromFile(path string) (*ARC, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return Load(contents)
}

// Load takes the given ARC and breaks it down into a more easily dealt with format.
func Load(contents []byte) (*ARC, error) {
	m := miniRead{
		contents: contents,
	}

	// Read the header so we know what's where.
	// The header takes up the first 0x20/32 bytes of the file.
	var header arcHeader
	err := readToType(m.readLen(32), &header)
	if err != nil {
		return nil, err
	}

	// Simple sanity check.
	if header.Magic != ARCHeader {
		return nil, ErrInvalidMagic
	}

	// Read the root node. It resides 32 bytes in and has a length of 12.
	// It should also be a directory.
	var rootNode arcNode
	err = readToType(m.readLen(12), &rootNode)
	if rootNode.Type != Directory {
		return nil, ErrInvalidRootNode
	}

	// Create our root node. This will be the top directory for our returned ARC.
	result := ARCDir{
		childCount: rootNode.Size,
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
	// The first directory will always be our root node.
	directories := []ARCDir{result}

	var size uint32
	for size = 0; size != rootNode.Size; {
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
			return nil, err
		}

		// Ensure it is a type we know.
		switch currentNode.Type {
		case Directory:
		case File:
			break
		default:
			return nil, ErrUnknownNode
		}

		// If this is a directory, we need to keep track of it for path purposes.
		if currentNode.Type == Directory {
			dir := ARCDir{
				Filename:   currentNode.name(stringTable),
				childCount: currentNode.Size,
			}
			directories = append(directories, dir)
		}

		// Add data to the highest directory if it is a file.
		if currentNode.Type == File {
			contentOffset := currentNode.DataOffset
			data := contents[contentOffset : contentOffset+currentNode.Size]
			file := ARCFile{
				Filename: currentNode.name(stringTable),
				Data:     data,
				Length:   int(currentNode.Size),
			}

			// Determine the highest directory.
			directories[len(directories)-1].AddFile(file)
		}

		// Evaluate if we need to remove any directories by size.
		// If the current size is equivalent to their size, they will contain no other contents.
		// We loop in reverse to properly remove contents.
		for index := len(directories) - 1; index >= 0; index-- {
			currentDir := directories[index]

			// We cannot close the last directory if it is present.
			// Ensure we have more than one directory before closing (the expected root node)
			if currentDir.childCount == size && len(directories) > 1 {
				// Add to the parent directory for hierarchy.
				directories[index-1].AddDir(currentDir)
				directories = removeDir(directories, index)
			}
		}

		if rootNode.Size == size {
			// We have finished iterating through all children.
			break
		}
	}

	return &ARC{directories[0]}, nil
}

// removeDir removes an element from our directory slice.
func removeDir(dirs []ARCDir, pos int) []ARCDir {
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
