package magicfs

import (
	"net/http"
)

type MagicFs struct{
	head http.FileSystem
}

func New(baseFs http.FileSystem) *MagicFs {
	return &MagicFs{head: baseFs}
}

func (fs *MagicFs) Open(path string) (http.File, error) {
	return fs.head.Open(path)
}

func (fs *MagicFs) Process(pattern string, processor Processor) {
	fs.head = NewProcessFs(fs.head, processor)
}
