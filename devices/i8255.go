package devices

import (
	"errors"
)

// Port A: input
// 		bit 7: up
// 		bit 6: down
// 		bit 5: left
// 		bit 4: right
// 		bit 3: back
// 		bit 2: select
// Port B: unused
// Port C: unused

type I8255 struct {
	portA          uint8
	portB          uint8
	portC          uint8
	portAInput     bool
	portBInput     bool
	portCHighInput bool
	portCLowInput  bool
}

func NewI8255() *I8255 {
	return &I8255{
		portAInput:     true,
		portBInput:     true,
		portCHighInput: true,
		portCLowInput:  true,
	}
}

func (p *I8255) GetPortA() uint8 {
	return p.portA
}

func (p *I8255) SetPortA(port uint8) bool {
	if p.portAInput {
		p.portA = port
		return true
	}
	return false
}

func (p *I8255) SetPortB(port uint8) bool {
	if p.portBInput {
		p.portB = port
		return true
	}
	return false
}

func (p *I8255) IsMapped(address uint16) bool {
	if (address&(1<<15) == 0) && (address&(1<<14) != 0) && (address&(1<<13) == 0) && (address&(1<<12) != 0) {
		return true
	}
	return false
}

func (p *I8255) ReadByte(address uint16) uint8 {
	maskedAddress := address & 3
	if maskedAddress == 0 {
		return p.portA
	} else if maskedAddress == 1 {
		return p.portB
	} else if maskedAddress == 2 {
		return p.portC
	}
	return 0xFF // illegal condition
}

func (p *I8255) WriteByte(address uint16, data uint8) {
	maskedAddress := address & 3
	if maskedAddress == 0 {
		p.portA = data
	} else if maskedAddress == 1 {
		p.portB = data
	} else if maskedAddress == 2 {
		p.portC = data
	} else if maskedAddress == 3 {
		// control
		modeSetFlag := data & (1 << 7)

		if modeSetFlag == 0 {
			portAMode := data & ((1 << 6) | (1 << 5))
			portADirection := data & (1 << 4)
			portCHighDirection := data & (1 << 3)
			portBMode := data & (1 << 2)
			portBDirection := data & (1 << 1)
			portCLowDirection := data & 1

			if portAMode != 0 || portBMode != 0 {
				panic(errors.New("non-zero modes not implemented"))
			}

			p.portAInput = (portADirection == 1)
			p.portBInput = (portBDirection == 1)
			p.portCLowInput = (portCLowDirection == 1)
			p.portCHighInput = (portCHighDirection == 1)
		} else {
			selectedBit := data & ((1 << 3) | (1 << 2) | (1 << 1))
			bitValue := data & 1
			if bitValue == 1 {
				p.portC = p.portC | (1 << selectedBit)
			} else {
				p.portC = p.portC & ^(1 << selectedBit)
			}
		}
	}
}
