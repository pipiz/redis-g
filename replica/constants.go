package replica

const (
	Cr     = '\r'
	Lf     = '\n'
	Star   = '*'
	Dollar = '$'
	Plus   = '+'
	Minus  = '-'
	Colon  = ':'

	Aux           = 0xFA
	DbSelector    = 0xFE
	DbResize      = 0xFB
	String        = 0
	Eof           = 0xFF
	HashZipList   = 13
	ListQuickList = 14
	Set           = 2
	List          = 1
	ZsetZipList   = 12

	ZipInt8Bit  = 0xFE // 11111110
	ZipInt16Bit = 0xC0 // 11000000
	ZipInt24Bit = 0xF0 // 11110000
	ZipInt32Bit = 0xD0 // 11010000
	ZipInt64Bit = 0xE0 // 11100000
)
