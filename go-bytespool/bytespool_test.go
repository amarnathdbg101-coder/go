package bytespool

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func TestBuffer_Writers(t *testing.T) {
	buf := Get(100)
	defer Put(buf)

	buf.WriteString("hello ")
	buf.Write([]byte("world! "))
	buf.WriteByte('#')
	buf.WriteInt(42)
	buf.WriteString(" pi=")
	buf.WriteFloat(3.14159, 2)

	expected := "hello world! #42 pi=3.14"
	if buf.String() != expected {
		t.Fatalf("expected '%s', got '%s'", expected, buf.String())
	}

	if buf.Len() != len(expected) {
		t.Fatalf("expected len %d, got %d", len(expected), buf.Len())
	}
}

func TestBuffer_ReadFrom(t *testing.T) {
	buf := Get(TierSmall)
	defer Put(buf)

	sample := strings.Repeat("A", 1000)
	reader := strings.NewReader(sample)

	n, err := buf.ReadFrom(reader)
	if err != nil {
		t.Fatalf("ReadFrom failed: %v", err)
	}
	if n != 1000 {
		t.Fatalf("expected 1000 bytes read, got %d", n)
	}
	if buf.String() != sample {
		t.Fatalf("buffer content mismatch")
	}
}

func TestTieredPool_Tiers(t *testing.T) {
	bSmall := Get(256)
	if cap(bSmall.buf) < TierSmall {
		t.Fatalf("expected cap >= %d, got %d", TierSmall, cap(bSmall.buf))
	}
	Put(bSmall)

	bMed := Get(2048)
	if cap(bMed.buf) < TierMedium {
		t.Fatalf("expected cap >= %d, got %d", TierMedium, cap(bMed.buf))
	}
	Put(bMed)

	bLarge := Get(32 * 1024)
	if cap(bLarge.buf) < TierLarge {
		t.Fatalf("expected cap >= %d, got %d", TierLarge, cap(bLarge.buf))
	}
	Put(bLarge)

	bHuge := Get(200 * 1024)
	if cap(bHuge.buf) < TierHuge {
		t.Fatalf("expected cap >= %d, got %d", TierHuge, cap(bHuge.buf))
	}
	Put(bHuge)
}

func TestTieredPool_BloatDiscard(t *testing.T) {
	pool := New(1024) // max cap 1KB
	buf := pool.Get(256)

	// Expand beyond maxCap
	buf.Write(bytes.Repeat([]byte("X"), 2048))
	if cap(buf.buf) <= 1024 {
		t.Fatalf("expected cap to exceed 1024")
	}

	// Put should discard without panic
	pool.Put(buf)
}

func TestWithBuffer(t *testing.T) {
	var result string
	WithBuffer(TierSmall, func(buf *Buffer) {
		buf.WriteString("scoped execution")
		result = buf.String()
	})

	if result != "scoped execution" {
		t.Fatalf("unexpected result: %s", result)
	}
}

func TestTieredPool_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			WithBuffer(500, func(buf *Buffer) {
				buf.WriteString("goroutine-")
				buf.WriteInt(int64(id))
				if !strings.HasPrefix(buf.String(), "goroutine-") {
					t.Errorf("corrupt buffer content")
				}
			})
		}(i)
	}
	wg.Wait()
}
