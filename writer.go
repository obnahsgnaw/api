package api

import (
	"io"
)

type MultiWriter struct {
	w []io.Writer
}

func newMultiWriter(w ...io.Writer) *MultiWriter {
	return &MultiWriter{
		w: w,
	}
}

func (s *MultiWriter) Write(p []byte) (int, error) {
	for _, w := range s.w {
		if n, err := w.Write(p); err != nil {
			return n, err
		}
	}
	return len(p), nil
}
