package bitfield

type BitField []byte

func (b BitField) HasPiece(index int) bool {
	byteIndex := index / 8
	bitIndex := index % 8
	return b[byteIndex]>>(7-bitIndex)&1 != 0
}

func (b BitField) SetPiece(index int) {
	byteIndex := index / 8
	bitIndex := index % 8
	b[byteIndex] |= 1 << (7 - bitIndex)
}
