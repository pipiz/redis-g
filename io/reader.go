package io

import (
	"bufio"
	"io"
)

type Reader struct {
	Input   *bufio.Reader
	markLen int
	marked  bool
}

func (reader *Reader) ReadByte() (byte, error) {
	oneByte, err := reader.Input.ReadByte()
	if reader.marked {
		reader.markLen += 1
	}
	return oneByte, err
}

func (reader *Reader) Read(p []byte) (n int, err error) {
	i, e := io.ReadFull(reader.Input, p)
	if reader.marked {
		reader.markLen += i
	}
	return i, e
}

func (reader *Reader) Mark() {
	if !reader.marked {
		reader.marked = true
	}
}

func (reader *Reader) UnMark() (rs int) {
	if reader.marked {
		rs = reader.markLen
		reader.markLen = 0
		reader.marked = false
		return rs
	}
	panic("unmarked")
}
