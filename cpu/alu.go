package cpu

func (c *CPU) setAlu8OpFlags(result uint8, noOverflowResult uint16, subtract bool) {
	c.setFlag(FlagSign, ((result & (1 << 7)) != 0))
	c.setFlag(FlagZero, (result == 0))
	c.setFlag(FlagParityOverflow, (uint16(result) != noOverflowResult))
	c.setFlag(FlagSubtract, subtract)
	c.setFlag(FlagCarry, (noOverflowResult & (1 << 8) != 0))
}

func (c *CPU) Add8WithFlags(a uint8, b uint8) uint8 {
	result := a + b
	noOverflowResult := uint16(a) + uint16(b)
	c.setAlu8OpFlags(result, noOverflowResult, false)
	return result
}

func (c *CPU) Subtract8WithFlags(a uint8, b uint8) uint8 {
	result := a - b
	noOverflowResult := uint16(a) - uint16(b)
	c.setAlu8OpFlags(result, noOverflowResult, true)
	return result
}

func (c *CPU) DoALUOperation(op ALUOperationType, operand uint8) {
	var result uint8
	switch op {
	case ALUOperationAdd:
		c.Registers.A = c.Add8WithFlags(c.Registers.A, operand)
	case ALUOperationAdc:
		panic("cpu: unimplemented alu operation")
	case ALUOperationSub:
		c.Registers.A = c.Subtract8WithFlags(c.Registers.A, operand)
	case ALUOperationSbc:
		panic("cpu: unimplemented alu operation")
	case ALUOperationAnd:
		panic("cpu: unimplemented alu operation")
	case ALUOperationXor:
		result = c.Registers.A ^ operand
	case ALUOperationOr:
		result = c.Registers.A | operand
	case ALUOperationCp:
		c.Subtract8WithFlags(c.Registers.A, operand)
	default:
		panic("cpu: unknown operation type passed to DoALUOperation")
	}

	if op == ALUOperationOr || op == ALUOperationXor {
		c.setFlag(FlagCarry, false)
		c.setFlag(FlagZero, (result == 0))
		c.setFlag(FlagParityOverflow, calcParity(result))
		c.setFlag(FlagSign, (result & (1 << 7) != 0))
		c.setFlag(FlagSubtract, false)
		c.setFlag(FlagHalfCarry, false)
		c.Registers.A = result
	}
}