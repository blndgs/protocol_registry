package pkg

import (
	"context"
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
	Recipient   common.Address
	ReferalCode any
}

// Ultimate Beneficiary Owner of the token to be minted
func (g GenerateCalldataOptions) UBO() common.Address {
	if len(g.Recipient.Bytes()) == 0 {
		return g.Sender
	}

	return g.Recipient
}

// ProtocolOperation defines a generic interface for protocol operations.
type ProtocolOperation interface {
	// Generates the calldata based on the dynamic operation details
	GenerateCalldata(ContractAction, GenerateCalldataOptions) (string, error)

	// retrieves the address for the contract interaction.
	// Sometimes this might be static but some protocols do not use a static address
	// like Rocketpool and others. The current deposit pool address would need to be dynamically
	// retrieved
	GetContractAddress(ctx context.Context) (common.Address, error)

	// Validate checks if the given asset is a valid one for this operation
	// This will not be automatically called by GenerateCalldata.
	// The client must call this to validate against the current known action type
	Validate(asset common.Address) error
}
