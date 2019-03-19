package lzf

/*
lzf算法解压缩
参考: https://cloud.tencent.com/developer/article/1013207
*/
func Decompress(input []byte, inputLength int, output []byte, outputLength int) {
	var iidx = 0
	var oidx = 0

	for iidx < inputLength {
		var ctrl = int(input[iidx])
		iidx++

		if ctrl < (1 << 5) {
			ctrl++

			if oidx+ctrl > outputLength {
				return
			}

			for ; ctrl > 0; ctrl-- {
				output[oidx] = input[iidx]
				oidx++
				iidx++
			}
		} else {
			var length = ctrl >> 5
			var reference = int(oidx - ((ctrl & 0x1f) << 8) - 1)
			if length == 7 {
				length += int(input[iidx])
				iidx++
			}
			reference -= int(input[iidx])
			iidx++
			if oidx+length+2 > outputLength || reference < 0 {
				return
			}

			output[oidx] = output[reference]
			oidx++
			reference++
			output[oidx] = output[reference]
			oidx++
			reference++

			for ; length > 0; length-- {
				output[oidx] = output[reference]
				oidx++
				reference++
			}
		}
	}
}
