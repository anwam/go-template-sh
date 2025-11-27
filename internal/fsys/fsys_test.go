package fsys

import (
	"io/fs"
	"testing"
)

func TestMemoryFileSystem_WriteAndRead(t *testing.T) {
	mfs := NewMemory()

	content := []byte("hello world")
	err := mfs.WriteFile("/test/file.txt", content, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := mfs.ReadFile("/test/file.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("expected %q, got %q", content, data)
	}
}

func TestMemoryFileSystem_ReadNonExistent(t *testing.T) {
	mfs := NewMemory()

	_, err := mfs.ReadFile("/nonexistent")
	if err != fs.ErrNotExist {
		t.Errorf("expected ErrNotExist, got %v", err)
	}
}

func TestMemoryFileSystem_MkdirAll(t *testing.T) {
	mfs := NewMemory()

	err := mfs.MkdirAll("/a/b/c", 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Verify directories were created (including root /)
	dirs := mfs.Dirs()
	required := []string{"/a/b/c", "/a/b", "/a"}

	for _, req := range required {
		found := false
		for _, dir := range dirs {
			if dir == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing directory: %s", req)
		}
	}
}

func TestMemoryFileSystem_Stat(t *testing.T) {
	mfs := NewMemory()

	// Write a file
	err := mfs.WriteFile("/test/file.txt", []byte("content"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Stat the file
	info, err := mfs.Stat("/test/file.txt")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if info.Name() != "file.txt" {
		t.Errorf("expected name file.txt, got %s", info.Name())
	}
	if info.IsDir() {
		t.Error("expected file, got directory")
	}
	if info.Size() != 7 {
		t.Errorf("expected size 7, got %d", info.Size())
	}

	// Stat a directory
	info, err = mfs.Stat("/test")
	if err != nil {
		t.Fatalf("Stat directory failed: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected directory, got file")
	}
}

func TestMemoryFileSystem_Remove(t *testing.T) {
	mfs := NewMemory()

	// Write a file
	err := mfs.WriteFile("/test/file.txt", []byte("content"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Remove the file
	err = mfs.Remove("/test/file.txt")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify it's gone
	if mfs.HasFile("/test/file.txt") {
		t.Error("file should have been removed")
	}
}

func TestMemoryFileSystem_RemoveAll(t *testing.T) {
	mfs := NewMemory()

	// Write files in a directory tree
	_ = mfs.WriteFile("/test/a/file1.txt", []byte("1"), 0644)
	_ = mfs.WriteFile("/test/a/file2.txt", []byte("2"), 0644)
	_ = mfs.WriteFile("/test/b/file3.txt", []byte("3"), 0644)
	_ = mfs.WriteFile("/other/file.txt", []byte("other"), 0644)

	// Remove /test
	err := mfs.RemoveAll("/test")
	if err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	// Verify /test files are gone
	if mfs.HasFile("/test/a/file1.txt") {
		t.Error("file1.txt should have been removed")
	}
	if mfs.HasFile("/test/a/file2.txt") {
		t.Error("file2.txt should have been removed")
	}
	if mfs.HasFile("/test/b/file3.txt") {
		t.Error("file3.txt should have been removed")
	}

	// Verify /other is still there
	if !mfs.HasFile("/other/file.txt") {
		t.Error("other/file.txt should still exist")
	}
}

func TestMemoryFileSystem_HasFile(t *testing.T) {
	mfs := NewMemory()

	if mfs.HasFile("/test/file.txt") {
		t.Error("file should not exist initially")
	}

	_ = mfs.WriteFile("/test/file.txt", []byte("content"), 0644)

	if !mfs.HasFile("/test/file.txt") {
		t.Error("file should exist after write")
	}
}

func TestMemoryFileSystem_FileContent(t *testing.T) {
	mfs := NewMemory()

	_ = mfs.WriteFile("/test/file.txt", []byte("hello world"), 0644)

	content := mfs.FileContent("/test/file.txt")
	if content != "hello world" {
		t.Errorf("expected 'hello world', got %q", content)
	}

	// Non-existent file returns empty string
	content = mfs.FileContent("/nonexistent")
	if content != "" {
		t.Errorf("expected empty string for nonexistent file, got %q", content)
	}
}

func TestOSFileSystem_Interface(t *testing.T) {
	// Verify OSFileSystem implements FileSystem
	var _ FileSystem = (*OSFileSystem)(nil)
	var _ FileSystem = (*MemoryFileSystem)(nil)
}
