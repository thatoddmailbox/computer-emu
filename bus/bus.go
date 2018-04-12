package bus

const (
	ROMSize = 8*1024
	RAMSize = 4*1024
)

type BusMemoryIODevice interface {
	IsMapped(address uint16) bool
	ReadByte(address uint16) uint8
	WriteByte(address uint16, data uint8)
}

type BusDataIODevice interface {
	IsMapped(address uint8) bool
	ReadByte(address uint8) uint8
	WriteByte(address uint8, data uint8)
}

type EmulatorBus struct {
	ROM [ROMSize]byte
	RAM [RAMSize]byte
	DataDevices []BusDataIODevice
}

func (b *EmulatorBus) ReadMemoryByte(address uint16) uint8 {
	if address < ROMSize {
		return b.ROM[address]
	}
	if address > 0xFFFF-RAMSize {
		return b.RAM[address&(RAMSize-1)]
	}
	panic("bus: read from unmapped memory")
}

func (b *EmulatorBus) WriteMemoryByte(address uint16, data uint8) {
	if address < ROMSize {
		panic("bus: write to read-only memory")
		return
	}
	if address > 0xFFFF-RAMSize {
		b.RAM[address&(RAMSize-1)] = data
		return
	}
	panic("bus: write to unmapped memory")
}

func (b *EmulatorBus) ReadIOByte(address uint8) uint8 {
	for _, device := range b.DataDevices {
		if device.IsMapped(address) {
			return device.ReadByte(address)
		}
	}
	panic("bus: read from unmapped IO")
}

func (b *EmulatorBus) WriteIOByte(address uint8, data uint8) {
	for _, device := range b.DataDevices {
		if device.IsMapped(address) {
			device.WriteByte(address, data)
			return
		}
	}
	panic("bus: write to unmapped IO")
}
