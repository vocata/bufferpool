package bufferpool

import (
	"bytes"
	"runtime"
	"runtime/debug"
	"testing"
)

func TestAllocations(t *testing.T) {
	pool := NewBufferPool[byte]()
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	for i := 0; i < 10000; i++ {
		b := pool.Get(1010)
		pool.Put(b)
	}
	runtime.GC()
	runtime.ReadMemStats(&m2)
	frees := m2.Frees - m1.Frees
	if frees > 100 {
		t.Fatalf("expected less than 100 frees after GC, got %d", frees)
	}
}

func TestPool(t *testing.T) {
	// disable GC so we can control when it happens.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))
	var p BufferPool[byte]

	a := make([]byte, 21)
	a[0] = 1
	b := make([]byte, 2050)
	b[0] = 2
	p.Put(a)
	p.Put(b)
	if g := p.Get(16); &g[0] != &a[0] {
		t.Fatalf("got [%d,...]; want [1,...]", g[0])
	}
	if g := p.Get(2048); &g[0] != &b[0] {
		t.Fatalf("got [%d,...]; want [2,...]", g[0])
	}
	if g := p.Get(16); cap(g) != 16 || !bytes.Equal(g[:16], make([]byte, 16)) {
		t.Fatalf("got existing slice; want new slice")
	}
	if g := p.Get(2048); cap(g) != 2048 || !bytes.Equal(g[:2048], make([]byte, 2048)) {
		t.Fatalf("got existing slice; want new slice")
	}
	if g := p.Get(1); cap(g) != 1 || !bytes.Equal(g[:1], make([]byte, 1)) {
		t.Fatalf("got existing slice; want new slice")
	}
	d := make([]byte, 1023)
	d[0] = 3
	p.Put(d)
	if g := p.Get(1024); cap(g) != 1024 || !bytes.Equal(g, make([]byte, 1024)) {
		t.Fatalf("got existing slice; want new slice")
	}
	if g := p.Get(512); cap(g) != 1023 || g[0] != 3 {
		t.Fatalf("got [%d,...]; want [3,...]", g[0])
	}
	p.Put(a)

	debug.SetGCPercent(100) // to allow following GC to actually run
	runtime.GC()
	// For some reason, you need to run GC twice on go 1.16 if you want it to reliably work.
	runtime.GC()
	if g := p.Get(10); &g[0] == &a[0] {
		t.Fatalf("got a; want new slice after GC")
	}
}

func BenchmarkPool(b *testing.B) {
	var p BufferPool[int]
	b.RunParallel(func(pb *testing.PB) {
		i := 7
		for pb.Next() {
			b := p.Get(int64(i))
			b[0] = i
			p.Put(b)

			i = i << 1
			if i > 1<<20 {
				i = 7
			}
		}
	})
}

func BenchmarkAlloc(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		i := 7
		for pb.Next() {
			b := make([]int, i)
			b[1] = i

			i = i << 1
			if i > 1<<20 {
				i = 7
			}
		}
	})
}
