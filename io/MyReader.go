package io

import (
	"bufio"
	"io"
	"log"
)

type MyReader struct {
	Input   *bufio.Reader
	markLen int
	marked  bool
}

func (reader *MyReader) ReadByte() (byte, error) {
	oneByte, err := reader.Input.ReadByte()
	if reader.marked {
		reader.markLen += 1
	}
	return oneByte, err
}

func (reader *MyReader) Read(p []byte) (n int, err error) {
	i, e := io.ReadFull(reader.Input, p)
	log.Println("len: ", len(p), "实际读取的: ", i)
	if reader.marked {
		reader.markLen += i
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
