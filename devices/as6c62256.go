package devices

type AS6C62256 struct {
	RAM [4 * 1024]byte
}

func NewAS6C62256() *AS6C62256 {
	return &AS6C62256{}
}

func (r *AS6C62256) IsMapped(address uint16) bool {
	if (address & 0xF000) == 0x3000 {
		return true
	} else {
		return false
	}
}

func (r *AS6C62256) ReadByte(address uint16) uint8 {
	accessAddress := address & 0x0FFF
	return r.RAM[accessAddress]
}

func (r *AS6C62256) WriteByte(address uint16, data uint8) {
	accessAddress := address & 0x0FFF
	r.RAM[accessAddress] = data
}
