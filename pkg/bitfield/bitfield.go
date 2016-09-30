package bitfield

type Bitfield struct {
    Bits []byte
}

func (b *Bitfield) IsSet(index int) bool {
    return b.Bits[index>>3] & byte(128>>byte(index&7)) != 0
}

func (b *Bitfield) Set(index int) {
    b.Bits[index>>3] |= byte(128>>byte(index&7))
}
