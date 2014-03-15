package locate

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func init() {
	gob.Register(dmnResult(""))
}

type dmnResult string

func (r dmnResult) Name() string { return string(r) }

type dmnQueryMsg struct {
	query string
	reply chan []Result
}

type daemon struct {
	queries chan dmnQueryMsg
	root    string
	index   map[string]*FileInfo
	idxMut  sync.RWMutex
}

type FileInfo struct {
	osInfo os.FileInfo
	sync.Mutex
}

func newFileInfo(info os.FileInfo) *FileInfo {
	return &FileInfo{
		osInfo: info,
	}
}

func (info *FileInfo) Update(osInfo os.FileInfo) {
	info.Lock()
	defer info.Unlock()

	info.osInfo = osInfo
}

func (info *FileInfo) Info() os.FileInfo {
	info.Lock()
	defer info.Unlock()

	return info.osInfo
}

func NewDaemon(root string) (Locator, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	dmn := &daemon{
		queries: make(chan dmnQueryMsg),
		root:    absRoot,
		index:   make(map[string]*FileInfo),
	}

	go dmn.director()

	return dmn, nil
}

func (d *daemon) reIndex() {
	const nThreads = 20

	updatePaths := make(chan string)

	updater := func() {
		for fpath := range updatePaths {
			info, err := os.Stat(fpath)
			if err != nil {
				fmt.Println("stat error:", err)
				continue
			}
			if info == nil {
				fmt.Println("stat returned nil info:", err)
				continue
			}

			d.idxMut.RLock()
			if _, ok := d.index[fpath]; ok {
				d.index[fpath].Update(info)
				d.idxMut.RUnlock()
			} else {
				d.idxMut.RUnlock()

				d.idxMut.Lock()
				d.index[fpath] = newFileInfo(info)
				d.idxMut.Unlock()
			}
		}
	}

	for i := 0; i < nThreads; i++ {
		go updater()
	}

	fmt.Println("root:", d.root)
	err := filepath.Walk(d.root, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		d.idxMut.RLock()
		entry, ok := d.index[fpath]
		d.idxMut.RUnlock()
		if ok {
			if info.ModTime().After(entry.Info().ModTime()) {
				updatePaths <- fpath
			}
		} else {
			updatePaths <- fpath
		}

		return nil
	})
	if err != nil {
		fmt.Println("reIndex error:", err)
		return
	}

	close(updatePaths)
}

func (d *daemon) handleQuery(msg dmnQueryMsg) {
	ret := []Result{}

	d.idxMut.RLock()
	for fpath, info := range d.index {
		if !strings.Contains(filepath.Base(fpath), msg.query) {
			continue
		}

		idxInfo := info.Info()

		ret = append(ret, dmnResult(idxInfo.Name()))
	}
	d.idxMut.RUnlock()

	msg.reply <- ret
}

func (d *daemon) director() {
	go func() {
		for {
			start := time.Now()
			shim := time.After(time.Minute * 2)
			d.reIndex()
			fmt.Println("index complete in", time.Since(start))
			<-shim
		}
	}()

	for {
		go d.handleQuery(<-d.queries)
	}
}

func (d *daemon) Locate(query string) []Result {
	reply := make(chan []Result)
	d.queries <- dmnQueryMsg{query, reply}
	return <-reply
}
