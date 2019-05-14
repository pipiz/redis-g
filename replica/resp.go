package replica

import (
	"strconv"
	"strings"
)

func replyStr() (reply string) {
	return parseReply(nil).(string)
}

func handleBulkReply(length int) (resp interface{}) {
	if length == 0 {
		return []byte{}
	}
	bytes := make([]byte, length)
	reader.Read(bytes)

	b, _ := reader.ReadByte()
	if b != Cr {
		panic("excepted cr but not")
	}
	b, _ = reader.ReadByte()
	if b != Lf {
		panic("excepted lf but not")
	}
	return bytes
}

func parseDump() {
	parseReply(parseRdb)
}

func parse(callback func(length int)) interface{} {
	reader.Mark()
	reply := parseReply(handleBulkReply)
	length := reader.UnMark()
	callback(length)
	return reply
}

func parseReply(callback func(length int) interface{}) interface{} {
	for {
		b, _ := reader.ReadByte()
		switch b {
		case Plus: // RESP Simple Strings
			var builder strings.Builder
			for {
				for byt, e := reader.ReadByte(); byt != Cr && e == nil; {
					builder.WriteByte(byt)
					byt, _ = reader.ReadByte()
				}
				if byt, e := reader.ReadByte(); byt == Lf {
					return builder.String()
				} else if e == nil {
					builder.WriteByte(byt)
				}
			}
		case Minus: // RESP Errors
			var builder strings.Builder
			for {
				for byt, e := reader.ReadByte(); byt != Cr && e == nil; {
					builder.WriteByte(byt)
					byt, e = reader.ReadByte()
				}
				if byt, e := reader.ReadByte(); byt == Lf {
					return builder.String()
				} else if e == nil {
					builder.WriteByte(byt)
				}
			}
		case Dollar: // RESP Bulk Strings
			var builder strings.Builder
			for {
				for byt, e := reader.ReadByte(); byt != Cr && e == nil; {
					builder.WriteByte(byt)
					byt, e = reader.ReadByte()
				}
				if byt, e := reader.ReadByte(); byt == Lf {
					break
				} else if e == nil {
					builder.WriteByte(byt)
				}
			}
			resp := builder.String()
			var size int
			if !strings.HasPrefix(resp, "EOF:") {
				_size, err := strconv.Atoi(resp)
				size = _size
				if err != nil {
					return -1
				}
			}
			return callback(size)
		case Star:
			var builder strings.Builder
			for {
				for byt, e := reader.ReadByte(); byt != Cr && e == nil; {
					builder.WriteByte(byt)
					byt, e = reader.ReadByte()
				}
				if byt, e := reader.ReadByte(); byt == Lf {
					break
				} else if e == nil {
					builder.WriteByte(byt)
				}
			}
			length, _ := strconv.Atoi(builder.String())
			if length == -1 {
				return nil
			}
			array := make([][]byte, length)
			for i := 0; i < length; i++ {
				reply := parseReply(handleBulkReply)
				typ, ok := reply.([]byte)
				if ok {
					array[i] = typ
				}
			}
			return array
		case Colon:
			// TODO
			panic("Colon: not implement")
		case '\n':
			break
		default:
			break
		}
	}
	return ""
}
