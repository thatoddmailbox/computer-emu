package cpu

import (
	"strconv"
)

func calcParity(p uint8) bool {
	// shhh it works
	str := strconv.FormatUint(uint64(p), 2)
	count := 0
	for _, c := range str {
		if c == '1' {
			count++
		}
	}
	return (count % 2) == 0
}

func registerPair(h uint8, l uint8) uint16 {
	return ((uint16(h) << 8) | uint16(l))
}

func (c *CPU) getFlag(flag uint8) bool {
	return (c.Registers.Flag & flag) != 0
}

func (c *CPU) setFlag(flag uint8, val bool) {
	if val {
		c.Registers.Flag |= flag
	} else {
		c.Registers.Flag &= ^flag
	}
}