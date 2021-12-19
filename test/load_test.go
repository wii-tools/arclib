package test

import (
	"bytes"
	"github.com/wii-tools/arclib"
	"testing"
)

// TestEmptyARC ensures that an empty ARC has no other contents added.
func TestEmptyARC(t *testing.T) {
	arc, err := arclib.LoadFromFile("./empty.arc")
	if err != nil {
		t.Error(err)
	}

	// An empty ARC should always have one directory - the root.
	if arc.RootRecord.RecursiveCount() != 1 {
		t.Error("invalid child count for empty ARC")
	}

	// It should have no children.
	if arc.RootRecord.Size() != 0 {
		t.Error("invalid directory size for empty ARC")
	}
}

// TestRootFile ensures the root file within hierarchy.arc is readable.
func TestRootFile(t *testing.T) {
	arc, err := arclib.LoadFromFile("./hierarchy.arc")
	if err != nil {
		t.Error(err)
	}

	rootContents, err := arc.Read("root_file")
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(rootContents, []byte("root file contents")) {
		t.Error("root contents did not match what was expected")
	}
}

// TestSubDir ensures that a file is readable within the sub directory.
func TestSubDir(t *testing.T) {
	arc, err := arclib.LoadFromFile("./hierarchy.arc")
	if err != nil {
		t.Error(err)
	}

	rootContents, err := arc.Read("subdir/sub_file")
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(rootContents, []byte("sub file contents")) {
		t.Error("sub contents did not match what was expected")
	}
}