package cpu

import (
	"strconv"
	"strings"
)

func DisassembleInstructionAt(sim CPU, pc uint16) (InstructionInfo, string, uint8) {
	instruction := sim.Bus.ReadMemoryByte(pc)
	table := DisassemblyTable_Unprefixed
	instructionOffset := 0

	if instruction == 0xCB {
		table = DisassemblyTable_CB
		instructionOffset = 1
	} else if instruction == 0xDD {
		if sim.Bus.ReadMemoryByte(pc + 1) == 0xCB {
			table = DisassemblyTable_DDCB
			instructionOffset = 2
		} else {
			table = DisassemblyTable_DD
			instructionOffset = 1
		}
	} else if instruction == 0xED {
		table = DisassemblyTable_ED
		instructionOffset = 1
	} else if instruction == 0xFD {
		if sim.Bus.ReadMemoryByte(pc + 1) == 0xCB {
			table = DisassemblyTable_FDCB
			instructionOffset = 2
		} else {
			table = DisassemblyTable_FD
			instructionOffset = 1
		}
	}

	if instructionOffset > 0 {
		instruction = sim.Bus.ReadMemoryByte(pc + uint16(instructionOffset))
	}

	info := table[instruction]
	formattedParams := info.Parameters
	if info.DataBytes == 1 {
		data := uint64(sim.Bus.ReadMemoryByte(pc + 1))
		formattedParams = strings.Replace(formattedParams, "%d8", "0x" + strconv.FormatUint(data, 16), -1)
	} else if info.DataBytes == 2 {
		data := uint64((uint16(sim.Bus.ReadMemoryByte(pc + 2)) << 8) | uint16(sim.Bus.ReadMemoryByte(pc + 1)))
		formattedParams = strings.Replace(formattedParams, "%d16", "0x" + strconv.FormatUint(data, 16), -1)
	}

	return info, formattedParams, (uint8(instructionOffset) + 1 + info.DataBytes)
}