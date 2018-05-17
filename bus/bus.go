package bus

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
	MemoryDevices []BusMemoryIODevice
	DataDevices   []BusDataIODevice
}

func (b *EmulatorBus) ReadMemoryByte(address uint16) uint8 {
	for _, device := range b.MemoryDevices {
		if device.IsMapped(address) {
			return device.ReadByte(address)
		}
	}
	panic("bus: read from unmapped memory")
}

func (b *EmulatorBus) WriteMemoryByte(address uint16, data uint8) {
	for _, device := range b.MemoryDevices {
		if device.IsMapped(address) {
			device.WriteByte(address, data)
			return
		}
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
