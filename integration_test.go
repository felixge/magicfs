package magicfs

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
)

type md5Processor struct{}

func (md5 *md5Processor) Process(input io.Reader) io.Reader {
	return bytes.NewBuffer([]byte("hello world"))
}

func Test_Get(t *testing.T) {
	fs := New(http.Dir(fixturesDir))
	fs.Process("*.txt", &md5Processor{})

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	
	addr := listener.Addr()

	go http.Serve(listener, http.FileServer(fs))

	url := "http://" + addr.String()
	res, err := http.Get(url + "/digits.txt")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(res)

	expected := "hello world"
	if string(data) != expected {
		t.Fatalf(`expected: %s, got: %s`, expected, data)
	}
}
