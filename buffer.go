package bufferpool

import (
	"errors"
	"io"
	"sync"
)

var (
	globalPool     *BufferPool[byte]
	globalPoolOnce sync.Once
)

// Buffer is not concurrency-safe
type Buffer struct {
	buf       []byte
	off       int64
	bootstrap [64]byte

	pool *BufferPool[byte]
}

func NewBuffer(size int) *Buffer {
	return NewBufferWithPool(size, nil)
}

func NewBufferWithPool(size int, pool *BufferPool[byte]) *Buffer {
	if pool == nil {
		globalPoolOnce.Do(func() {
			globalPool = new(BufferPool[byte])
		})
		pool = globalPool
	}
	b := &Buffer{pool: pool}

	b.buf = b.getBuf(int64(size))

	return b
}

func (b *Buffer) Len() int {
	return len(b.buf)
}

func (b *Buffer) Read(dst []byte) (int, error) {
	if b.off >= int64(len(b.buf)) {
		return 0, io.EOF
	}
	n := copy(dst, b.buf[b.off:])
	b.off += int64(n)
	return n, nil
}

func (b *Buffer) Write(src []byte) (int, error) {
	if need := b.off + int64(len(src)) - int64(len(b.buf)); need > 0 {
		b.grow(need)
	}
	n := copy(b.buf[b.off:], src)
	b.off += int64(n)
	return n, nil
}

func (b *Buffer) Grow(n int) {
	b.grow(int64(n))
}

func (b *Buffer) grow(n int64) {
	bLen := int64(len(b.buf))
	bCap := int64(cap(b.buf))

	if bCap >= bLen+n {
		b.buf = b.buf[:bLen+n]
		return
	}

	newBuf := b.getBuf(bLen + n)
	copy(newBuf, b.buf)
	b.putBuf()
	b.buf = newBuf
}

func (b *Buffer) getBuf(n int64) []byte {
	if n <= int64(len(b.bootstrap)) {
		return b.bootstrap[:n]
	}
	return b.pool.Get(n)
}

func (b *Buffer) putBuf() {
	if cap(b.buf) > len(b.bootstrap) {
		b.pool.Put(b.buf)
	}
	b.buf = nil
}

func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = b.off + offset
	case io.SeekEnd:
		abs = int64(len(b.buf)) + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("negative position")
	}
	b.off = abs

	return abs, nil
}

func (b *Buffer) Close() error {
	b.putBuf()
	return nil
}
