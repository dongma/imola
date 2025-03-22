package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
)

type GzipCompresser struct {
}

func (g GzipCompresser) Code() byte {
	return 1
}

func (g GzipCompresser) Compress(data []byte) ([]byte, error) {
	// res := &bytes.Buffer{}
	res := bytes.NewBuffer(nil)
	w := gzip.NewWriter(res)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Flush()
	if err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (g GzipCompresser) UnCompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = r.Close()
	}()
	res, err := ioutil.ReadAll(r)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return res, nil
}
