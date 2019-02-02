package main

const Cr = '\r'
const Lf = '\n'
const Star = '*'
const Dollar = '$'
const Plus = '+'
const Minus = '-'
const Colon = ':'

const Aux = 0xFA
const DbSelector = 0xFE
const DbResize = 0xFB
const String = 0
const Eof = 0xFF
const HashZipList = 13
const ListQuickList = 14

const ZipInt8Bit = 0xFE  /* 11111110*/
const ZipInt16Bit = 0xC0 /* 11000000*/
const ZipInt24Bit = 0xF0 /* 11110000*/
const ZipInt32Bit = 0xD0 /* 11010000*/
const ZipInt64Bit = 0xE0 /* 11100000*/
