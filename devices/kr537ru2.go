package devices

type KR537RU2 struct {
	RAM [4 * 1024]byte
}

func NewKR537RU2() *KR537RU2 {
	return &KR537RU2{}
}

func (r *KR537RU2) IsMapped(address uint16) bool {
	if (address & 0xF000) == 0xF000 {
		return true
	} else {
		return false
	}
}

func (r *KR537RU2) ReadByte(address uint16) uint8 {
	accessAddress := address & 0x0FFF
	return r.RAM[accessAddress]
}

func (r *KR537RU2) WriteByte(address uint16, data uint8) {
	accessAddress := address & 0x0FFF
	r.RAM[accessAddress] = data
}
