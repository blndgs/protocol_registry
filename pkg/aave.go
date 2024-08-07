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

// ENUM(aave,spark)
type AaveProtocolFork uint8

// AaveOperation is an implementation of calldata generation for Aave and it's other forks
type AaveOperation struct {
	parsedABI abi.ABI
	fork      AaveProtocolFork
}

// NewAaveOperation creates a new calldata generator for Aave3 and other forks
// of Aave3. Spark is the only supported fork at this time
func NewAaveOperation(fork AaveProtocolFork) (*AaveOperation, error) {
	if !fork.IsValid() {
		return nil, errors.New("unsupported aave protocol fork")
	}

	parsedABI, err := abi.JSON(strings.NewReader(aaveV3ABI))
	if err != nil {
		return nil, err
	}

	return &AaveOperation{
		parsedABI: parsedABI,
		fork:      fork,
	}, nil
}

// GenerateCalldata creates the required calldata based off the provided options
func (a *AaveOperation) GenerateCalldata(op ContractAction,
	opts GenerateCalldataOptions) (string, error) {

	var calldata []byte
	var err error

	switch op {
	case LoanSupply:

		referalCode, ok := opts.ReferalCode.(uint16)
		if !ok {
			return "", errors.New("referal code is not a uint16")
		}

		calldata, err = a.parsedABI.Pack("supply",
			opts.Asset, opts.Amount, opts.UBO(), referalCode)
		if err != nil {
			return "", err
		}

	case LoanWithdraw:

		calldata, err = a.parsedABI.Pack("withdraw",
			opts.Asset, opts.Amount, opts.UBO())
		if err != nil {
			return "", err
		}

	default:

		return "", errors.New("operation not supported")
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

// Validates makes sure the protocol supports the provided asset
func (a *AaveOperation) Validate(asset common.Address) error {

	protocols, ok := tokenSupportedMap[1]
	if !ok {
		return errors.New("unsupported chain for asset validation")
	}

	var protocol = AaveV3
	if a.fork == AaveProtocolForkSpark {
		protocol = SparkLend
	}

	addrs, ok := protocols[protocol]
	if !ok {
		return errors.New("unsupported protocol for asset validation")
	}

	if len(addrs) == 0 {
		if strings.EqualFold(strings.ToLower(asset.Hex()), nativeDenomAddress) {
			return nil
		}

		return fmt.Errorf("unsupported asset for %s ( %s )", AaveV3, asset)
	}

	for _, addr := range addrs {
		if strings.EqualFold(strings.ToLower(asset.Hex()), strings.ToLower(addr)) {
			return nil
		}
	}

	return fmt.Errorf("unsupported asset for %s ( %s )", AaveV3, asset)
}

func (a *AaveOperation) Name() string {
	if a.fork == AaveProtocolForkAave {
		return AaveV3
	}

	return SparkLend
}

// Name returns the human readable name for the protocol
func (a *AaveOperation) Register(registry *ProtocolRegistry) {
	addr := AaveV3ContractAddress
	if a.fork == AaveProtocolForkSpark {
		addr = SparkLendContractAddress
	}

	registry.RegisterProtocolOperation(addr, big.NewInt(1), a)
}
