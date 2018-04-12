package cpu

import (
	"strconv"
	"strings"
)

func Disassemble(sim CPU, pc uint16) (InstructionInfo, string) {
	instruction := sim.Bus.ReadMemoryByte(pc)

	info := DisassemblyTable_Unprefixed[instruction]
	formattedParams := info.Parameters
	if info.DataBytes == 1 {
		data := uint64(sim.Bus.ReadMemoryByte(pc + 1))
		formattedParams = strings.Replace(formattedParams, "%d8", "0x" + strconv.FormatUint(data, 16), -1)
	} else if info.DataBytes == 2 {
		data := uint64((uint16(sim.Bus.ReadMemoryByte(pc + 2)) << 8) | uint16(sim.Bus.ReadMemoryByte(pc + 1)))
		formattedParams = strings.Replace(formattedParams, "%d16", "0x" + strconv.FormatUint(data, 16), -1)
	}

	return info, formattedParams
}