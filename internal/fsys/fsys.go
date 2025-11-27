package fsys

import (
"io/fs"
"os"
"path/filepath"
)

// FileSystem defines the interface for file system operations.
// This abstraction enables testing without actual file I/O.
type FileSystem interface {
	// MkdirAll creates a directory and all necessary parents.
	MkdirAll(path string, perm fs.FileMode) error

	// WriteFile writes data to a file, creating it if necessary.
	WriteFile(name string, data []byte, perm fs.FileMode) error

	// ReadFile reads the contents of a file.
	ReadFile(name string) ([]byte, error)

	// Stat returns file info for the named file.
	Stat(name string) (fs.FileInfo, error)

	// Remove removes the named file or empty directory.
	Remove(name string) error

	// RemoveAll removes path and any children it contains.
	RemoveAll(path string) error
}

// OSFileSystem implements FileSystem using the real operating system.
type OSFileSystem struct{}

// New returns a new OSFileSystem instance.
func New() *OSFileSystem {
	return &OSFileSystem{}
}

func (f *OSFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *OSFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return err
	}
	return os.WriteFile(name, data, perm)
}

func (f *OSFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (f *OSFileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (f *OSFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (f *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}
