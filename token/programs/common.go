package programs

import "github.com/gagliardetto/solana-go"

func MapProgramToID(programID solana.PublicKey) uint8 {
	switch programID {
	case solana.SystemProgramID:
		return 1 // system
	case solana.TokenProgramID:
		return 2 // token
	case solana.Token2022ProgramID:
		return 3 // token2022
	default:
		return 0 // invalid
	}
}
