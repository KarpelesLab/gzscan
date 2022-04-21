package main

import (
	"io"
	"os"
	"sync/atomic"
)

type threadReader struct {
	atomicPos *uint64
	pos       int64
	end       int64
	f         *os.File
}

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
