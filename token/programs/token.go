package programs

const (
	// Initializes a new mint and optionally deposits all the newly minted
	// tokens in an account.
	//
	// The `InitializeMint` instruction requires no signers and MUST be
	// included within the same Transaction as the system program's
	// `CreateAccount` instruction that creates the account being initialized.
	// Otherwise another party can acquire ownership of the uninitialized
	// account.
	Token_Instruction_InitializeMint uint8 = iota

	// Initializes a new account to hold tokens.  If this account is associated
	// with the native mint then the token balance of the initialized account
	// will be equal to the amount of SOL in the account. If this account is
	// associated with another mint, that mint must be initialized before this
	// command can succeed.
	//
	// The `InitializeAccount` instruction requires no signers and MUST be
	// included within the same Transaction as the system program's
	// `CreateAccount` instruction that creates the account being initialized.
	// Otherwise another party can acquire ownership of the uninitialized
	// account.
	Token_Instruction_InitializeAccount

	// Initializes a multisignature account with N provided signers.
	//
	// Multisignature accounts can used in place of any single owner/delegate
	// accounts in any token instruction that require an owner/delegate to be
	// present.  The variant field represents the number of signers (M)
	// required to validate this multisignature account.
	//
	// The `InitializeMultisig` instruction requires no signers and MUST be
	// included within the same Transaction as the system program's
	// `CreateAccount` instruction that creates the account being initialized.
	// Otherwise another party can acquire ownership of the uninitialized
	// account.
	Token_Instruction_InitializeMultisig

	// Transfers tokens from one account to another either directly or via a
	// delegate.  If this account is associated with the native mint then equal
	// amounts of SOL and Tokens will be transferred to the destination
	// account.
	Token_Instruction_Transfer

	// Approves a delegate.  A delegate is given the authority over tokens on
	// behalf of the source account's owner.
	Token_Instruction_Approve

	// Revokes the delegate's authority.
	Token_Instruction_Revoke

	// Sets a new authority of a mint or account.
	Token_Instruction_SetAuthority

	// Mints new tokens to an account.  The native mint does not support
	// minting.
	Token_Instruction_MintTo

	// Burns tokens by removing them from an account.  `Burn` does not support
	// accounts associated with the native mint, use `CloseAccount` instead.
	Token_Instruction_Burn

	// Close an account by transferring all its SOL to the destination account.
	// Non-native accounts may only be closed if its token amount is zero.
	Token_Instruction_CloseAccount

	// Freeze an Initialized account using the Mint's freeze_authority (if set).
	Token_Instruction_FreezeAccount

	// Thaw a Frozen account using the Mint's freeze_authority (if set).
	Token_Instruction_ThawAccount

	// Transfers tokens from one account to another either directly or via a
	// delegate.  If this account is associated with the native mint then equal
	// amounts of SOL and Tokens will be transferred to the destination
	// account.
	//
	// This instruction differs from Transfer in that the token mint and
	// decimals value is checked by the caller.  This may be useful when
	// creating transactions offline or within a hardware wallet.
	Token_Instruction_TransferChecked

	// Approves a delegate.  A delegate is given the authority over tokens on
	// behalf of the source account's owner.
	//
	// This instruction differs from Approve in that the token mint and
	// decimals value is checked by the caller.  This may be useful when
	// creating transactions offline or within a hardware wallet.
	Token_Instruction_ApproveChecked

	// Mints new tokens to an account.  The native mint does not support minting.
	//
	// This instruction differs from MintTo in that the decimals value is
	// checked by the caller.  This may be useful when creating transactions
	// offline or within a hardware wallet.
	Token_Instruction_MintToChecked

	// Burns tokens by removing them from an account.  `BurnChecked` does not
	// support accounts associated with the native mint, use `CloseAccount`
	// instead.
	//
	// This instruction differs from Burn in that the decimals value is checked
	// by the caller. This may be useful when creating transactions offline or
	// within a hardware wallet.
	Token_Instruction_BurnChecked

	// Like InitializeAccount, but the owner pubkey is passed via instruction data
	// rather than the accounts list. This variant may be preferable when using
	// Cross Program Invocation from an instruction that does not need the owner's
	// `AccountInfo` otherwise.
	Token_Instruction_InitializeAccount2

	// Given a wrapped / native token account (a token account containing SOL)
	// updates its amount field based on the account's underlying `lamports`.
	// This is useful if a non-wrapped SOL account uses `system_instruction::transfer`
	// to move lamports to a wrapped token account, and needs to have its token
	// `amount` field updated.
	Token_Instruction_SyncNative

	// Like InitializeAccount2, but does not require the Rent sysvar to be provided.
	Token_Instruction_InitializeAccount3

	// Like InitializeMultisig, but does not require the Rent sysvar to be provided.
	Token_Instruction_InitializeMultisig2

	// Like InitializeMint, but does not require the Rent sysvar to be provided.
	Token_Instruction_InitializeMint2
)

// InstructionIDToName returns the name of the instruction given its ID.
func TokenInstructionIDToName(id uint8) string {
	switch id {
	case Token_Instruction_InitializeMint:
		return "InitializeMint"
	case Token_Instruction_InitializeAccount:
		return "InitializeAccount"
	case Token_Instruction_InitializeMultisig:
		return "InitializeMultisig"
	case Token_Instruction_Transfer:
		return "Transfer"
	case Token_Instruction_Approve:
		return "Approve"
	case Token_Instruction_Revoke:
		return "Revoke"
	case Token_Instruction_SetAuthority:
		return "SetAuthority"
	case Token_Instruction_MintTo:
		return "MintTo"
	case Token_Instruction_Burn:
		return "Burn"
	case Token_Instruction_CloseAccount:
		return "CloseAccount"
	case Token_Instruction_FreezeAccount:
		return "FreezeAccount"
	case Token_Instruction_ThawAccount:
		return "ThawAccount"
	case Token_Instruction_TransferChecked:
		return "TransferChecked"
	case Token_Instruction_ApproveChecked:
		return "ApproveChecked"
	case Token_Instruction_MintToChecked:
		return "MintToChecked"
	case Token_Instruction_BurnChecked:
		return "BurnChecked"
	case Token_Instruction_InitializeAccount2:
		return "InitializeAccount2"
	case Token_Instruction_SyncNative:
		return "SyncNative"
	case Token_Instruction_InitializeAccount3:
		return "InitializeAccount3"
	case Token_Instruction_InitializeMultisig2:
		return "InitializeMultisig2"
	case Token_Instruction_InitializeMint2:
		return "InitializeMint2"
	default:
		return ""
	}
}
