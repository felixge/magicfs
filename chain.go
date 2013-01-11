package magicfs

import (
	"net/http"
	"os"
)

// NewChainFs returns a new ChainFs for the given chain.
func NewChainFs(chain... http.FileSystem) http.FileSystem {
	return &ChainFs{chain: chain}
}

// ChainFs implements a chained file system.
type ChainFs struct {
	chain []http.FileSystem
}

// BUG(felixge): Only ENOENT errors should advance the chain, other errors
// should be returned right away. Otherwise Open() and Readdir() can get out
// of sync.

// Open calls Open() on every fs in the chain until it can succesfully open a
// file which it then returns wrapped in a ChainFile. If no fs in the chain can
// open the given path, the last error is returned.
func (chainFs *ChainFs) Open(path string) (http.File, error) {
	var err error
	for i, fs := range chainFs.chain {
		file, openErr := fs.Open(path)
		if openErr == nil {
			return &ChainFile{
				File: file,
				path: path,
				others: chainFs.chain[i+1:],
			}, nil
		}
		err = openErr
	}
	return nil, err
}

// ChainFile acts as a proxy to an embedded http.File. It intercepts all
// Readdir() calls.
type ChainFile struct {
	http.File
	others []http.FileSystem
	path string
}

// Readdir returns the same results as seen by the parent ChainFs.
func (file *ChainFile) Readdir(count int) ([]os.FileInfo, error) {
	dirs, err := file.File.Readdir(count)
	if err != nil {
		return dirs, err
	}

	for _, fs := range file.others {
		otherFile, err := fs.Open(file.path)
		if err != nil {
			continue
		}

		remaining := 0
		if count > 0 {
			remaining = count - len(dirs)
		}

		otherDirs, err := otherFile.Readdir(remaining)
		if err != nil {
			continue
		}

		dirs = append(dirs, otherDirs...)
	}

	return removeDuplicates(dirs), nil
}

func removeDuplicates(stats []os.FileInfo) []os.FileInfo {
	names := make(map[string]bool)
	results := make([]os.FileInfo, 0, len(stats))

	for _, stat := range stats {
		name := stat.Name()
		if _, ok := names[name]; ok {
			continue
		}

		names[name] = true
		results = append(results, stat)
	}
	return results
}
