package magicfs

import (
	"net/http"
)

func NewChainFs(chain... http.FileSystem) http.FileSystem {
	return &chainFs{chain: chain}
}

type chainFs struct {
	chain []http.FileSystem
}

func (chainFs *chainFs) Open(path string) (http.File, error) {
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
