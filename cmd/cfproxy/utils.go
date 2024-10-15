package main

import (
	"io"
)

type prependedReader struct {
	r         io.Reader
	prepend   []byte
	prepended bool
}

func (pr *prependedReader) Read(p []byte) (n int, err error) {
	if !pr.prepended {
		n = copy(p, pr.prepend)
		pr.prepended = true
		pr.prepend = nil
		return
	}
	return pr.r.Read(p)
}

func newPrependedReader(r io.Reader, prepend []byte) *prependedReader {
	return &prependedReader{
		r:       r,
		prepend: prepend,
	}
}

func newReaderWriter(r io.Reader, w io.Writer) io.ReadWriter {
	return struct {
		io.Reader
		io.Writer
	}{r, w}
}
