package locate

import (
	"math/rand"
	"os"
	"runtime"
	"testing"
	"time"
)

type fakeFileInfo struct {
	name  string
	isdir bool
}

func (s fakeFileInfo) IsDir() bool        { return s.isdir }
func (s fakeFileInfo) Name() string       { return s.name }
func (s fakeFileInfo) ModTime() time.Time { panic("not implemented") }
func (s fakeFileInfo) Mode() os.FileMode  { panic("not implemented") }
func (s fakeFileInfo) Size() int64        { panic("not implemented") }
func (s fakeFileInfo) Sys() interface{}   { panic("not implemented") }

type fakeWalker []struct {
	path string
	info fakeFileInfo
}

func randString(n int) string {
	length := 10 + rand.Intn(n-10)
	ret := make([]byte, length)

	rng := int(byte('Z') - byte('A'))
	for i := 0; i < length; i++ {
		ret[i] = byte(rand.Intn(rng)) + byte('A')
	}

	return string(ret)
}

func newFakeWalker(size int) fakeWalker {
	elems := make(fakeWalker, size)

	for i := range elems {
		rs := randString(20)
		elems[i].path = "/" + rs
		elems[i].info = fakeFileInfo{rs, false}
	}

	return elems
}

func (fw fakeWalker) Walk(f func(string, os.FileInfo, error) error) error {
	for _, r := range fw {
		if err := f(r.path, r.info, nil); err != nil {
			return err
		}
	}

	return nil
}

var benchIdx []Record

type logger interface {
	Logf(string, ...interface{})
	Log(...interface{})
}

const (
	benchIdxPath = "./index.db"
	benchRoot    = "/"
	benchPattern = "cpp"
)

func TestSearch(t *testing.T) {
	n := runtime.NumCPU()
	oldn := runtime.GOMAXPROCS(n)
	defer runtime.GOMAXPROCS(oldn)

	const walkerSize = 100000
	w := newFakeWalker(walkerSize)
	idx := newIndex(n, 10*time.Minute, w)

	pattern := w[rand.Intn(walkerSize)].info.name

	results := idx.Search(pattern)
	if len(results) < 1 {
		t.Fatal("Search didn't find any result")
	} else if results[0].Name != pattern {
		t.Fatal("Search returned non-matching result")
	}
}

func BenchmarkSearch(b *testing.B) {
	n := runtime.NumCPU()
	oldn := runtime.GOMAXPROCS(n)
	defer runtime.GOMAXPROCS(oldn)

	const walkerSize = 100000
	w := newFakeWalker(walkerSize)
	idx := newIndex(n, 10*time.Minute, w)

	pattern := w[rand.Intn(walkerSize)].info.name

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.Search(pattern)
	}
}
