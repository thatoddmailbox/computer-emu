package cpu

func (c *CPU) setAlu8OpFlags(result uint8, noOverflowResult int16, subtract bool) {
	c.setFlag(FlagSign, ((result & (1 << 7)) != 0))
	c.setFlag(FlagZero, (result == 0))
	c.setFlag(FlagParityOverflow, (int16(result) != noOverflowResult))
	c.setFlag(FlagSubtract, subtract)
	c.setFlag(FlagCarry, (noOverflowResult > 0xFF || noOverflowResult < 0))
}

func (c *CPU) setAlu16OpFlags(result uint16, noOverflowResult int32, subtract bool) {
	c.setFlag(FlagSign, ((result & (1 << 15)) != 0))
	c.setFlag(FlagZero, (result == 0))
	c.setFlag(FlagParityOverflow, (int32(result) != noOverflowResult))
	c.setFlag(FlagSubtract, subtract)
	c.setFlag(FlagCarry, (noOverflowResult > 0xFFFF || noOverflowResult < 0))
}

func (c *CPU) Add8WithFlags(a uint8, b uint8) uint8 {
	result := a + b
	noOverflowResult := int16(a) + int16(b)
	c.setAlu8OpFlags(result, noOverflowResult, false)
	return result
}

func (c *CPU) Subtract8WithFlags(a uint8, b uint8) uint8 {
	result := a - b
	noOverflowResult := int16(a) - int16(b)
	c.setAlu8OpFlags(result, noOverflowResult, true)
	return result
}

func (c *CPU) Subtract16WithFlags(a uint16, b uint16) uint16 {
	result := a - b
	noOverflowResult := int32(a) - int32(b)
	c.setAlu16OpFlags(result, noOverflowResult, true)
	return result
}

func (c *CPU) DoALUOperation(op ALUOperationType, operand uint8) {
	var result uint8
	switch op {
	case ALUOperationAdd:
		c.Registers.A = c.Add8WithFlags(c.Registers.A, operand)
	case ALUOperationAdc:
		carry := uint8(0)
		if c.getFlag(FlagCarry) {
			carry = 1
		}
		c.Registers.A = c.Add8WithFlags(c.Registers.A, operand+carry)
	case ALUOperationSub:
		c.Registers.A = c.Subtract8WithFlags(c.Registers.A, operand)
	case ALUOperationSbc:
		carry := uint8(0)
		if c.getFlag(FlagCarry) {
			carry = 1
		}
		c.Registers.A = c.Subtract8WithFlags(c.Registers.A, operand+carry)
	case ALUOperationAnd:
		result = c.Registers.A & operand
	case ALUOperationXor:
		result = c.Registers.A ^ operand
	case ALUOperationOr:
		result = c.Registers.A | operand
	case ALUOperationCp:
		c.Subtract8WithFlags(c.Registers.A, operand)
	default:
		panic("cpu: unknown operation type passed to DoALUOperation")
	}

	if op == ALUOperationAnd || op == ALUOperationXor || op == ALUOperationOr {
		c.setFlag(FlagCarry, false)
		c.setFlag(FlagZero, (result == 0))
		c.setFlag(FlagParityOverflow, calcParity(result))
		c.setFlag(FlagSign, (result&(1<<7) != 0))
		c.setFlag(FlagSubtract, false)
		c.setFlag(FlagHalfCarry, false)
		c.Registers.A = result
	}
}

func (c *CPU) DoALUShiftOperation(op ALUShiftOperationType, input uint8) uint8 {
	var result uint8
	switch op {
	case ALUShiftOperationSla:
		c.setFlag(FlagCarry, (input&(1<<7)) != 0)
		result = input << 1
	case ALUShiftOperationSra:
		c.setFlag(FlagCarry, (input&1) != 0)
		result = input >> 1
		result |= ((input & (1 << 6)) << 1)
	case ALUShiftOperationSrl:
		c.setFlag(FlagCarry, (input&1) != 0)
		result = input >> 1
	default:
		panic("cpu: unknown operation type passed to DoALUShiftOperation")
	}

	c.setFlag(FlagZero, (result == 0))
	c.setFlag(FlagParityOverflow, calcParity(result))
	c.setFlag(FlagSign, (result&(1<<7) != 0))
	c.setFlag(FlagSubtract, false)
	c.setFlag(FlagHalfCarry, false)

	return result
}
