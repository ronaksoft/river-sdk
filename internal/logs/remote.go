package logs

import (
	"bytes"
	"net/http"
)

type RemoteWrite struct {
	HttpClient http.Client
	Url        string
}

func (r RemoteWrite) Write(p []byte) (n int, err error) {
	n = len(p)
	_, err = r.HttpClient.Post(r.Url, "", bytes.NewBuffer(p))
	return
}

func (r RemoteWrite) Sync() error {
	return nil
}
