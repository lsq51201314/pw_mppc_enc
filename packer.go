package main

import "bytes"

type Packer struct {
	BlockSize int
	History   []byte
	Hash      []byte
	HistOff   int
	LegacyIn  int
}

func (e *Packer) Init() {
	e.BlockSize = 0x2000
	e.History = make([]byte, e.BlockSize)
	e.Hash = make([]byte, e.BlockSize)
	e.HistOff = 0
	e.LegacyIn = 0
}

func (e *Packer) Compress(input []byte) []byte {
	var ret []byte
	if len(input) > 0 || e.LegacyIn > 0 {
		ret = make([]byte, (9*(e.LegacyIn+len(input))>>3)+6)
		bits := e.Update(input, ret)
		e.CompressBlock(bits, e.LegacyIn)
		ret = bytes.TrimRight(ret, "\x00")
	} else {
		ret = input
	}
	return ret
}

func (e *Packer) Update(input, output []byte) *BitStream {
	remain := e.BlockSize - e.HistOff - e.LegacyIn
	isize := len(input)
	ioffset := 0
	ostream := new(BitStream)
	ostream.Init(output)
	if isize >= remain {
		_ = BlockCopy(input, 0, e.History, e.HistOff+e.LegacyIn, remain)
		isize -= remain
		ioffset += remain
		e.CompressBlock(ostream, remain+e.LegacyIn)
		e.HistOff = 0
		for isize >= e.BlockSize {
			_ = BlockCopy(input, ioffset, e.History, 0, e.BlockSize)
			e.CompressBlock(ostream, e.BlockSize)
			e.HistOff = 0
			isize -= e.BlockSize
			ioffset += e.BlockSize
		}
	}
	_ = BlockCopy(input, ioffset, e.History, e.HistOff+e.LegacyIn, isize)
	e.LegacyIn += isize
	return ostream
}

func (e *Packer) CompressBlock(stream *BitStream, size int) {
	r := e.HistOff + size
	s := e.HistOff
	for r-s > 2 {
		p := e.GetPredictAddr(s)
		if p < s {
			x := e.History[p] == e.History[s]
			p++
			s++
			y := e.History[p] == e.History[s]
			p++
			if x && y {
				s++
				z := e.History[p] == e.History[s]
				p++
				if z {
					s++
					for s < r && e.History[p] == e.History[s] {
						s++
						p++
					}
					l := s - e.HistOff
					e.HistOff = s
					e.PutOff(stream, s-p)
					val := 0
					n := 1
					if l > 3 {
						high := e.GetHighestBit(l)
						val = l&((1<<high)-1) | ((1<<(high-1))-1)<<(high+1)
						n = high << 1
					}
					e.PutBits(stream, val, n)
				} else {
					e.PutLit(stream, int(e.History[e.HistOff]))
					e.HistOff++
					s = e.HistOff
				}
			} else {
				e.PutLit(stream, int(e.History[e.HistOff]))
				e.HistOff++
			}
		} else {
			e.PutLit(stream, int(e.History[e.HistOff]))
			e.HistOff++
			s = e.HistOff
		}
	}
	if r-s == 1 || r-s == 2 {
		for i := 0; i < r-s; i++ {
			e.PutLit(stream, int(e.History[e.HistOff]))
			e.HistOff++
		}
	}
	e.PutOff(stream, 0)
	stream.Pad()
	e.LegacyIn = 0
}

func (e *Packer) PutBits(bits *BitStream, val, n int) {
	bits.WriteBits(val, n)
}

func (e *Packer) PutLit(bits *BitStream, c int) {
	i := ((c & 0xFFFF) | ((c & 0x80) << 1)) & 0x17F
	n := 8 | ((c & 0x80) >> 7)
	e.PutBits(bits, i, n)
}

func (e *Packer) PutOff(bits *BitStream, off int) {
	o := off | 0x3C0
	n := 0x0A
	if off > 0x3F {
		if off > 0x13F {
			o = (off - 0x140) | 0xC000
			n = 0x10
		} else {
			o = (off - 0x40) | 0xE00
			n = 0x0C
		}
	}
	e.PutBits(bits, o, n)
}

func (e *Packer) GetPredictAddr(offset int) int {
	i := (0x9E5F * (int(e.History[offset+2]) ^ 16*(int(e.History[offset+1])^16*int(e.History[offset]))) >> 4) & 0x1FFF
	ret := e.Hash[i]
	e.Hash[i] = byte(offset)
	return int(ret)
}

func (e *Packer) GetHighestBit(val int) int {
	ret := 31
	for (val & (1 << ret)) == 0 {
		ret--
	}
	return ret
}
