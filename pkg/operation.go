package pkg

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// DynamicOperation encapsulates information needed for any protocol operation.
type DynamicOperation struct {
	Protocol ProtocolName // Protocol name for logging or identification
	Method   ProtocolMethod
	ChainID  *big.Int // Target chain ID
	Address  common.Address
}

// ProtocolOperation defines a generic interface for protocol operations.
type ProtocolOperation interface {
	GenerateCalldata(args []interface{}) (string, error) // Generates the calldata based on the dynamic operation details

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

// GenericProtocolOperation provides a flexible implementation for generating calldata for any protocol operation.
type GenericProtocolOperation struct {
	DynamicOperation
}

// GetContractAddress fetch the contract address for this protocol
func (gpo *GenericProtocolOperation) GetContractAddress(
	_ context.Context) (common.Address, error) {
	return gpo.Address, nil
}

// Validate checks if the given asset is a valid one for this operation
func (gpo *GenericProtocolOperation) Validate(asset common.Address) error {

	protocols, ok := tokenSupportedMap[gpo.ChainID.Int64()]
	if !ok {
		return errors.New("unsupported chain for asset validation")
	}

	addrs, ok := protocols[gpo.Protocol]
	if !ok {
		return errors.New("unsupported protocol for asset validation")
	}

	if len(addrs) == 0 {
		if strings.EqualFold(strings.ToLower(asset.Hex()), nativeDenomAddress) {
			return nil
		}

		return fmt.Errorf("unsupported asset for %s ( %s )", gpo.Protocol, asset)
	}

	for _, addr := range addrs {
		if strings.EqualFold(strings.ToLower(asset.Hex()), strings.ToLower(addr)) {
			return nil
		}
	}

	return fmt.Errorf("unsupported asset for %s ( %s )", gpo.Protocol, asset)
}

// GenerateCalldata dynamically generates calldata for a contract method call based on the operation's ABI, method, and arguments.
func (gpo *GenericProtocolOperation) GenerateCalldata(args []interface{}) (string, error) {
	var protocol Protocol
	found := false
	for _, protocols := range staticProtocols {
		for _, p := range protocols {
			if p.Method == gpo.Method {
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
		return "", fmt.Errorf("protocol %s with method %s not found", gpo.Protocol, gpo.Method)
	}

	method, exists := protocol.ParsedABI.Methods[string(gpo.Method)]
	if !exists {
		return "", fmt.Errorf("method %s not found in ABI for %s", gpo.Method, gpo.Protocol)
	}

	if len(args) != len(method.Inputs) {
		return "", fmt.Errorf("incorrect number of arguments for %s: expected %d, got %d", gpo.Method, len(method.Inputs), len(args))
	}

	for i, input := range method.Inputs {
		if !MatchType(input.Type, args[i]) {
			return "", fmt.Errorf("type mismatch for argument %d: expected %s, got %T", i, input.Type.String(), args[i])
		}
	}

	calldata, err := protocol.ParsedABI.Pack(string(gpo.Method), args...)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", gpo.Method, err)
	}

	calldataHex := HexPrefix + hex.EncodeToString(calldata)
	return calldataHex, nil
}
