package main

import "errors"

func BlockCopy(src []byte, srcOffset int, dst []byte, dstOffset, count int) (err error) {
	srcLen := len(src)
	if srcOffset > srcLen || count > srcLen || srcOffset+count > srcLen {
		err = errors.New("源缓冲区:索引超出范围。")
		return
	}
	dstLen := len(dst)
	if dstOffset > dstLen || count > dstLen || dstOffset+count > dstLen {
		err = errors.New("目标缓冲:索引超出范围。")
		return
	}
	index := 0
	for i := srcOffset; i < srcOffset+count; i++ {
		dst[dstOffset+index] = src[srcOffset+index]
		index++
	}
	return
}
