package bufferpool

import (
	"math/bits"
	"sync"
)

const MaxMemorySize int64 = 1<<32 - 1 // 4GB

type BufferPool[T any] struct {
	pools [32]sync.Pool
	ptr   sync.Pool
}

type bufPtr[T any] struct {
	buf []T
}

func NewBufferPool[T any]() *BufferPool[T] {
	return &BufferPool[T]{}
}

func (p *BufferPool[T]) Get(size int64) []T {
	if size == 0 {
		return nil
	}
	if size < 0 {
		panic("negative allocated size")
	}
	if size > MaxMemorySize {
		panic("reach to maximum allocated memory")
	}
	idx := calcNextBin(uint64(size))
	if ptr := p.pools[idx].Get(); ptr != nil {
		bp := ptr.(*bufPtr[T])
		buf := bp.buf[:size]
		bp.buf = nil
		p.ptr.Put(ptr)
		return buf
	}

	return make([]T, size, 1<<idx)
}

func (p *BufferPool[T]) Put(buf []T) {
	capacity := cap(buf)
	if capacity == 0 || int64(capacity) > MaxMemorySize {
		return
	}

	idx := calcPrevBin(uint64(capacity))
	var bp *bufPtr[T]
	if ptr := p.ptr.Get(); ptr != nil {
		bp = ptr.(*bufPtr[T])
	} else {
		bp = new(bufPtr[T])
	}
	bp.buf = buf
	p.pools[idx].Put(bp)
}

func calcNextBin(size uint64) int {
	return bits.Len64(size - 1)
}

func calcPrevBin(size uint64) int {
	next := calcNextBin(size)
	if size == (1 << next) {
		return next
	}

	return next - 1
}
