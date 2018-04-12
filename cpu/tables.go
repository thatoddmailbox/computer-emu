package cpu

var DecodeTable_R = map[uint8]Register8bitType{
	0: RegisterB,
	1: RegisterC,
	2: RegisterD,
	3: RegisterE,
	4: RegisterH,
	5: RegisterL,
	6: RegisterIndirectHL,
	7: RegisterA,
}

var DecodeTable_RP = map[uint8]RegisterPairType{
	0: RegisterPairBC,
	1: RegisterPairDE,
	2: RegisterPairHL,
	3: RegisterPairSP,
}

var DecodeTable_RP2 = map[uint8]RegisterPairType{
	0: RegisterPairBC,
	1: RegisterPairDE,
	2: RegisterPairHL,
	3: RegisterPairAF,
}

var DecodeTable_CC = map[uint8]ConditionCodeType{
	0: ConditionCodeNZ,
	1: ConditionCodeZ,
	2: ConditionCodeNC,
	3: ConditionCodeC,
	4: ConditionCodePO,
	5: ConditionCodePE,
	6: ConditionCodeP,
	7: ConditionCodeM,
}

var DecodeTable_ALU = map[uint8]ALUOperationType{
	0: ALUOperationAdd,
	1: ALUOperationAdc,
	2: ALUOperationSub,
	3: ALUOperationSbc,
	4: ALUOperationAnd,
	5: ALUOperationXor,
	6: ALUOperationOr,
	7: ALUOperationCp,
}

var DecodeTable_ROT = map[uint8]ALUShiftOperationType{
	0: ALUShiftOperationRlc,
	1: ALUShiftOperationRrc,
	2: ALUShiftOperationRl,
	3: ALUShiftOperationRr,
	4: ALUShiftOperationSla,
	5: ALUShiftOperationSra,
	6: ALUShiftOperationSll,
	7: ALUShiftOperationSrl,
}