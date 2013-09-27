package locate

import (
	"runtime"
	"testing"
)

var benchIdx []Record

func initBench() {
	if benchIdx == nil {
		benchIdx = CreateIndex(benchRoot)
	}
}

const (
	benchIdxPath = "./index.db"
	benchRoot    = "/"
	benchPattern = "cpp"
)

func bBenchmarkWriteSingle(b *testing.B) {
	if err := WriteIndex(benchIdxPath, CreateIndex(benchRoot)); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if index, err := ReadIndex(benchIdxPath); err != nil {
			b.Fatal(err)
		} else {
			SearchIndex(index, benchPattern)
		}
	}
}

func bBenchmarkSearchSplit(b *testing.B) {
	n := runtime.NumCPU()
	oldn := runtime.GOMAXPROCS(n)
	defer runtime.GOMAXPROCS(oldn)
	if err := WriteSplitIndex(benchIdxPath, CreateIndex(benchRoot), n); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if _, err := SearchSplitIndex(benchIdxPath, benchPattern, n); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInMemSearch(b *testing.B) {
	initBench()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SearchIndex(benchIdx, "15-411")
	}
}

func BenchmarkInMemHashSearch(b *testing.B) {
	idx := NewIndex()

	idx.Create(benchRoot)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.Search("15-411")
	}
}
