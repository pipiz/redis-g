package util

/*
lzf算法解压缩
参考: https://cloud.tencent.com/developer/article/1013207
*/
func Decompress(input []byte, inputLength int, output []byte, outputLength int) {
	var iidx uint8 = 0
	var oidx uint8 = 0

	for {
		var ctrl = input[iidx]
		iidx++

		if ctrl < (1 << 5) {
			ctrl++

			if int(oidx+ctrl) > outputLength {
				return
			}

			for {
				output[oidx] = input[iidx]
				oidx++
				iidx++
				ctrl--
				if ctrl == 0 {
					break
				}
			}
		} else {
			var length = ctrl >> 5
			var reference = int(oidx - ((ctrl & 0x1f) << 8) - 1)
			if length == 7 {
				length += input[iidx]
				iidx++
			}
			reference -= int(input[iidx])
			iidx++
			if int(oidx+length+2) > outputLength {
				return
			}
			if reference < 0 {
				return
			}

			output[oidx] = output[reference]
			oidx++
			reference++
			output[oidx] = output[reference]
			oidx++
			reference++

			for {
				output[oidx] = output[reference]
				oidx++
				reference++
				length--
				if length == 0 {
					break
				}
			}
		}

		if !(int(iidx) < inputLength) {
			break
		}
	}
}
