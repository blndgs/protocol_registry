package pkg

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// DynamicOperation encapsulates information needed for any protocol operation.
type DynamicOperation struct {
	Protocol ProtocolName   // Protocol name for logging or identification
	Action   ContractAction // The action to perform (e.g., supply, withdraw)
	ChainID  *big.Int       // Target chain ID
	Address  common.Address
}

// ProtocolOperation defines a generic interface for protocol operations.
type ProtocolOperation interface {
	GenerateCalldata(kind AssetKind, args []interface{}) (string, error) // Generates the calldata based on the dynamic operation details

	// retrieves the address for the contract interaction.
	// Sometimes this might be static but some protocols do not use a static address
	// like Rocketpool and others. The current deposit pool address would need to be dynamically
	// retrieved
	GetContractAddress(ctx context.Context) (common.Address, error)
}

// GenericProtocolOperation provides a flexible implementation for generating calldata for any protocol operation.
type GenericProtocolOperation struct {
	DynamicOperation
}

// GetContractAddress fetch the contract address for this protocol
func (gpo *GenericProtocolOperation) GetContractAddress(
	_ context.Context) (common.Address, error) {
	return gpo.Address, nil
}

// GenerateCalldata dynamically generates calldata for a contract method call based on the operation's ABI, method, and arguments.
func (gpo *GenericProtocolOperation) GenerateCalldata(kind AssetKind, args []interface{}) (string, error) {
	var protocol Protocol
	found := false
	for _, protocols := range SupportedProtocols {
		for _, p := range protocols {
			if p.Name == gpo.Protocol && p.Action == gpo.Action {
				protocol = p
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return "", fmt.Errorf("protocol %s with action %s not found", gpo.Protocol, gpo.Action)
	}

	method, exists := protocol.ParsedABI.Methods[string(gpo.Action)]
	if !exists {
		return "", fmt.Errorf("method %s not found in ABI for %s", gpo.Action, gpo.Protocol)
	}

	if len(args) != len(method.Inputs) {
		return "", fmt.Errorf("incorrect number of arguments for %s: expected %d, got %d", gpo.Action, len(method.Inputs), len(args))
	}

	for i, input := range method.Inputs {
		if !MatchType(input.Type, args[i]) {
			return "", fmt.Errorf("type mismatch for argument %d: expected %s, got %T", i, input.Type.String(), args[i])
		}
	}

	calldata, err := protocol.ParsedABI.Pack(string(gpo.Action), args...)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", gpo.Action, err)
	}

	calldataHex := HexPrefix + hex.EncodeToString(calldata)
	return calldataHex, nil
}
