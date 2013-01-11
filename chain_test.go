package magicfs

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func TestChain(t *testing.T) {
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
}

func readFile(fs http.FileSystem, path string) ([]byte, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}