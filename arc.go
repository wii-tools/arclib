package arclib

import (
	"errors"
	"strings"
)

var (
	ErrInvalidRootNode = errors.New("root node was not a directory")
	ErrInvalidMagic    = errors.New("invalid ARC magic")
	ErrUnknownNode     = errors.New("unknown node type")
)

// ARC describes a hierarchy suitable for serialization and deserialization of an ARC file.
type ARC struct {
	RootRecord ARCDir
}

func (a *ARC) FileAtPath(path string) (*ARCFile, error) {
	components := strings.Split(path, "/")

	var dirs []string
	var filename string

	if len(components) == 1 {
		dirs = []string{}
		filename = components[0]
	} else {
		// Directories are all but the last element in our split string.
		dirs = components[0 : len(components)-1]
		filename = components[len(components)-1]
	}

	// The root node is where we start our loop.
	current := &a.RootRecord
	for _, dirName := range dirs {
		testDir, err := current.GetDir(dirName)
		if err != nil {
			return nil, err
		}
		current = testDir
	}

	// Retrieve our file from the last directory in our hierarchy.
	return current.GetFile(filename)
}

// Read returns the contents of a file at the given path.
func (a *ARC) Read(path string) ([]byte, error) {
	file, err := a.FileAtPath(path)
	if err != nil {
		return nil, err
	}

	return file.Data, nil
}

// Size returns the size in bytes of the file for the given path.
func (a *ARC) Size(path string) (int, error) {
	file, err := a.FileAtPath(path)
	if err != nil {
		return 0, err
	}

	return file.Length, nil
}

// miniRead exists to help with iterating through our ARC.
type miniRead struct {
	contents []byte
	position int
}

// readLen reads the specified length and returns its bytes.
// Opposed to bytes.NewReader, we do not throw an error for bounds.
func (m *miniRead) readLen(len int) []byte {
	contents := m.contents[m.position : m.position+len]
	m.position += len
	return contents
}
