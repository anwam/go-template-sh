package fsys

import (
"io/fs"
"path/filepath"
"strings"
"sync"
"time"
)

// MemoryFileSystem implements FileSystem using in-memory storage.
// Useful for testing without touching the real file system.
type MemoryFileSystem struct {
	mu    sync.RWMutex
	files map[string][]byte
	dirs  map[string]bool
}

// NewMemory creates a new in-memory file system.
func NewMemory() *MemoryFileSystem {
	return &MemoryFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

func (m *MemoryFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path = filepath.Clean(path)
	m.dirs[path] = true

	// Create parent directories
	for {
		parent := filepath.Dir(path)
		if parent == path || parent == "." {
			break
		}
		m.dirs[parent] = true
		path = parent
	}
	return nil
}

func (m *MemoryFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name = filepath.Clean(name)

	// Create parent directories
	dir := filepath.Dir(name)
	m.dirs[dir] = true
	for {
		parent := filepath.Dir(dir)
		if parent == dir || parent == "." {
			break
		}
		m.dirs[parent] = true
		dir = parent
	}

	// Store file content
	m.files[name] = make([]byte, len(data))
	copy(m.files[name], data)
	return nil
}

func (m *MemoryFileSystem) ReadFile(name string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name = filepath.Clean(name)
	data, ok := m.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}

	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

func (m *MemoryFileSystem) Stat(name string) (fs.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name = filepath.Clean(name)

	if data, ok := m.files[name]; ok {
		return &memoryFileInfo{
			name: filepath.Base(name),
			size: int64(len(data)),
			mode: 0644,
		}, nil
	}

	if _, ok := m.dirs[name]; ok {
		return &memoryFileInfo{
			name:  filepath.Base(name),
			size:  0,
			mode:  fs.ModeDir | 0755,
			isDir: true,
		}, nil
	}

	return nil, fs.ErrNotExist
}

func (m *MemoryFileSystem) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name = filepath.Clean(name)
	if _, ok := m.files[name]; ok {
		delete(m.files, name)
		return nil
	}
	if _, ok := m.dirs[name]; ok {
		delete(m.dirs, name)
		return nil
	}
	return fs.ErrNotExist
}

func (m *MemoryFileSystem) RemoveAll(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path = filepath.Clean(path)

	// Remove all files under path
	for name := range m.files {
		if strings.HasPrefix(name, path) {
			delete(m.files, name)
		}
	}

	// Remove all dirs under path
	for name := range m.dirs {
		if strings.HasPrefix(name, path) {
			delete(m.dirs, name)
		}
	}

	return nil
}

// Files returns a copy of all files in the memory file system.
// Useful for testing assertions.
func (m *MemoryFileSystem) Files() map[string][]byte {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]byte)
	for k, v := range m.files {
		result[k] = v
	}
	return result
}

// Dirs returns a copy of all directories in the memory file system.
func (m *MemoryFileSystem) Dirs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]string, 0, len(m.dirs))
	for k := range m.dirs {
		result = append(result, k)
	}
	return result
}

// HasFile checks if a file exists at the given path.
func (m *MemoryFileSystem) HasFile(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.files[filepath.Clean(path)]
	return ok
}

// FileContent returns the content of a file as a string.
func (m *MemoryFileSystem) FileContent(path string) string {
	data, err := m.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

type memoryFileInfo struct {
	name  string
	size  int64
	mode  fs.FileMode
	isDir bool
}

func (fi *memoryFileInfo) Name() string       { return fi.name }
func (fi *memoryFileInfo) Size() int64        { return fi.size }
func (fi *memoryFileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *memoryFileInfo) ModTime() time.Time { return time.Now() }
func (fi *memoryFileInfo) IsDir() bool        { return fi.isDir }
func (fi *memoryFileInfo) Sys() any           { return nil }
