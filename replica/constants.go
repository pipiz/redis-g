package replica

const (
	Cr     = '\r'
	Lf     = '\n'
	Star   = '*'
	Dollar = '$'
	Plus   = '+'
	Minus  = '-'
	Colon  = ':'

	Aux        = 0xFA
	DbSelector = 0xFE
	DbResize   = 0xFB
	Eof        = 0xFF

	String          = 0
	List            = 1
	Set             = 2
	ZSet            = 3
	Hash            = 4
	ZSet2           = 5
	Module          = 6
	Module2         = 7
	HashZipMap      = 9
	ListZipList     = 10
	SetIntSet       = 11
	ZSetZipList     = 12
	HashZipList     = 13
	ListQuickList   = 14
	StreamListPacks = 15

	ZipInt8Bit  = 0xFE // 11111110
	ZipInt16Bit = 0xC0 // 11000000
	ZipInt24Bit = 0xF0 // 11110000
	ZipInt32Bit = 0xD0 // 11010000
	ZipInt64Bit = 0xE0 // 11100000
)
