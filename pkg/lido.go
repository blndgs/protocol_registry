package pkg

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// LidoOperation is an implementation of calldata generation for Aave and it's other forks
type LidoOperation struct {
	parsedABI abi.ABI
}

// NewLidoOperation creates an implementation of Lido protocol for generating calldata
func NewLidoOperation() (*LidoOperation, error) {
	parsedABI, err := abi.JSON(strings.NewReader(lidoABI))
	if err != nil {
		return nil, err
	}

	return &LidoOperation{
		parsedABI: parsedABI,
	}, nil
}

// GenerateCalldata creates the required calldata based off the provided options
func (a *LidoOperation) GenerateCalldata(op ContractAction,
	opts GenerateCalldataOptions) (string, error) {

	var calldata []byte
	var err error

	switch op {
	case NativeStake:

		calldata, err = a.parsedABI.Pack("submit", opts.UBO())
		if err != nil {
			return "", err
		}

	default:

		return "", errors.New("operation not supported")
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

// Validates makes sure the protocol supports the provided asset
func (a *LidoOperation) Validate(asset common.Address) error {
	if !IsNativeToken(asset) {
		return fmt.Errorf("unsupported asset for Lido staking ( %s )", asset)
	}

	return nil
}

func (a *LidoOperation) Register(registry *ProtocolRegistry) {
	registry.RegisterProtocolOperation(LidoContractAddress, big.NewInt(1), a)
}

// Name returns the human readable name for the protocol
func (a *LidoOperation) Name() string { return Lido }
