package vfs

import (
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"runtime"
	"testing"
)

var (
	_, filename, _, _ = runtime.Caller(0)
	fixturesDir       = path.Join(path.Dir(filename), "fixtures")
)

type digitReader struct {
	r io.Reader
}

func newDigitReader(r io.Reader) io.Reader {
	return &digitReader{r}
}

func (r *digitReader) Read(buf []byte) (int, error) {
	n, err := r.r.Read(buf)
	if err != nil {
		return n, err
	}

	digitIndex := 0
	for i := 0; i < n; i++ {
		c := buf[i]
		if c >= '0' && c <= '9' {
			buf[digitIndex] = c
			digitIndex++
		}
	}

	return digitIndex, nil
}

func open(path string, t *testing.T) http.File {
	parentFs := http.Dir(fixturesDir)
	fs := NewProcessFs(parentFs, newDigitReader)

	file, err := fs.Open(path)
	if err != nil {
		t.Fatalf("Open: %s", err)
	}

	return file
}

// verify that Read() returns the processed file
func TestProcessFs_Read(t *testing.T) {
	file := open("/digits.txt", t)

	expectedData := "12345"
	if data, err := ioutil.ReadAll(file); err != nil {
		t.Fatalf("ReadAll: %s", err)
	} else if string(data) != expectedData {
		t.Fatalf(`ReadAll: exepected: "%s", got: "%s"`, expectedData, data)
	}
}

// verify that Stat() returns the file size of the processed file
func TestProcessFs_Stat_Size(t *testing.T) {
	file := open("/digits.txt", t)

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Stat: %s", err)
	}

	expectedSize := int64(5)
	size := stat.Size()
	if size != expectedSize {
		t.Fatalf("Wrong size: expected: %d, got: %d", expectedSize, size)
	}
}

// verify that directories are left untouched by the processor
func TestProcessFs_Stat_IsDir(t *testing.T) {
	file := open("/", t)

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Stat: %s", err)
	}

	if !stat.IsDir() {
		t.Errorf("exected directory, got file")
	}
}
