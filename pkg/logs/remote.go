package logs

import (
	"bytes"
	"net/http"
)

type RemoteWrite struct {
	Url string
}

func (r RemoteWrite) Write(p []byte) (n int, err error) {
	n = len(p)
	_, err = http.DefaultClient.Post(r.Url, "", bytes.NewBuffer(p))
	return
}

func (r RemoteWrite) Sync() error {
	return nil
}
