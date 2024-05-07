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

// IsValid checks if the provided arguments match the expected types from the ABI.
func (gpo *GenericProtocolOperation) IsValid() error {
	protocol, exists := SupportedProtocols[gpo.Protocol]
	if !exists {
		return fmt.Errorf("unsupported protocol: %s", gpo.Protocol)
	}
	parsedABI, err := abi.JSON(strings.NewReader(protocol.ABI))
	if err != nil {
		return err
	}

	method, exists := parsedABI.Methods[string(gpo.Action)]
	if !exists {
		return fmt.Errorf("method %s not found in ABI for %s", gpo.Action, gpo.Protocol)
	}

	if len(gpo.Args) != len(method.Inputs) {
		return fmt.Errorf("incorrect number of arguments: expected %d, got %d", len(method.Inputs), len(gpo.Args))
	}

	for i, input := range method.Inputs {
		if !MatchType(input.Type, gpo.Args[i]) {
			return fmt.Errorf("type mismatch for argument %d: expected %s, got %T", i, input.Type.String(), gpo.Args[i])
		}
	}

	return nil
}

// GenerateCalldata dynamically generates calldata for a contract method call based on the operation's ABI, method, and arguments.
func (gpo *GenericProtocolOperation) GenerateCalldata() (string, error) {
	if err := gpo.IsValid(); err != nil {
		return "", err
	}

	protocol, ok := SupportedProtocols[gpo.Protocol]
	if !ok {
		return "", fmt.Errorf("unsupported protocol: %s", gpo.Protocol)
	}

	method, exists := protocol.ParsedABI.Methods[string(gpo.Action)]
	if !exists {
		return "", fmt.Errorf("method %s not found in ABI for %s", gpo.Action, gpo.Protocol)
	}

	if len(gpo.Args) != len(method.Inputs) {
		return "", fmt.Errorf("incorrect number of arguments: expected %d, got %d", len(method.Inputs), len(gpo.Args))
	}

	typedArgs := make([]interface{}, len(gpo.Args))
	for i, input := range method.Inputs {
		arg := gpo.Args[i]
		if !MatchType(input.Type, arg) {
			return "", fmt.Errorf("type mismatch for argument %d: expected %s, got %T", i, input.Type.String(), arg)
		}
		typedArgs[i] = ConvertToABIType(input.Type, arg)
	}

	calldata, err := protocol.ParsedABI.Pack(string(gpo.Action), typedArgs...)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", gpo.Action, err)
	}

	calldataHex := HexPrefix + hex.EncodeToString(calldata[:])
	return calldataHex, nil
}
