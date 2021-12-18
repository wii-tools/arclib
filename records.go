package arclib

import (
	"os"
)

// ARCFile represents a file available within an ARC.
type ARCFile struct {
	// Filename is the name of this file. It must not be empty.
	Filename string
	// Length is the size (in bytes) of this file.
	Length int
	// Data is the contents of this file.
	Data []byte
}

// ARCDir represents a directory available within an ARC.
type ARCDir struct {
	// Filename is the name of this directory. It must not be empty.
	Filename string
	// Files is an array of files within this directory.
	Files []ARCFile
	// Subdirs is an array of directories within this directory.
	Subdirs []ARCDir

	// childCount is utilized during deserialization to track addition of children.
	childCount uint32
}

// AddDir adds a directory to the list of subdirectories.
func (d *ARCDir) AddDir(dir ARCDir) {
	d.Subdirs = append(d.Subdirs, dir)
}

// AddFile adds a file to the list of files in this directory.
func (d *ARCDir) AddFile(file ARCFile) {
	d.Files = append(d.Files, file)
}

// WriteFile adds a file with the specified contents to the directory.
func (d *ARCDir) WriteFile(name string, contents []byte) {
	// Determine whether this file already exists.
	existingFile, err := d.GetFile(name)
	if err == os.ErrNotExist {
		// Add a new file by the given name.
		file := ARCFile{
			Filename: name,
			Length:   len(contents),
			Data:     contents,
		}

		d.AddFile(file)
	} else {
		// Overwrite its existing data.
		existingFile.Data = contents
		existingFile.Length = len(contents)
	}
}

// GetFile retrieves the file by the given name.
func (d *ARCDir) GetFile(name string) (*ARCFile, error) {
	if name == "" {
		return nil, os.ErrInvalid
	}

	for _, file := range d.Files {
		if file.Filename == name {
			return &file, nil
		}
	}

	return nil, os.ErrNotExist
}

// GetDir retrieves the directory by the given name.
func (d *ARCDir) GetDir(name string) (*ARCDir, error) {
	if name == "" {
		return nil, os.ErrInvalid
	}

	for _, dir := range d.Subdirs {
		if dir.Filename == name {
			return &dir, nil
		}
	}

	return nil, os.ErrNotExist
}

// RecursiveCount returns the amount of files and sub-directories within.
func (d *ARCDir) RecursiveCount() int {
	// We start with this current record.
	count := 1
	// Add file count.
	count += len(d.Files)

	// Recurse through subdirectories for their sum.
	for _, subDir := range d.Subdirs {
		count += subDir.RecursiveCount()
	}

	return count
}

// Size returns the count of files and directories immediately within.
func (d *ARCDir) Size() int {
	return len(d.Files) + len(d.Subdirs)
}
