package cclient

// this code originally came from: https://github.com/x04/cclient/blob/master/client_test.go

import (
	"encoding/json"
	tls "github.com/refraction-networking/utls"
	"github.com/sleeyax/gotcha"
	"io"
	"io/ioutil"
	"testing"
)

const Chrome83Hash = "b32309a26951912be7dba376398abc3b"

var client, _ = gotcha.NewClient(&gotcha.Options{
	Adapter: NewAdapter(tls.HelloChrome_83),
})

type JA3Response struct {
	JA3Hash   string `json:"ja3_hash"`
	JA3       string `json:"ja3"`
	UserAgent string `json:"User-Agent"`
}

func readAndClose(r io.ReadCloser) ([]byte, error) {
	readBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return readBytes, r.Close()
}

func TestCClient_JA3(t *testing.T) {
	resp, err := client.Get("https://ja3er.com/json")
	if err != nil {
		t.Fatal(err)
	}

	respBody, err := readAndClose(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var ja3Response JA3Response
	if err := json.Unmarshal(respBody, &ja3Response); err != nil {
		t.Fatal(err)
	}

	if ja3Response.JA3Hash != Chrome83Hash {
		t.Error("unexpected JA3 hash; expected:", Chrome83Hash, "| got:", ja3Response.JA3Hash)
	}
}

func TestCClient_HTTP2(t *testing.T) {
	resp, err := client.Get("https://www.google.com")
	if err != nil {
		t.Fatal(err)
	}

	_, err = readAndClose(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.ProtoMajor != 2 || resp.ProtoMinor != 0 {
		t.Error("unexpected response proto; expected: HTTP/2.0 | got: ", resp.Proto)
	}
}
