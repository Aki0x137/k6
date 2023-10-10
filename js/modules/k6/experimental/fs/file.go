package fs

import (
	"io"
	"path/filepath"
	"sync/atomic"
)

// file is an abstraction for interacting with files.
type file struct {
	path string

	// data holds a pointer to the file's data
	data []byte

	// offset holds the current offset in the file
	offset int

	// closed indicates whether the file has been closed
	closed atomic.Bool
}

// Stat returns a FileInfo describing the named file.
func (f *file) stat() (*FileInfo, error) {
	if f.closed.Load() {
		return nil, newFsError(BadResourceError, "cannot stat closed file")
	}

	filename := filepath.Base(f.path)
	return &FileInfo{Name: filename, Size: len(f.data)}, nil
}

// FileInfo holds information about a file.
type FileInfo struct {
	// Name holds the base name of the file.
	Name string `json:"name"`

	// Size holds the size of the file in bytes.
	Size int `json:"size"`
}

// Read reads up to len(into) bytes into the provided byte slice.
//
// It returns the number of bytes read (0 <= n <= len(into)) and any error
// encountered.
//
// If the end of the file has been reached, it returns EOFError.
func (f *file) Read(into []byte) (n int, err error) {
	if f.closed.Load() {
		return 0, newFsError(BadResourceError, "cannot read from closed file")
	}

	start := f.offset
	if start == len(f.data) {
		return 0, newFsError(EOFError, "EOF")
	}

	end := f.offset + len(into)
	if end > len(f.data) {
		end = len(f.data)
		// We align with the [io.Reader.Read] method's behavior
		// and return EOFError when we reach the end of the
		// file, regardless of how much data we were able to
		// read.
		err = newFsError(EOFError, "EOF")
	}

	n = copy(into, f.data[start:end])

	f.offset += n

	return n, err
}

// Ensure that `file` implements the io.Reader interface.
var _ io.Reader = (*file)(nil)

// Seek sets the offset for the next operation on the file, under the mode given by `whence`.
//
// `offset` indicates the number of bytes to move the offset. Based on
// the `whence` parameter, the offset is set relative to the start,
// current offset or end of the file.
//
// When using SeekModeStart, the offset must be positive.
// Negative offsets are allowed when using `SeekModeCurrent` or `SeekModeEnd`.
func (f *file) Seek(offset int, whence SeekMode) (int, error) {
	if f.closed.Load() {
		return 0, newFsError(BadResourceError, "cannot seek in closed file")
	}

	newOffset := f.offset

	switch whence {
	case SeekModeStart:
		if offset < 0 {
			return 0, newFsError(TypeError, "offset cannot be negative when using SeekModeStart")
		}

		newOffset = offset
	case SeekModeCurrent:
		newOffset += offset
	case SeekModeEnd:
		if offset > 0 {
			return 0, newFsError(TypeError, "offset cannot be positive when using SeekModeEnd")
		}

		newOffset = len(f.data) + offset
	default:
		return 0, newFsError(TypeError, "invalid seek mode")
	}

	if newOffset < 0 {
		return 0, newFsError(TypeError, "seeking before start of file")
	}

	if newOffset > len(f.data) {
		return 0, newFsError(TypeError, "seeking beyond end of file")
	}

	// Note that the implementation assumes one `file` instance per file/vu.
	// If that assumption was invalidated, we would need to atomically update
	// the offset instead.
	f.offset = newOffset

	return newOffset, nil
}

// SeekMode is used to specify the seek mode when seeking in a file.
type SeekMode = int

const (
	// SeekModeStart sets the offset relative to the start of the file.
	SeekModeStart SeekMode = iota + 1

	// SeekModeCurrent seeks relative to the current offset.
	SeekModeCurrent

	// SeekModeEnd seeks relative to the end of the file.
	//
	// When using this mode the seek operation will move backwards from
	// the end of the file.
	SeekModeEnd
)

// Close closes the file instance, and marks it as closed.
func (f *file) Close() error {
	if f.closed.Load() {
		return newFsError(BadResourceError, "file already closed")
	}

	f.closed.Store(true)

	return nil
}
