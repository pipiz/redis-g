package main

import (
	bytes2 "bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

type Length struct {
	val     int
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
			fmt.Println("db:", length.val)
		case DbResize:
			len1 := readLength()
			len2 := readLength()
			fmt.Println("db total keys:", len1.val)
			fmt.Println("db expired keys:", len2.val)
		case String:
			fmt.Printf("%s: %s\n", readString(), readString())
		case HashZipList:
			name := readString()
			fmt.Printf("%s: {\n", name)

			length := readLength()
			bytes := make([]byte, length.val)
			reader.Read(bytes)
			byteReader := bytes2.NewReader(bytes)

			arr := make([]byte, 4)
			byteReader.Read(arr)

			byteReader.Read(arr)

			arr = arr[:2]
			byteReader.Read(arr)
			zlLen := binary.LittleEndian.Uint16(arr)

			for ; zlLen > 0; zlLen -= 2 {
				field := readZipListEntry(byteReader)
				value := readZipListEntry(byteReader)
				fmt.Printf("\t%s: %s\n", string(field), string(value))
			}
			fmt.Println("}")
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

func readZipListEntry(byteReader *bytes2.Reader) []byte {
	byt, _ := byteReader.ReadByte()
	prevLen := uint32(byt)
	if prevLen >= 254 {
		bytes := make([]byte, 4)
		byteReader.Read(bytes)
		prevLen = binary.LittleEndian.Uint32(bytes)
	}
	specialFlag, _ := byteReader.ReadByte()
	switch specialFlag >> 6 {
	case 0:
		length := specialFlag & 0x3F // 余下6位就是长度
		bytes := make([]byte, length)
		byteReader.Read(bytes)
		return bytes
	case 1:
		b, _ := byteReader.ReadByte() // 再读一个字节, 合并14位就是长度
		length := ((int16(specialFlag) & 0x3F) << 8) | int16(b)
		bytes := make([]byte, length)
		byteReader.Read(bytes)
		return bytes
	case 2:
		bytes := make([]byte, 4) // 再读四个字节，就是长度
		byteReader.Read(bytes)
		length := binary.BigEndian.Uint32(bytes)
		bytes = make([]byte, length)
		byteReader.Read(bytes)
		return bytes
	default:
		break
	}
	switch specialFlag {
	case ZipInt8Bit:
		b, _ := byteReader.ReadByte()
		return []byte(strconv.Itoa(int(int8(b))))
	case ZipInt16Bit: // little endian
		bytes := make([]byte, 2)
		byteReader.Read(bytes)
		u := toInt16(bytes, false)
		return []byte(strconv.Itoa(int(u)))
	case ZipInt24Bit: // little endian
		bytes := make([]byte, 4)
		bytes[0], _ = byteReader.ReadByte()
		bytes[1], _ = byteReader.ReadByte()
		bytes[2], _ = byteReader.ReadByte()
		u := toInt32(bytes, false)
		return []byte(strconv.Itoa(int(u)))
	case ZipInt32Bit: // little endian
		bytes := make([]byte, 4)
		byteReader.Read(bytes)
		u := toInt32(bytes, false)
		return []byte(strconv.Itoa(int(u)))
	case ZipInt64Bit: // little endian
		bytes := make([]byte, 8)
		byteReader.Read(bytes)
		u := toInt64(bytes, false)
		return []byte(strconv.Itoa(int(u)))
	default:
		result := specialFlag - 0xF1
		return []byte(strconv.Itoa(int(result)))
	}
	return nil
}

func readString() string {
	length := readLength()
	if length.Special {
		switch length.val {
		case 0:
			b, _ := reader.ReadByte()
			return strconv.Itoa(int(b))
		case 1:
			integer := readInteger(2, false)
			return strconv.Itoa(int(integer))
		case 2:
			integer := readInteger(4, false)
			return strconv.Itoa(int(integer))
		case 4:
			// TODO LZF压缩的字符串
		}
	}
	bytes := make([]byte, length.val)
	reader.Read(bytes)
	return string(bytes)
}

func readLength() Length {
	var length int
	var special bool

	b, _ := reader.ReadByte()

	switch (b & 0xC0) >> 6 { // 取高二位
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
	if size == 2 {
		return int(toInt16(bytes, isBigEndian))
	} else if size == 4 {
		return int(toInt32(bytes, isBigEndian))
	} else if size == 8 {
		return int(toInt64(bytes, isBigEndian))
	}
	return -1
}
