package pkg

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// DynamicOperation encapsulates information needed for any protocol operation.
type DynamicOperation struct {
	Protocol ProtocolName // Protocol name for logging or identification
	Method   ProtocolMethod
	ChainID  *big.Int // Target chain ID
	Address  common.Address
}

type GenerateCalldataOptions struct {
	Amount *big.Int
	Sender common.Address
	// Not used right now because the sender is also the recipient
	// but here for future proofing in case we decide to let people send minted
	// tokens from these protocols to another address instead of themselves
	// e.g With Aave, when you supply DAI, the ADAI can be programmed to be deposite
	// into another address
	Recipient   common.Address
	Asset       common.Address
	ReferalCode any
}

// Ultimate Beneficiary Owner of the token to be minted
// Defaults to the sender
func (g GenerateCalldataOptions) UBO() common.Address {
	if g.Recipient.Hex() == "0x0000000000000000000000000000000000000000" {
		return g.Sender
	}

	return g.Recipient
}

// ProtocolOperation defines a generic interface for protocol operations.
type ProtocolOperation interface {
	// Generates the calldata based on the dynamic operation details
	GenerateCalldata(ContractAction, GenerateCalldataOptions) (string, error)

	// Validate checks if the given asset is a valid one for this operation
	// This will not be automatically called by GenerateCalldata.
	// The client must call this to validate against the current known action type
	Validate(asset common.Address) error

	// Name returns the protcol name. This can be useful for debugging purposes
	Name() string
}
