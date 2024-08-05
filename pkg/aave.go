package pkg

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// AaveOperation is an implementation of calldata generation for Aave and it's other forks
type AaveOperation struct {
	parsedABI abi.ABI
}

func NewAaveOperation() (*AaveOperation, error) {
	parsedABI, err := abi.JSON(strings.NewReader(aaveV3ABI))
	if err != nil {
		return nil, err
	}

	return &AaveOperation{
		parsedABI: parsedABI,
	}, nil
}

func (a *AaveOperation) GenerateCalldata(op ContractAction,
	opts GenerateCalldataOptions) (string, error) {

	var calldata []byte
	var err error

	switch op {
	case LoanSupply:

		calldata, err = a.parsedABI.Pack("supply",
			[]interface{}{opts.Sender, opts.Amount, opts.UBO(), uint16(0)})
		if err != nil {
			return "", err
		}

	case LoanWithdraw:

		calldata, err = a.parsedABI.Pack("withdraw",
			[]interface{}{opts.Sender, opts.Amount, opts.UBO()})
		if err != nil {
			return "", err
		}

	default:

		return "", errors.New("operation not supported")
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

func (a *AaveOperation) GetContractAddress(ctx context.Context) (
	common.Address, error) {
	return common.HexToAddress(""), nil
}

func (a *AaveOperation) Validate(asset common.Address) error {

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
