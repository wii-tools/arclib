package arclib

import (
	"errors"
	"strings"
)

const (
	ARCHeader = 0x55AA382D
)

var (
	ErrInvalidRootNode = errors.New("root node was not a directory")
	ErrInvalidMagic    = errors.New("invalid ARC magic")
	ErrUnknownNode     = errors.New("unknown node type")
)

// ARC describes a hierarchy suitable for serialization and deserialization of an ARC file.
type ARC struct {
	// RootRecord holds the root record for this ARC.
	//
	// It's important to note that this root record is nameless.
	// Many officially-provided ARCs within Nintendo games have one folder
	// containing all data, typically named "arc".
	// This folder has "arc" as one of its contents.
	RootRecord ARCDir
}

// OpenDir returns the directory at the given path, or an errir if not possible.
func (a *ARC) OpenDir(path string) (*ARCDir, error) {
	components := strings.Split(path, "/")

	// The root node is where we start our loop.
	current := &a.RootRecord
	for _, dirName := range components {
		testDir, err := current.GetDir(dirName)
		if err != nil {
			return nil, err
		}
		current = testDir
	}

	return current, nil
}

// OpenFile returns a file at the given path, or an error if not possible.
func (a *ARC) OpenFile(path string) (*ARCFile, error) {
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

	// Obtain directories leading up if needed
	if len(dirs) != 0 {
		dir, err := a.OpenDir(strings.Join(dirs, "/"))
		if err != nil {
			return nil, err
		}

		// Retrieve our file from the last directory in our hierarchy.
		return dir.GetFile(filename)
	} else {
		// Retrieve our file from the root record.
		return a.RootRecord.GetFile(filename)
	}
}

// ReadFile returns the contents of a file at the given path.
func (a *ARC) ReadFile(path string) ([]byte, error) {
	file, err := a.OpenFile(path)
	if err != nil {
		return nil, err
	}

	return file.Data, nil
}

// WriteFile writes the passed contents at the given path.
func (a *ARC) WriteFile(path string, contents []byte) error {
	file, err := a.OpenFile(path)
	if err != nil {
		return err
	}

	file.Write(contents)
	return nil
}

// FileSize returns the size in bytes of the file for the given path.
func (a *ARC) FileSize(path string) (int, error) {
	file, err := a.OpenFile(path)
	if err != nil {
		return 0, err
	}

	return file.Size(), nil
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
