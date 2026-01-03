package main

import (
	"io"
	"os"
	"sync/atomic"
)

// threadReader provides concurrent, bounded reading from a file.
// It is similar to io.SectionReader but additionally updates an atomic
// position counter, enabling external monitoring of read progress.
//
// Each threadReader instance reads from a specific byte range [pos, end)
// of the underlying file, using ReadAt for thread-safe concurrent access.
type threadReader struct {
	atomicPos *uint64
	pos       int64
	end       int64
	f         *os.File
}

// Read implements io.Reader. It reads up to len(b) bytes from the file
// starting at the current position, respecting the end boundary. After each
// read, the atomic position counter is updated to reflect progress.
// Returns io.EOF when the end boundary is reached.
func (t *threadReader) Read(b []byte) (int, error) {
	if int64(len(b))+t.pos > t.end {
		// reduce b
		b = b[:t.end-t.pos]
		if len(b) == 0 {
			return 0, io.EOF
		}
	}
	n, err := t.f.ReadAt(b, t.pos)
	if err != nil {
		return 0, err
	}
	t.pos += int64(n)
	atomic.StoreUint64(t.atomicPos, uint64(t.pos))
	return n, nil
}
