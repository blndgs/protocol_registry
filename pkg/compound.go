package pkg

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type CompoundV3Operation struct {
	// current chain
	chainID int64
	// Compound has different deployments for each single market
	proxyContract string
	// assets that are supported in this pool
	supportedAssets []string
	// make sure to parse the abi only once
	parsedABI abi.ABI

	// current lending action
	action ContractAction
}

// dynamically registers all supported pools
func registerCompoundRegistry(registry *ProtocolRegistry) {
	// for chainID, v := range compoundSupportedAssets {
	// 	for poolAddr := range v {
	// 		for _, action := range []ContractAction{LoanSupply, LoanWithdraw} {
	// 			c, err := NewCompoundV3(big.NewInt(chainID), common.HexToAddress(poolAddr), action)
	// 			if err != nil {
	// 				panic(fmt.Sprintf("Failed to create compound client for %s", poolAddr))
	// 			}
	//
	// 			c.Register(registry)
	// 		}
	// 	}
	// }
}

// NewCompoundV3 creates a new compound v3 instance
func NewCompoundV3(chainID *big.Int,
	proxyContractAddress common.Address, action ContractAction) (*CompoundV3Operation, error) {

	supportedChain, ok := compoundSupportedAssets[chainID.Int64()]
	if !ok {
		return nil, errors.New("unsupported chain for Compound in Protocol registry")
	}

	supportedAssets, ok := supportedChain[strings.ToLower(proxyContractAddress.Hex())]
	if !ok {
		return nil, errors.New("unsupported Compound pool address")
	}

	if action != LoanSupply && action != LoanWithdraw {
		return nil, errors.New("unsupported action for Compound")
	}

	var abiDefinition = CompoundV3SupplyABI
	if action == LoanWithdraw {
		abiDefinition = CompoundV3WithdrawABI
	}

	parsedABI, err := abi.JSON(strings.NewReader(abiDefinition))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI for %s: %v", Compound, err)
	}

	return &CompoundV3Operation{
		proxyContract:   strings.ToLower(proxyContractAddress.Hex()),
		chainID:         chainID.Int64(),
		supportedAssets: supportedAssets,
		parsedABI:       parsedABI,
		action:          action,
	}, nil
}

// Register registers the CompoundV3Operation client into the protocol registry so it can be used by any user of
// the registry library
func (c *CompoundV3Operation) Register(registry *ProtocolRegistry) {
	registry.RegisterProtocolOperation(common.HexToAddress(c.proxyContract), c.action, big.NewInt(c.chainID), c)
}

// GetContractAddress retrieves the current lending market contract
func (c *CompoundV3Operation) GetContractAddress(_ context.Context) (common.Address, error) {
	return common.HexToAddress(c.proxyContract), nil
}

// Validate ensures the current asset can be supplied to the market
func (c *CompoundV3Operation) Validate(asset common.Address) error {

	for _, addr := range c.supportedAssets {
		if strings.EqualFold(strings.ToLower(asset.Hex()), strings.ToLower(addr)) {
			return nil
		}
	}

	return fmt.Errorf("unsupported asset for %s ( %s )", Compound, asset)
}

// GenerateCalldata creates a dynamic calldata that can be sent onchain
// to carry out lending operations
func (c *CompoundV3Operation) GenerateCalldata(operation ContractAction,
	opts GenerateCalldataOptions) (string, error) {

	if len(args) > 2 {
		return "", errors.New("provided args more than supported abi arguments. there can only be 2 args")
	}

	tokenAddress, ok := args[0].(common.Address)
	if !ok {
		return "", errors.New("argument 1 must be of type common.Address")
	}

	amount, ok := args[1].(*big.Int)
	if !ok {
		return "", errors.New("argument 2 must be of type *big.Int")
	}

	if c.action == LoanSupply {
		return c.supply(tokenAddress, amount)
	}

	return c.withdraw(tokenAddress, amount)
}

func (c *CompoundV3Operation) withdraw(tokenAddress common.Address, amount *big.Int) (string, error) {
	calldata, err := c.parsedABI.Pack("withdraw", tokenAddress, amount)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", "withdraw", err)
	}

	calldataHex := HexPrefix + hex.EncodeToString(calldata)
	return calldataHex, nil
}

func (c *CompoundV3Operation) supply(tokenAddress common.Address, amount *big.Int) (string, error) {

	calldata, err := c.parsedABI.Pack("supply", tokenAddress, amount)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", "supply", err)
	}

	calldataHex := HexPrefix + hex.EncodeToString(calldata)
	return calldataHex, nil
}
