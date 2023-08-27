package bufferpool

import (
	"bytes"
	"io"
	"math"
	"math/rand"
	"sync"
	"testing"
)

func newRandomBytes(size int) []byte {
	candidates := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[i] = candidates[rand.Intn(len(candidates))]
	}
	return b
}

func TestBufferReadWriteSeek(t *testing.T) {
	bufSize := 10000
	data := newRandomBytes(bufSize)

	buf := NewBuffer(0)
	defer buf.Close()
	buf.Write(data)
	if buf.Len() != len(data) {
		t.Fatalf("got %d, want %d", buf.Len(), len(data))
	}

	b, err := io.ReadAll(buf)
	if err != nil {
		t.Fatalf("error occurs, err: %s", err.Error())
	}
	if len(b) != 0 {
		t.Fatalf("got %d, want %d", len(b), 0)
	}

	buf.Seek(0, io.SeekStart)
	b, _ = io.ReadAll(buf)
	if !bytes.Equal(data, b) {
		t.Fatalf("data inconsistent")
	}

	buf.Seek(0, io.SeekStart)
	b, _ = io.ReadAll(buf)
	if !bytes.Equal(data, b) {
		t.Fatalf("data inconsistent")
	}

	buf.Seek(0, io.SeekEnd)
	buf.Write(data)
	buf.Seek(0, io.SeekStart)
	b, _ = io.ReadAll(buf)
	if !bytes.Equal(data, b[:len(data)]) || !bytes.Equal(data, b[len(data):]) {
		t.Fatalf("data inconsistent")
	}

	buf.Seek(0, io.SeekEnd)
	buf.Seek(-int64(buf.Len()), io.SeekCurrent)
	off, _ := buf.Seek(int64(buf.Len()), io.SeekCurrent)
	if off != int64(buf.Len()) {
		t.Fatalf("got %d, want %d", off, buf.Len())
	}
}

func TestBufferGrow(t *testing.T) {
	buf := NewBuffer(1000)
	sum := buf.Len()
	for i := 0; i < 1; i++ {
		buf.Grow(i)
		sum += i
	}
	if buf.Len() != sum {
		t.Fatalf("got %d, want %d", buf.Len(), sum)
	}
}

func TestReadAtSequential(t *testing.T) {
	bufSize, chunkSize := 1000, 100
	buf := NewBuffer(0)
	if _, err := buf.Write(newRandomBytes(bufSize)); err != nil {
		t.Fatalf("error occurs, err: %s", err.Error())
	}

	buf.Seek(0, io.SeekStart)
	expected, err := io.ReadAll(buf)
	if err != nil {
		t.Fatalf("error occurs, err: %s", err.Error())
	}

	actual := make([]byte, bufSize)
	chunks := int(math.Ceil(float64(bufSize) / float64(chunkSize)))
	for i := 0; i < chunks; i++ {
		offset := i * chunkSize
		buf.ReadAt(actual[offset:offset+chunkSize], int64(offset))
	}
	if !bytes.Equal(expected, actual) {
		t.Fatalf("data inconsistent")
	}
}

func TestReadAtParallel(t *testing.T) {
	bufSize, chunkSize := 1000, 100
	buf := NewBuffer(0)
	if _, err := buf.Write(newRandomBytes(bufSize)); err != nil {
		t.Fatalf("error occurs, err: %s", err.Error())
	}

	buf.Seek(0, io.SeekStart)
	expected, err := io.ReadAll(buf)
	if err != nil {
		t.Fatalf("error occurs, err: %s", err.Error())
	}

	var wg sync.WaitGroup
	actual := make([]byte, bufSize)
	chunks := int(math.Ceil(float64(bufSize) / float64(chunkSize)))
	for i := 0; i < chunks; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			offset := i * chunkSize
			buf.ReadAt(actual[offset:offset+chunkSize], int64(offset))
		}(i)
	}
	wg.Wait()
	if !bytes.Equal(expected, actual) {
		t.Fatalf("data inconsistent")
	}
}

func BenchmarkBuiltinBuffer(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		i := 1
		for pb.Next() {
			_ = make([]byte, i)
			i = i << 1
			if i > 1<<25 {
				i = 1
			}
		}
	})
}

func BenchmarkBuffer(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		i := 1
		for pb.Next() {
			buf := NewBuffer(i)
			i = i << 1
			if i > 1<<25 {
				i = 1
			}
			buf.Close()
		}
	})
}
