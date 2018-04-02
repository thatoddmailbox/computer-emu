package cpu

func (c *CPU) Get8bitRegister(reg Register8bitType) uint8 {
	switch reg {
	case RegisterA:
		return c.Registers.A
	case RegisterB:
		return c.Registers.B
	case RegisterC:
		return c.Registers.C
	case RegisterD:
		return c.Registers.D
	case RegisterE:
		return c.Registers.E
	case RegisterH:
		return c.Registers.H
	case RegisterL:
		return c.Registers.L
	case RegisterIndirectHL:
		return c.Bus.ReadMemoryByte(registerPair(c.Registers.H, c.Registers.L))
	default:
		panic("cpu: unknown register passed to Get8bitRegister")
	}
}

func (c *CPU) Set8bitRegister(reg Register8bitType, val uint8) {
	switch reg {
	case RegisterA:
		c.Registers.A = val
	case RegisterB:
		c.Registers.B = val
	case RegisterC:
		c.Registers.C = val
	case RegisterD:
		c.Registers.D = val
	case RegisterE:
		c.Registers.E = val
	case RegisterH:
		c.Registers.H = val
	case RegisterL:
		c.Registers.L = val
	case RegisterIndirectHL:
		c.Bus.WriteMemoryByte(registerPair(c.Registers.H, c.Registers.L), val)
	default:
		panic("cpu: unknown register passed to Set8bitRegister")
	}
}

func (c *CPU) Get16bitRegister(reg RegisterPairType) uint16 {
	switch reg {
	case RegisterPairAF:
		return (uint16(c.Registers.A) << 8) | uint16(c.Registers.Flag)
	case RegisterPairBC:
		return (uint16(c.Registers.B) << 8) | uint16(c.Registers.C)
	case RegisterPairDE:
		return (uint16(c.Registers.D) << 8) | uint16(c.Registers.E)
	case RegisterPairHL:
		return (uint16(c.Registers.H) << 8) | uint16(c.Registers.L)
	case RegisterPairSP:
		return c.Registers.SP
	default:
		panic("cpu: unknown register passed to Get16bitRegister")
	}
}

func (c *CPU) Set16bitRegister(reg RegisterPairType, val uint16) {
	high := uint8((val & 0xFF00) >> 8)
	low := uint8(val & 0xFF)
	switch reg {
	case RegisterPairAF:
		c.Registers.A = high
		c.Registers.Flag = low
	case RegisterPairBC:
		c.Registers.B = high
		c.Registers.C = low
	case RegisterPairDE:
		c.Registers.D = high
		c.Registers.E = low
	case RegisterPairHL:
		c.Registers.H = high
		c.Registers.L = low
	case RegisterPairSP:
		c.Registers.SP = val
	default:
		panic("cpu: unknown register passed to Set16bitRegister")
	}
}