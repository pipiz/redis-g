package util

import "strconv"

func ToInt16(b []byte, isBigEndian bool) int16 {
	_ = b[1]
	if isBigEndian {
		return int16(b[1]) | int16(b[0])<<8
	} else {
		return int16(b[0]) | int16(b[1])<<8
	}
}

func ToInt32(b []byte, isBigEndian bool) int32 {
	_ = b[3]
	if isBigEndian {
		return int32(b[3]) | int32(b[2])<<8 | int32(b[1])<<16 | int32(b[0])<<24
	} else {
		return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
	}
}

func ToInt64(b []byte, isBigEndian bool) int64 {
	_ = b[7]
	if isBigEndian {
		return int64(b[7]) | int64(b[6])<<8 | int64(b[5])<<16 | int64(b[4])<<24 |
			int64(b[3])<<32 | int64(b[2])<<40 | int64(b[1])<<48 | int64(b[0])<<56
	} else {
		return int64(b[0]) | int64(b[1])<<8 | int64(b[2])<<16 | int64(b[3])<<24 |
			int64(b[4])<<32 | int64(b[5])<<40 | int64(b[6])<<48 | int64(b[7])<<56
	}
}

func ToBytes(i int) []byte {
	return []byte(strconv.Itoa(i))
}
