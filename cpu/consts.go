package cpu

type ALUOperationType int
type ConditionCodeType int
type Register8bitType int
type RegisterPairType int

const (
	ALUOperationAdd ALUOperationType = iota
	ALUOperationAdc
	ALUOperationSub
	ALUOperationSbc
	ALUOperationAnd
	ALUOperationXor
	ALUOperationOr
	ALUOperationCp
)

const (
	ConditionCodeNZ ConditionCodeType = iota
	ConditionCodeZ
	ConditionCodeNC
	ConditionCodeC
	ConditionCodePO
	ConditionCodePE
	ConditionCodeP
	ConditionCodeM
)

const (
	RegisterA Register8bitType = iota
	RegisterB
	RegisterC
	RegisterD
	RegisterE
	RegisterH
	RegisterL
	RegisterIndirectHL
)

const (
	RegisterPairAF RegisterPairType = iota
	RegisterPairBC
	RegisterPairDE
	RegisterPairHL
	RegisterPairSP
)

const (
	FlagSign = (1 << 7)
	FlagZero = (1 << 6)
	FlagHalfCarry = (1 << 4)
	FlagParityOverflow = (1 << 2)
	FlagSubtract = (1 << 1)
	FlagCarry = (1 << 0)
)