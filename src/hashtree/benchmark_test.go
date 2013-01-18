package hashtree

import (
	"crypto/sha256"
	"hash"
	"testing"
)

var bench = NewTree()
var refBench = sha256.New()
var buf = make([]byte, 81920)

func benchmarkSize(b *testing.B, hash hash.Hash, size int) {
	b.SetBytes(int64(size))
	sum := make([]byte, hash.Size())
	for i := 0; i < b.N; i++ {
		hash.Reset()
		hash.Write(buf[:size])
		hash.Sum(sum[:0])
	}
}

func BenchmarkHash8Bytes(b *testing.B) {
	benchmarkSize(b, bench, 8)
}

func BenchmarkHash1K(b *testing.B) {
	benchmarkSize(b, bench, 1024)
}

func BenchmarkHash20K(b *testing.B) {
	benchmarkSize(b, bench, 20480)
}

func BenchmarkRef8Bytes(b *testing.B) {
	benchmarkSize(b, refBench, 8)
}

func BenchmarkRef1K(b *testing.B) {
	benchmarkSize(b, refBench, 1024)
}

func BenchmarkRef20K(b *testing.B) {
	benchmarkSize(b, refBench, 20480)
}
