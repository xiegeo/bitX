package hashtree

import (
	"crypto/sha256"
	"hash"
	"testing"
)

var fileBench = NewFile()
var treeBench = NewTree()
var refBench = sha256.New()
var buf = make([]byte, 20480)

func benchmarkSize(b *testing.B, hash hash.Hash, size int) {
	b.SetBytes(int64(size))
	for i := 0; i < b.N; i++ {
		hash.Reset()
		hash.Write(buf[:size])
		hash.Sum(nil)
	}
}

func BenchmarkFile8Bytes(b *testing.B) {
	benchmarkSize(b, fileBench, 8)
}

func BenchmarkFile1K(b *testing.B) {
	benchmarkSize(b, fileBench, 1024)
}

func BenchmarkFile20K(b *testing.B) {
	benchmarkSize(b, fileBench, 20480)
}

func BenchmarkTree8Bytes(b *testing.B) {
	benchmarkSize(b, treeBench, 8)
}

func BenchmarkTree1K(b *testing.B) {
	benchmarkSize(b, treeBench, 1024)
}

func BenchmarkTree20K(b *testing.B) {
	benchmarkSize(b, treeBench, 20480)
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
