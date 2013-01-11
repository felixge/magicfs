package magicfs

import (
	"net/http"
)

// NewChainFs returns a new ChainFs holding the given chain.
func NewChainFs(chain... http.FileSystem) http.FileSystem {
	return &ChainFs{chain: chain}
}

type ChainFs struct {
	chain []http.FileSystem
}

// Open calls Open() on every fs in the chain until it can succesfully open a
// file which it then returns. If no fs in the chain can open the given path,
// the last error is returned.
func (chainFs *ChainFs) Open(path string) (http.File, error) {
	var err error
	for _, fs := range chainFs.chain {
		file, openErr := fs.Open(path)
		if openErr == nil {
			return file, nil
		}
		err = openErr
	}
	return nil, err
}
