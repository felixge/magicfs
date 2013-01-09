package vfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type Processor interface{
	Process(r io.Reader) io.Reader
}

type processFs struct {
	parent    http.FileSystem
	processor Processor
}

type file struct {
	http.File
	processor Processor
	reader    io.Reader
}

// TODO(felixge) This type should be named "stat", but I get a comiler error.
type processedStat struct {
	os.FileInfo
	size int64
}

func (stat *processedStat) Size() int64 {
	return stat.size
}

func (file *file) Read(buf []byte) (int, error) {
	if file.reader == nil {
		file.reader = file.processor.Process(file.File)
	}

	return file.reader.Read(buf)
}

func (file *file) Stat() (os.FileInfo, error) {
	stat, err := file.File.Stat()
	if err != nil || stat.IsDir() {
		return stat, err
	}

	size, err := io.Copy(ioutil.Discard, file)
	if err != nil {
		return nil, fmt.Errorf("processFs: could not determine size: %s", err)
	}

	return &processedStat{FileInfo: stat, size: size}, nil
}

func NewProcessFs(parent http.FileSystem, processor Processor) http.FileSystem {
	return &processFs{
		parent:    parent,
		processor: processor,
	}
}

func (fs *processFs) Open(path string) (http.File, error) {
	input, err := fs.parent.Open(path)
	if err != nil {
		return input, err
	}

	output := &file{File: input, processor: fs.processor}
	return output, nil
}
