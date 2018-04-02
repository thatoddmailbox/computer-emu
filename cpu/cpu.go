package cpu

import (
	"errors"
	"log"

	"github.com/thatoddmailbox/minemu/bus"
)

var ErrNotImplemented = errors.New("cpu: instruction not implemented")

type CPU struct {
	Bus bus.EmulatorBus
	Registers       RegisterFile
	ShadowRegisters RegisterFile
	PC              uint16
}

type RegisterFile struct {
	Flag uint8
	A    uint8
	B    uint8
	C    uint8
	D    uint8
	E    uint8
	H    uint8
	L    uint8
	SP   uint16
	IX   uint16
	IY   uint16
}

func (c *CPU) ConditionMet(condition ConditionCodeType) bool {
	switch condition {
	case ConditionCodeNZ:
		return !c.getFlag(FlagZero)
	case ConditionCodeZ:
		return c.getFlag(FlagZero)
	case ConditionCodeNC:
		return !c.getFlag(FlagCarry)
	case ConditionCodeC:
		return c.getFlag(FlagCarry)
	case ConditionCodePO:
		return !c.getFlag(FlagParityOverflow)
	case ConditionCodePE:
		return c.getFlag(FlagParityOverflow)
	case ConditionCodeP:
		return !c.getFlag(FlagSign)
	case ConditionCodeM:
		return c.getFlag(FlagSign)
	default:
		panic("unknown condition")
	}
}

func (c *CPU) Step() error {
	instruction := c.Bus.ReadMemoryByte(c.PC)
	instructionLength := 1
	
	x := (instruction & 0xC0) >> 6 // 0b11000000
	y := (instruction & 0x38) >> 3 // 0b00111000
	z := (instruction & 0x07)      // 0b00000111
	p := y >> 1
	q := y % 2

	validInstruction := false
	shouldIncrementPC := true
	if x == 0 {
		if z == 0 {
			if y == 0 {
				// nop
				validInstruction = true
			}
			// TODO: y: [1, 7]
		} else if z == 1 {
			if q == 0 {
				// ld rp[p], nn
				validInstruction = true
				c.Set16bitRegister(DecodeTable_RP[p], registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1)))
				instructionLength = 3
			}
			// TODO: q == 1
		} else if z == 2 {
			if q == 0 {
				if p == 0 {
					// ld [bc], a
					validInstruction = true
					c.Bus.WriteMemoryByte(registerPair(c.Registers.B, c.Registers.C), c.Registers.A)
				} else if p == 1 {
					// ld [de], a
					validInstruction = true
					c.Bus.WriteMemoryByte(registerPair(c.Registers.D, c.Registers.E), c.Registers.A)
				} else if p == 2 {
					// ld [nn], hl
					validInstruction = true
					c.Bus.WriteMemoryByte(registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1)), c.Registers.L)
					c.Bus.WriteMemoryByte(registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1)) + 1, c.Registers.H)
					instructionLength = 3
				} else if p == 3 {
					// ld [nn], a
					validInstruction = true
					c.Bus.WriteMemoryByte(registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1)), c.Registers.A)
					instructionLength = 3
				}
			} else if q == 1 {
				if p == 0 {
					// ld a, [bc]
					validInstruction = true
					c.Registers.A = c.Bus.ReadMemoryByte(registerPair(c.Registers.B, c.Registers.C))
				} else if p == 1 {
					// ld a, [de]
					validInstruction = true
					c.Registers.A = c.Bus.ReadMemoryByte(registerPair(c.Registers.D, c.Registers.E))
				} else if p == 2 {
					// ld hl, [nn]
					validInstruction = true
					c.Registers.L = c.Bus.ReadMemoryByte(registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1)))
					c.Registers.H = c.Bus.ReadMemoryByte(registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1)) + 1)
					instructionLength = 3
				} else if p == 3 {
					// ld a, [nn]
					validInstruction = true
					c.Registers.A = c.Bus.ReadMemoryByte(registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1)))
					instructionLength = 3
				}
			}
		} else if z == 3 {
			if q == 0 {
				// inc rp[p]
				validInstruction = true
				c.Set16bitRegister(DecodeTable_RP[p], c.Get16bitRegister(DecodeTable_RP[p]) + 1)
			} else if q == 1 {
				// dec rp[p]
				validInstruction = true
				c.Set16bitRegister(DecodeTable_RP[p], c.Get16bitRegister(DecodeTable_RP[p]) - 1)
			}
		} else if z == 4 {
			// TODO: everything
		} else if z == 5 {
			// TODO: everything
		} else if z == 6 {
			// ld r[y], n
			validInstruction = true
			c.Set8bitRegister(DecodeTable_R[y], c.Bus.ReadMemoryByte(c.PC + 1))
			instructionLength = 2
		}
	} else if x == 1 {
		if z == 6 && y == 6 {
			// halt
			// TODO
		} else {
			// ld r[y], r[z]
			validInstruction = true
			target := DecodeTable_R[y]
			source := DecodeTable_R[z]
			c.Set8bitRegister(target, c.Get8bitRegister(source))
		}
	} else if x == 2 {
		// alu[y] r[z]
		validInstruction = true
		operation := DecodeTable_ALU[y]
		operand := DecodeTable_R[z]
		c.DoALUOperation(operation, c.Get8bitRegister(operand))
	} else if x == 3 {
		if z == 1 {
			// TODO
		} else if z == 2 {
			// jp cc[y], nn
			validInstruction = true
			if c.ConditionMet(DecodeTable_CC[y]) {
				c.PC = registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1))
				shouldIncrementPC = false
			}
			instructionLength = 3
		} else if z == 3 {
			// TODO: y != 2
			if y == 0 {
				// jp nn
				validInstruction = true
				c.PC = registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1))
				shouldIncrementPC = false
				instructionLength = 3
			} else if y == 1 {
				// TODO
			} else if y == 2 {
				// out [n], a
				validInstruction = true
				c.Bus.WriteIOByte(c.Bus.ReadMemoryByte(c.PC + 1), c.Registers.A)
				instructionLength = 2
			} else if y == 3 {
				// TODO
			} else if y == 4 {
				// TODO
			} else if y == 5 {
				// TODO
			} else if y == 6 {
				// TODO
			} else if y == 7 {
				// TODO
			}
		} else if z == 4 {
			// TODO
		} else if z == 5 {
			// TODO
		} else if z == 6 {
			// alu[y] n
			validInstruction = true
			operation := DecodeTable_ALU[y]
			operand := c.Bus.ReadMemoryByte(c.PC + 1)
			c.DoALUOperation(operation, operand)
			instructionLength = 2
		} else if z == 7 {
			// TODO
		}
	}

	if shouldIncrementPC {
		c.PC += uint16(instructionLength)
	}

	if !validInstruction {
		log.Println("unimplemented instruction!")
		log.Printf("x: %d, y: %d, z: %d, p: %d, q: %d", x, y, z, p, q)
		return ErrNotImplemented
	}

	return nil
}