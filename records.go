package arclib

import (
	"log"
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

// GetFile retrieves the file by the given name.
func (d *ARCDir) GetFile(name string) (ARCFile, error) {
	for _, file := range d.Files {
		if file.Filename == name {
			return file, nil
		}
	}

	return ARCFile{}, os.ErrNotExist
}

// GetDir retrieves the directory by the given name.
func (d *ARCDir) GetDir(name string) (ARCDir, error) {
	log.Println("currently in", d.Filename)
	for _, dir := range d.Subdirs {
		log.Println("considering", dir.Filename, "for", name)
		if dir.Filename == name {
			return dir, nil
		}
	}

	return ARCDir{}, os.ErrNotExist
}
