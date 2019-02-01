package main

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

type Length struct {
	Len     int
	Special bool
}

func parseRdb(size int) string {
	if size != -1 {
		fmt.Printf("RDB size: %d byte\n", size)
	}
	bytes := make([]byte, 5)
	reader.Read(bytes)

	reader.Read(bytes[:4])
	version, _ := strconv.Atoi(string(bytes[:4]))
	if version < 2 || version > 9 {
		panic(fmt.Sprintf("can't handle RDB format version %d", version))
	}
	fmt.Println("RDB version:", version)
	for {
		_type, _ := reader.ReadByte()
		switch _type {
		case Aux:
			parseAUX()
		case DbSelector:
			length := readLength()
			fmt.Println("db:", length.Len)
		case DbResize:
			len1 := readLength()
			len2 := readLength()
			fmt.Println("db total keys:", len1.Len)
			fmt.Println("db expired keys:", len2.Len)
		case _string:
			fmt.Printf("key: %s, value: %s\n", readString(), readString())
		case Eof:
			if version >= 5 {
				checksum := readInteger(8, true)
				fmt.Println("checksum ", checksum)
			}
			return "OK"
		}
	}
	return "OK"
}

func parseAUX() {
	fmt.Printf("%s: %s\n", readString(), readString())
}

func readString() string {
	length := readLength()
	if length.Special {
		switch length.Len {
		case 0:
			b, _ := reader.ReadByte()
			return strconv.Itoa(int(b))
		case 1:
			integer := readInteger(2, true)
			return strconv.Itoa(int(integer))
		case 2:
			integer := readInteger(4, true)
			return strconv.Itoa(int(integer))
		case 4:
			// TODO LZF压缩的字符串
		}
	}
	bytes := make([]byte, length.Len)
	reader.Read(bytes)
	return string(bytes)
}

func readLength() Length {
	var length int
	var special bool

	b, _ := reader.ReadByte()

	switch b & 0xC0 >> 6 { // 取高二位
	case 0: // 00, 余下6位表示长度
		length = int(b & 0x3F)
	case 1: // 01, 再读一字节, 组合14位表示长度
		nextByte, _ := reader.ReadByte()
		i := ((int16(b) & 0x3F) << 8) | int16(nextByte)
		length = int(i)
	case 3: // 11, 特殊格式
		length = int(b & 0x3F)
		special = true
	case 0x80: // 再读4字节，表示长度
		length = readInteger(4, true)
	case 0x81: // 再读8字节，表示长度
		length = readInteger(8, true)
	}
	return Length{length, special}
}

func readInteger(size int, isBigEndian bool) int {
	bytes := make([]byte, size)
	reader.Read(bytes)
	if isBigEndian {
		if size == 2 {
			return int(binary.BigEndian.Uint16(bytes))
		} else if size == 4 {
			return int(binary.BigEndian.Uint32(bytes))
		} else if size == 8 {
			return int(binary.BigEndian.Uint64(bytes))
		}
	} else {
		if size == 2 {
			return int(binary.LittleEndian.Uint16(bytes))
		} else if size == 4 {
			return int(binary.LittleEndian.Uint32(bytes))
		} else if size == 8 {
			return int(binary.LittleEndian.Uint64(bytes))
		}
	}
	return -1
}
