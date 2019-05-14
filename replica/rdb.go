package replica

import (
	bytes2 "bytes"
	"encoding/binary"
	"fmt"
	"redis-g/command"
	"redis-g/utils/lzf"
	"redis-g/utils/numbers"
	"strconv"
)

func parseRdb(size int) interface{} {
	if size != -1 {
		logger.Printf("RDB size: %d byte\n", size)
	}
	bytes := make([]byte, 5)
	reader.Read(bytes)

	reader.Read(bytes[:4])
	version, _ := strconv.Atoi(string(bytes[:4]))
	if version < 2 || version > 9 {
		panic(fmt.Sprintf("can't handle RDB format version %d", version))
	}
	logger.Println("RDB version:", version)
	for {
		_type, _ := reader.ReadByte()
		switch _type {
		case Aux:
			parseAUX()
		case DbSelector:
			length, _ := readLength()
			logger.Println("db:", length)
		case DbResize:
			len1, _ := readLength()
			len2, _ := readLength()
			logger.Println("db total keys:", len1)
			logger.Println("db expired keys:", len2)
		case String:
			commChan <- parseSetCommand()
		case HashZipList:
			key := readString()
			bytes := readString()
			byteReader := bytes2.NewReader(bytes)
			readZlBytes(byteReader)
			readZlTail(byteReader)

			zlLen := readZlLen(byteReader)
			args := make([][]byte, zlLen+1)
			args[0] = key
			for index := 1; zlLen > 0; zlLen -= 2 {
				field := readZipListEntry(byteReader)
				value := readZipListEntry(byteReader)
				args[index] = field
				args[index+1] = value
				index += 2
			}
			commChan <- command.New("HSET", args)
		case ListQuickList:
			key := readString()
			count, _ := readLength()
			args := make([][]byte, 1)
			args[0] = key
			for i := 0; i < count; i++ {
				element := readString()
				byteReader := bytes2.NewReader(element)
				readZlBytes(byteReader)
				readZlTail(byteReader)

				zlLen := readZlLen(byteReader)
				for ; zlLen > 0; zlLen-- {
					element := readZipListEntry(byteReader)
					args = append(args, element)
				}
			}
			commChan <- command.New("RPUSH", args)
		case Set, List:
			key := readString()

			count, _ := readLength()
			args := make([][]byte, count+1)
			args[0] = key
			for i := 1; i <= count; i++ {
				element := readString()
				args[i] = element
			}
			commandName := "SADD"
			if _type == List {
				commandName = "RPUSH"
			}
			commChan <- command.New(commandName, args)
		case ZSetZipList:
			key := readString()

			bytes := readString()
			byteReader := bytes2.NewReader(bytes)
			readZlBytes(byteReader)
			readZlTail(byteReader)

			zlLen := readZlLen(byteReader)
			args := make([][]byte, zlLen+1)
			args[0] = key
			for index := 1; zlLen > 0; zlLen -= 2 {
				element := readZipListEntry(byteReader)
				score := readZipListEntry(byteReader)
				args[index] = score
				args[index+1] = element
				index += 2
			}
			commChan <- command.New("ZADD", args)
		case ZSet:
			// TODO
			panic("ZSet: not implement")
		case Hash:
			// TODO
			panic("Hash: not implement")
		case ZSet2:
			// TODO
			panic("ZSet2: not implement")
		case Module:
			// TODO
			panic("Module: not implement")
		case Module2:
			// TODO
			panic("Module2: not implement")
		case HashZipMap:
			// TODO
			panic("HashZipMap: not implement")
		case ListZipList:
			// TODO
			panic("ListZipList: not implement")
		case SetIntSet:
			// TODO
			panic("SetIntSet: not implement")
		case StreamListPacks:
			// TODO
			panic("StreamListPacks: not implement")
		case Eof:
			if version >= 5 {
				checksum := readInteger(8, true)
				logger.Println("checksum ", checksum)
			}
			return "OK"
		}
	}
	return "OK"
}

func parseSetCommand() command.Command {
	key := readString()
	value := readString()
	args := make([][]byte, 2)
	args[0] = key
	args[1] = value
	return command.New("SET", args)
}

func readZlLen(byteReader *bytes2.Reader) uint16 {
	bytes := make([]byte, 2)
	byteReader.Read(bytes)
	zlLen := binary.LittleEndian.Uint16(bytes)
	return zlLen
}

func readZlBytes(byteReader *bytes2.Reader) {
	arr := make([]byte, 4)
	byteReader.Read(arr)
}

func readZlTail(byteReader *bytes2.Reader) {
	arr := make([]byte, 4)
	byteReader.Read(arr)
}

func parseAUX() {
	logger.Printf("%s: %s\n", string(readString()), string(readString()))
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
		u := numbers.ToInt16(bytes, false)
		return []byte(strconv.Itoa(int(u)))
	case ZipInt24Bit: // little endian
		bytes := make([]byte, 4)
		bytes[0], _ = byteReader.ReadByte()
		bytes[1], _ = byteReader.ReadByte()
		bytes[2], _ = byteReader.ReadByte()
		u := numbers.ToInt32(bytes, false)
		return []byte(strconv.Itoa(int(u)))
	case ZipInt32Bit: // little endian
		bytes := make([]byte, 4)
		byteReader.Read(bytes)
		u := numbers.ToInt32(bytes, false)
		return []byte(strconv.Itoa(int(u)))
	case ZipInt64Bit: // little endian
		bytes := make([]byte, 8)
		byteReader.Read(bytes)
		u := numbers.ToInt64(bytes, false)
		return []byte(strconv.Itoa(int(u)))
	default:
		result := specialFlag - 0xF1
		return []byte(strconv.Itoa(int(result)))
	}
	return nil
}

func readString() []byte {
	length, special := readLength()
	if special {
		switch length {
		case 0:
			b, _ := reader.ReadByte()
			return []byte(strconv.Itoa(int(int8(b))))
		case 1:
			integer := readInteger(2, false)
			return []byte(strconv.Itoa(int(integer)))
		case 2:
			integer := readInteger(4, false)
			return []byte(strconv.Itoa(int(integer)))
		case 3:
			clenth, _ := readLength()
			length, _ := readLength()
			bytes := make([]byte, clenth)
			reader.Read(bytes)
			out := make([]byte, length)
			lzf.Decompress(bytes, clenth, out, length)
			return out
		}
	}
	bytes := make([]byte, length)
	reader.Read(bytes)
	return bytes
}

func readLength() (length int, special bool) {
	b, _ := reader.ReadByte()

	var rawByte = int(b & 0xff)

	_type := (rawByte & 0xC0) >> 6

	if _type == 3 {
		length = int(b & 0x3F)
		special = true
	} else if _type == 0 {
		length = int(b & 0x3F)
	} else if _type == 1 {
		nextByte, _ := reader.ReadByte()
		i := ((int16(b) & 0x3F) << 8) | int16(nextByte)
		length = int(i)
	} else if rawByte == 0x80 {
		length = readInteger(4, true)
	} else if rawByte == 0x81 {
		length = readInteger(8, true)
	}
	return length, special
}

func readInteger(size int, isBigEndian bool) int {
	bytes := make([]byte, size)
	reader.Read(bytes)
	if size == 2 {
		return int(numbers.ToInt16(bytes, isBigEndian))
	} else if size == 4 {
		return int(numbers.ToInt32(bytes, isBigEndian))
	} else if size == 8 {
		return int(numbers.ToInt64(bytes, isBigEndian))
	}
	return -1
}
