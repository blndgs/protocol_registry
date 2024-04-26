package pkg

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// DynamicOperation encapsulates information needed for any protocol operation.
type DynamicOperation struct {
	Protocol string         // Protocol name for logging or identification
	Action   ContractAction // The action to perform (e.g., supply, withdraw)
	Args     []interface{}  // Arguments for the method call
	ChainID  *big.Int       // Target chain ID
}

// ProtocolOperation defines a generic interface for protocol operations.
type ProtocolOperation interface {
	GenerateCalldata() (string, error) // Generates the calldata based on the dynamic operation details
}

// GenericProtocolOperation provides a flexible implementation for generating calldata for any protocol operation.
type GenericProtocolOperation struct {
	DynamicOperation
}

// GenerateCalldata dynamically generates calldata for a contract method call based on the operation's ABI, method, and arguments.
func (gpo *GenericProtocolOperation) GenerateCalldata() (string, error) {
	protocol, ok := SupportedProtocols[gpo.Protocol]
	if !ok {
		return "", fmt.Errorf("unsupported protocol: %s", gpo.Protocol)
	}

	parsedABI, err := abi.JSON(strings.NewReader(protocol.ABI))
	if err != nil {
		return "", fmt.Errorf("failed to parse ABI for %s: %w", gpo.Protocol, err)
	}

	calldata, err := parsedABI.Pack(string(gpo.Action), gpo.Args...)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", gpo.Action, err)
	}

	calldataHex := HexPrefix + hex.EncodeToString(calldata[:])
	return calldataHex, nil
}
