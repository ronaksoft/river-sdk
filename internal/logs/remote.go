package logs

import (
	"bytes"
	"github.com/Workiva/go-datastructures/queue"
	"github.com/ronaksoft/rony/pools"
	"net/http"
	"time"
)

var (
	remoteLogWriter *RemoteWrite
)

type RemoteWrite struct {
	HttpClient http.Client
	Url        string
	rBuf       *queue.RingBuffer
}

func newRemoteWrite(url string) *RemoteWrite {
	rw := &RemoteWrite{
		HttpClient: http.Client{
			Timeout: time.Millisecond * 250,
		},
		Url:  url,
		rBuf: queue.NewRingBuffer(100),
	}
	go rw.flusher()
	remoteLogWriter = rw
	return rw
}

func (r *RemoteWrite) flusher() {
	var (
		writeBuff = &bytes.Buffer{}
		qty       = 0
	)
	for {
		v, err := r.rBuf.Get()
		if err != nil {
			continue
		}
		chunkBuff := v.(*pools.ByteBuffer)
		writeBuff.Write(*chunkBuff.Bytes())
		pools.Buffer.Put(chunkBuff)
		if r.rBuf.Len() > 0 && qty < 10 {
			qty++
			continue
		}
		_, err = r.HttpClient.Post(r.Url, "", writeBuff)
		writeBuff.Reset()
	}
}

func (r *RemoteWrite) Write(p []byte) (n int, err error) {
	buf := pools.Buffer.FromBytes(p)
	return len(p), r.rBuf.Put(buf)
}

func (r *RemoteWrite) Sync() error {
	return nil
}
