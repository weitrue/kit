package programs

const (
	Token2022_Instruction_Transfer = 3
	// This instruction differs from Transfer in that the token mint and
	// decimals value is checked by the caller.  This may be useful when
	// creating transactions offline or within a hardware wallet.
	Token2022_Instruction_TransferChecked = 12
)

// InstructionIDToName returns the name of the instruction given its ID.
func Token2022InstructionIDToName(id uint8) string {
	switch id {
	case Token2022_Instruction_Transfer:
		return "Transfer"
	case Token2022_Instruction_TransferChecked:
		return "TransferChecked"
	default:
		return ""
	}
}
