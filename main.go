package main

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func main() {
	d, _ := hex.DecodeString("531100000000FFFFFFFF000000200000000000")
	mp := new(Packer)
	mp.Init()
	b := mp.Compress(d)
	fmt.Println(strings.ToUpper(hex.EncodeToString(b)))
}
