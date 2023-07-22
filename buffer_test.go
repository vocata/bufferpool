package bufferpool

import (
	"bytes"
	"io"
	"testing"
)

func TestBufferReadWriteSeek(t *testing.T) {
	data := make([]byte, 10000)
	for i := range data {
		data[i] = byte(i)
	}

	buf := NewBuffer(len(data))
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

func BenchmarkBuiltinBuffer(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		i := 1
		for pb.Next() {
			_ = make([]byte, i)
			i = i << 1
			if i > 1<<30 {
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
			if i > 1<<30 {
				i = 1
			}
			buf.Close()
		}
	})
}
