package magicfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

var errBadSeek = fmt.Errorf("processFs: bad seek")

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
	offset		int64
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
		file.resetReader()
	}

	n, err := file.reader.Read(buf)
	file.offset += int64(n)
	return n, err
}

func (file *file) resetReader() {
	file.offset = 0
	file.reader = file.processor.Process(file.File)

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

func (file *file) Seek(offset int64, whence int) (int64, error) {
	switch whence{
		case os.SEEK_SET:
			if offset, err := file.File.Seek(0, os.SEEK_SET); err != nil {
				return offset, err
			} else if offset != 0 {
				return offset, errBadSeek
			}

			file.resetReader()

			return file.Seek(offset - file.offset, os.SEEK_CUR)
		case os.SEEK_CUR:
			if offset < 0 {
				offset = file.offset + offset
				if _, err := file.Seek(0, os.SEEK_SET); err != nil {
					return file.offset, err
				}
			}

			n, err := io.CopyN(ioutil.Discard, file, offset)
			if err != nil {
				return n, err
			} else if n != offset {
				return n, errBadSeek
			}

			return file.offset, nil
	}

	return 0, fmt.Errorf(
		"processFs: whence: %d not implemented for Seek() yet",
		whence,
	)
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
