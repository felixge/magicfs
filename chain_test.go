package magicfs

import (
	"io/ioutil"
	"net/http"
	"os"
	"syscall"
	"testing"
)

func TestChain_Open(t *testing.T) {
	aFs := http.Dir(fixturesDir+"/a")
	bFs := http.Dir(fixturesDir+"/b")
	chainFs := NewChainFs(aFs, bFs)

	if data, err := readFile(chainFs, "/1.txt"); err != nil {
		t.Fatal(err)
	} else if string(data) != "a1\n" {
		t.Fatalf("unexpected: %s", data)
	}

	if data, err := readFile(chainFs, "/2.txt"); err != nil {
		t.Fatal(err)
	} else if string(data) != "b2\n" {
		t.Fatalf("unexpected: %s", data)
	}

	_, err := readFile(chainFs, "/does-not-exist.txt")
	if pathErr, ok := err.(*os.PathError); !ok {
		t.Fatalf("unexpected: %#v", err)
	} else if pathErr.Err != syscall.ENOENT {
		t.Fatalf("unexpected: %#v", pathErr.Err)
	}
}

func TestChain_Readdir(t *testing.T) {
	aFs := http.Dir(fixturesDir+"/a")
	bFs := http.Dir(fixturesDir+"/b")
	chainFs := NewChainFs(aFs, bFs)

	dir, err := chainFs.Open("/")
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Close()

	stats, err := dir.Readdir(0)
	if err != nil {
		t.Fatal(err)
	}

	var expected = map[string]int64{
		"1.txt": 3,
		"2.txt": 3,
		"3.txt": 3,
	}

	for _, stat := range stats {
		name := stat.Name()
		// skip hidden files (my editor creates some)
		if string(name[0]) == "." {
			continue
		}

		size := stat.Size()
		if expectedSize, ok := expected[name]; !ok {
			t.Errorf("unexpected file: %s", name)
		} else if size != expectedSize {
			t.Errorf("expected size: %d for %s, got: %d", expectedSize, name, size)
		} else {
			delete(expected, name)
		}
	}

	if len(expected) > 0 {
		t.Errorf("missing files: %+v", expected)
	}
}

func readFile(fs http.FileSystem, path string) ([]byte, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}
