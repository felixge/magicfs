package vfs

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

type digitProcessor struct{}

func (d *digitProcessor) Process(r io.Reader) io.Reader {
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
	fs := NewProcessFs(parentFs, &digitProcessor{})

	file, err := fs.Open(path)
	if err != nil {
		t.Fatalf("Open: %s", err)
	}

	return file
}

// verify that Read() returns the processed file
func TestProcessFs_Read(t *testing.T) {
	file := open("/digits.txt", t)

	expectedData := "123456"
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

	expectedSize := int64(6)
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

// verify that Seek() with whence = os.SEEK_SET works
func TestProcessFs_Seek_Set(t *testing.T) {
	file := open("/digits.txt", t)

	firstOffset := int64(2)
	if offset, err := file.Seek(firstOffset, os.SEEK_SET); err != nil {
		t.Fatalf("%s", err)
	} else if offset != firstOffset {
		t.Fatalf("expected offset: %d, got: %d", firstOffset, offset)
	}

	data := make([]byte, 2)
	expectedData := "34"
	if _, err := io.ReadFull(file, data); err != nil {
		t.Fatalf("%s", err)
	} else if string(data) != expectedData {
		t.Fatalf(`exepected: "%s", got: "%s"`, expectedData, data)
	}
	
	secondOffset := int64(5)
	if offset, err := file.Seek(secondOffset, os.SEEK_SET); err != nil {
		t.Fatalf("%s", err)
	} else if offset != secondOffset {
		t.Fatalf("expected offset: %d, got: %d", secondOffset, offset)
	}

	expectedData = "6"
	if data, err := ioutil.ReadAll(file); err != nil {
		t.Fatalf("%s", err)
	} else if string(data) != expectedData {
		t.Fatalf(`exepected: "%s", got: "%s"`, expectedData, data)
	}
}

// verify that Seek() with whence = os.SEEK_CUR works
func TestProcessFs_Seek_Cur(t *testing.T) {
	file := open("/digits.txt", t)

	firstOffset := int64(2)
	if offset, err := file.Seek(firstOffset, os.SEEK_CUR); err != nil {
		t.Fatalf("%s", err)
	} else if offset != firstOffset {
		t.Fatalf("expected offset: %d, got: %d", firstOffset, offset)
	}

	data := make([]byte, 2)
	expectedData := "34"
	if _, err := io.ReadFull(file, data); err != nil {
		t.Fatalf("%s", err)
	} else if string(data) != expectedData {
		t.Fatalf(`exepected: "%s", got: "%s"`, expectedData, data)
	}
	
	secondOffset := int64(1)
	if offset, err := file.Seek(secondOffset, os.SEEK_CUR); err != nil {
		t.Fatalf("%s", err)
	} else if offset != 5 {
		t.Fatalf("expected offset: %d, got: %d", 5, offset)
	}

	expectedData = "6"
	if data, err := ioutil.ReadAll(file); err != nil {
		t.Fatalf("%s", err)
	} else if string(data) != expectedData {
		t.Fatalf(`exepected: "%s", got: "%s"`, expectedData, data)
	}

	thirdOffset := int64(-2)
	if offset, err := file.Seek(thirdOffset, os.SEEK_CUR); err != nil {
		t.Fatalf("%s", err)
	} else if offset != 4 {
		t.Fatalf("expected offset: %d, got: %d", 4, offset)
	}

	expectedData = "56"
	if data, err := ioutil.ReadAll(file); err != nil {
		t.Fatalf("%s", err)
	} else if string(data) != expectedData {
		t.Fatalf(`exepected: "%s", got: "%s"`, expectedData, data)
	}
}

// verify that Stat() and Read() work together
func TestProcessFs_Read_Stat_Read(t *testing.T) {
	t.Logf("Skipping: Implementing Seek() first")
	return

	file := open("/digits.txt", t)

	buf := make([]byte, 2)
	n, err := io.ReadFull(file, buf)
	if err != nil {
		t.Fatalf("Read #1: %s", err)
	}

	data := string(buf[:n])
	expectedData := "12"
	if data != expectedData {
		t.Fatalf("Read #1: expected: %s, got: %s", expectedData, data)
	}

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
