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

func (c *CPU) Push(data uint8) {
	c.Registers.SP = c.Registers.SP - 1
	c.Bus.WriteMemoryByte(c.Registers.SP, data)
}

func (c *CPU) Pop() uint8 {
	data := c.Bus.ReadMemoryByte(c.Registers.SP)
	c.Registers.SP = c.Registers.SP + 1
	return data
}

func (c *CPU) Step() error {
	instruction := c.Bus.ReadMemoryByte(c.PC)
	instructionLength := 1

	prefix := uint16(0)

	if instruction == 0xCB {
		prefix = 0xCB
		instructionLength += 1
	} else if instruction == 0xED {
		prefix = 0xED
		instructionLength += 1
	} else if instruction == 0xDD {
		prefix = 0xDD
		instructionLength += 1
		if c.Bus.ReadMemoryByte(c.PC + 1) == 0xCB {
			prefix = 0xDDCB
			instructionLength += 1
		}
	} else if instruction == 0xFD {
		prefix = 0xFD
		instructionLength += 1
		if c.Bus.ReadMemoryByte(c.PC + 1) == 0xCB {
			prefix = 0xFDCB
			instructionLength += 1
		}
	}

	if instructionLength > 1 {
		instruction = c.Bus.ReadMemoryByte(c.PC + uint16(instructionLength) - 1)
	}
	
	x := (instruction & 0xC0) >> 6 // 0b11000000
	y := (instruction & 0x38) >> 3 // 0b00111000
	z := (instruction & 0x07)      // 0b00000111
	p := y >> 1
	q := y % 2

	validInstruction := false
	shouldIncrementPC := true

	if prefix == 0 {
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
					instructionLength += 2
				} else if q == 1 {
					// add hl, rp[p]
					validInstruction = true
					hl := c.Get16bitRegister(RegisterPairHL)
					operand := c.Get16bitRegister(DecodeTable_RP[p])
					result := hl + operand
					resultNoOverflow := uint32(hl) + uint32(operand)
					c.Set16bitRegister(RegisterPairHL, result)
					c.setFlag(FlagCarry, (resultNoOverflow & (1 << 16) != 0))
				}
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
						instructionLength += 2
					} else if p == 3 {
						// ld [nn], a
						validInstruction = true
						c.Bus.WriteMemoryByte(registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1)), c.Registers.A)
						instructionLength += 2
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
						instructionLength += 2
					} else if p == 3 {
						// ld a, [nn]
						validInstruction = true
						c.Registers.A = c.Bus.ReadMemoryByte(registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1)))
						instructionLength += 2
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
				// inc r[y]
				validInstruction = true
				orig := c.Get8bitRegister(DecodeTable_R[y])
				c.Set8bitRegister(DecodeTable_R[y], c.Add8WithFlags(orig, 1))
			} else if z == 5 {
				// dec r[y]
				validInstruction = true
				orig := c.Get8bitRegister(DecodeTable_R[y])
				c.Set8bitRegister(DecodeTable_R[y], c.Subtract8WithFlags(orig, 1))
			} else if z == 6 {
				// ld r[y], n
				validInstruction = true
				c.Set8bitRegister(DecodeTable_R[y], c.Bus.ReadMemoryByte(c.PC + 1))
				instructionLength += 1
			} else if z == 7 {
				if y == 0 {
					// rlca
					validInstruction = true
					msb := c.Registers.A & (1 << 7)
					c.setFlag(FlagCarry, msb != 0)
					c.Registers.A = c.Registers.A << 1
					if msb != 0 {
						c.Registers.A |= 1
					}
				} else if y == 1 {
					// TODO
				} else if y == 2 {
					// rla
					validInstruction = true
					msb := c.Registers.A & (1 << 7)
					c.Registers.A = c.Registers.A << 1
					if c.getFlag(FlagCarry) {
						c.Registers.A |= 1
					}
					c.setFlag(FlagCarry, msb != 0)
				} else if y == 3 {
					// rra
					validInstruction = true
					lsb := c.Registers.A & 1
					c.Registers.A = c.Registers.A >> 1
					if c.getFlag(FlagCarry) {
						c.Registers.A |= (1 << 7)
					}
					c.setFlag(FlagCarry, lsb != 0)
				} else if y == 4 {
					// TODO
				} else if y == 5 {
					// cpl
					validInstruction = true
					c.Registers.A = ^c.Registers.A
				} else if y == 6 {
					// scf
					validInstruction = true
					c.setFlag(FlagCarry, true)
					c.setFlag(FlagSubtract, false)
				} else if y == 7 {
					// ccf
					validInstruction = true
					c.setFlag(FlagCarry, !c.getFlag(FlagCarry))
					c.setFlag(FlagSubtract, false)
				}
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
			if z == 0 {
				// ret cc[y]
				validInstruction = true
				if c.ConditionMet(DecodeTable_CC[y]) {
					low := c.Pop()
					high := c.Pop()
					c.PC = (uint16(high) << 8) | uint16(low)
					shouldIncrementPC = false
				}
			} else if z == 1 {
				if q == 0 {
					// pop rp2[p]
					validInstruction = true
					low := c.Pop()
					high := c.Pop()
					value := (uint16(high) << 8) | uint16(low)
					c.Set16bitRegister(DecodeTable_RP2[p], value)
				} else if q == 1 {
					if p == 0 {
						// ret
						validInstruction = true
						low := c.Pop()
						high := c.Pop()
						c.PC = (uint16(high) << 8) | uint16(low)
						shouldIncrementPC = false
					} else if p == 1 {
						// TODO
					} else if p == 2 {
						// jp hl
						validInstruction = true
						c.PC = registerPair(c.Registers.H, c.Registers.L)
						shouldIncrementPC = false
					} else if p == 3 {
						// TODO
					}
				}
			} else if z == 2 {
				// jp cc[y], nn
				validInstruction = true
				if c.ConditionMet(DecodeTable_CC[y]) {
					c.PC = registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1))
					shouldIncrementPC = false
				}
				instructionLength += 2
			} else if z == 3 {
				if y == 0 {
					// jp nn
					validInstruction = true
					c.PC = registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1))
					shouldIncrementPC = false
					instructionLength += 2
				} else if y == 1 {
					// TODO
				} else if y == 2 {
					// out [n], a
					validInstruction = true
					c.Bus.WriteIOByte(c.Bus.ReadMemoryByte(c.PC + 1), c.Registers.A)
					instructionLength += 1
				} else if y == 3 {
					// in a, [n]
					validInstruction = true
					c.Registers.A = c.Bus.ReadIOByte(c.Bus.ReadMemoryByte(c.PC + 1))
					instructionLength += 1
				} else if y == 4 {
					// ex [sp], hl
					validInstruction = true
					swapHigh := c.Bus.ReadMemoryByte(c.Registers.SP + 1)
					swapLow := c.Bus.ReadMemoryByte(c.Registers.SP)
					c.Bus.WriteMemoryByte(c.Registers.SP + 1, c.Registers.H)
					c.Bus.WriteMemoryByte(c.Registers.SP, c.Registers.L)
					c.Registers.H = swapHigh
					c.Registers.L = swapLow
				} else if y == 5 {
					// ex de, hl
					validInstruction = true
					swapHigh := c.Registers.D
					swapLow := c.Registers.E
					c.Registers.D = c.Registers.H
					c.Registers.E = c.Registers.L
					c.Registers.H = swapHigh
					c.Registers.L = swapLow
				} else if y == 6 {
					// di
					validInstruction = true
					// interrupts are not used, so no-op
				} else if y == 7 {
					// ei
					validInstruction = true
					// interrupts are not used, so no-op
				}
			} else if z == 4 {
				// call cc[y], nn
				validInstruction = true
				
				if c.ConditionMet(DecodeTable_CC[y]) {
					returnAddress := c.PC + 3
					c.Push(uint8((returnAddress & 0xFF00) >> 8))
					c.Push(uint8(returnAddress & 0xFF))
					
					c.PC = registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1))
					shouldIncrementPC = false
				}

				instructionLength += 2
			} else if z == 5 {
				if q == 0 {
					// push rp2[p]
					validInstruction = true
					value := c.Get16bitRegister(DecodeTable_RP2[p])
					c.Push(uint8((value & 0xFF00) >> 8))
					c.Push(uint8(value & 0xFF))
				} else if q == 1 {
					if p == 0 {
						// call nn
						validInstruction = true
						
						returnAddress := c.PC + 3
						c.Push(uint8((returnAddress & 0xFF00) >> 8))
						c.Push(uint8(returnAddress & 0xFF))
						
						c.PC = registerPair(c.Bus.ReadMemoryByte(c.PC + 2), c.Bus.ReadMemoryByte(c.PC + 1))
						shouldIncrementPC = false

						instructionLength += 2
					}
					// if p != 0 and we're here, it's a prefix which should have been caught earlier
					// so just ignore it
				}
			} else if z == 6 {
				// alu[y] n
				validInstruction = true
				operation := DecodeTable_ALU[y]
				operand := c.Bus.ReadMemoryByte(c.PC + 1)
				c.DoALUOperation(operation, operand)
				instructionLength += 1
			} else if z == 7 {
				// rst y*8
				validInstruction = true
				returnAddress := c.PC + 1
				c.Push(uint8((returnAddress & 0xFF00) >> 8))
				c.Push(uint8(returnAddress & 0xFF))
				c.PC = uint16(y)*8
				shouldIncrementPC = false
			}
		}
	} else if prefix == 0xCB {
		if x == 0 {
			// rot[y] r[z]
			validInstruction = true
			operation := DecodeTable_ROT[y]
			operand := DecodeTable_R[z]
			value := c.Get8bitRegister(operand)
			result := c.DoALUShiftOperation(operation, value)
			c.Set8bitRegister(operand, result)
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