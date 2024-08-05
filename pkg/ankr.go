package pkg

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// AnkrOperation is an implementation of calldata generation for Aave and it's other forks
type AnkrOperation struct {
	parsedABI abi.ABI
}

func NewAnkrOperation() (*AnkrOperation, error) {
	parsedABI, err := abi.JSON(strings.NewReader(ankrABI))
	if err != nil {
		return nil, err
	}

	return &AnkrOperation{
		parsedABI: parsedABI,
	}, nil
}

func (a *AnkrOperation) GenerateCalldata(op ContractAction,
	opts GenerateCalldataOptions) (string, error) {

	var calldata []byte
	var err error

	switch op {
	case NativeStake:

		calldata, err = a.parsedABI.Pack("stakeAndClaimAethC")
		if err != nil {
			return "", err
		}

	case NativeUnStake:

		calldata, err = a.parsedABI.Pack("unstakeAETH", opts.Amount)
		if err != nil {
			return "", err
		}

	default:

		return "", errors.New("operation not supported")
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

func (a *AnkrOperation) Validate(asset common.Address) error {

	protocols, ok := tokenSupportedMap[1]
	if !ok {
		return errors.New("unsupported chain for asset validation")
	}

	addrs, ok := protocols[AaveV3]
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
