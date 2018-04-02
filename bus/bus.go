package bus

import (
	"fmt"
)

type BusIO interface {
	ReadByte(address uint16) uint8
	WriteByte(address uint16, data uint8)
}

type EmulatorBus struct {
	ROM [8 * 1024]byte
	RAM [4 * 1024]byte
}

func (b *EmulatorBus) ReadMemoryByte(address uint16) uint8 {
	if address < 8*1024 {
		return b.ROM[address]
	}
	if address > 0xFFFF-(4*1024) {
		return b.RAM[address&((4*1024)-1)]
	}
	panic("bus: read from unmapped memory")
}

func (b *EmulatorBus) WriteMemoryByte(address uint16, data uint8) {
	if address < 8*1024 {
		panic("bus: write to read-only memory")
	}
	if address > 0xFFFF-(4*1024) {
		b.RAM[address] = data
	}
	panic("bus: write to unmapped memory")
}

func (b *EmulatorBus) ReadIOByte(address uint8) uint8 {
	return 0
}

func (b *EmulatorBus) WriteIOByte(address uint8, data uint8) {
	fmt.Print(string(data))
}
