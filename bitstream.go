package main

type BitStream struct {
	Buffer []byte
	Offset int
}

func (e *BitStream) Init(buf []byte) {
	e.Buffer = buf
	e.Offset = 0
}

func (e *BitStream) WriteBits(val, n int) {
	off := e.Offset & 7
	idx := e.Offset >> 3
	v := (val << (32 - n - off)) | int(e.Buffer[idx])<<24
	for i := 0; i < 4; i++ {
		e.Buffer[idx+i] = byte(v >> (24 - (i << 3)) & 0xFF)
	}
	e.Offset += n
}

func (e *BitStream) Pad() {
	pl := (8 - (e.Offset & 7)) & 7
	if pl > 0 {
		e.WriteBits(0, pl)
	}
}
