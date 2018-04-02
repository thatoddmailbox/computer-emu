package bus

import (
	"fmt"
)

const (
	ROMSize = 16*1024
	RAMSize = 48*1024
)

type BusIO interface {
	ReadByte(address uint16) uint8
	WriteByte(address uint16, data uint8)
}

type EmulatorBus struct {
	ROM [ROMSize]byte
	RAM [RAMSize]byte
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
		//panic("bus: write to read-only memory")
		b.ROM[address] = data
		return
	}
	if address > 0xFFFF-RAMSize {
		b.RAM[address&(RAMSize-1)] = data
		return
	}
	panic("bus: write to unmapped memory")
}

func (b *EmulatorBus) ReadIOByte(address uint8) uint8 {
	return 0
}

func (b *EmulatorBus) WriteIOByte(address uint8, data uint8) {
	fmt.Print(string(data))
}
