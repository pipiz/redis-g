package io

import (
	"bufio"
)

type MyReader struct {
	Input   *bufio.Reader
	markLen int
	marked  bool
}

func (reader *MyReader) ReadByte() (byte, error) {
	bytes := make([]byte, 1)
	_, err := reader.Input.Read(bytes)
	return bytes[0], err
}

func (reader *MyReader) Read(p []byte) (n int, err error) {
	i, e := reader.Input.Read(p)
	if reader.marked {
		reader.markLen += len(p)
	}
	return i, e
}

func (reader *MyReader) Mark() {
	if !reader.marked {
		reader.marked = true
	}
}

func (reader *MyReader) UnMark() (rs int) {
	if reader.marked {
		rs = reader.markLen
		reader.markLen = 0
		reader.marked = false
		return rs
	}
	panic("unmarked")
}
