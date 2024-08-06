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

type CompoundV3Operation struct {
	// current chain
	chainID int64
	// assets that are supported in this pool
	supportedAssets []string
	// make sure to parse the abi only once
	parsedABI abi.ABI
}

// dynamically registers all supported pools
func registerCompoundRegistry(registry *ProtocolRegistry) {
	for chainID, v := range compoundSupportedAssets {
		for poolAddr := range v {
			c, err := NewCompoundV3(big.NewInt(chainID), common.HexToAddress(poolAddr))
			if err != nil {
				panic(fmt.Sprintf("Failed to create compound client for %s", poolAddr))
			}

			c.Register(registry, common.HexToAddress(poolAddr))
		}
	}
}

// chainID -> Contract address of pool market -> ERC20s that can be used as collateral
// Compound has different markets and each market only supports a
// few assets as collateral
var compoundSupportedAssets = map[int64]map[string][]string{
	// Ethereum
	1: {
		// USDC pool
		"0xc3d688b66703497daa19211eedff47f25384cdc3": []string{
			nativeDenomAddress,                           // ETH
			"0x514910771AF9Ca656af840dff83E8264EcF986CA", // LINK
			"0xc00e94Cb662C3520282E6f5717214004A7f26888", // COMP
			"0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984", // UNI
			"0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", // WBTC
		},
		// ETH pool
		"0xa17581a9e3356d9a858b789d68b4d866e593ae94": []string{
			"0xBe9895146f7AF43049ca1c1AE358B0541Ea49704", // cbETH
			"0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0", // wsETH (Lido)
			"0xae78736Cd615f374D3085123A210448E74Fc6393", // RocketPool ETH
			"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", // WETH
		},
	},
}

// NewCompoundV3 creates a new compound v3 instance
func NewCompoundV3(chainID *big.Int, marketContractAddress common.Address) (*CompoundV3Operation, error) {

	supportedChain, ok := compoundSupportedAssets[chainID.Int64()]
	if !ok {
		return nil, errors.New("unsupported chain for Compound in Protocol registry")
	}

	supportedAssets, ok := supportedChain[strings.ToLower(marketContractAddress.Hex())]
	if !ok {
		return nil, fmt.Errorf("unsupported Compound market...%s", marketContractAddress)
	}

	parsedABI, err := abi.JSON(strings.NewReader(compoundv3ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI for %s: %v", Compound, err)
	}

	return &CompoundV3Operation{
		chainID:         chainID.Int64(),
		supportedAssets: supportedAssets,
		parsedABI:       parsedABI,
	}, nil
}

// Register registers the CompoundV3Operation client into the protocol registry so it can be used by any user of
// the registry library
func (c *CompoundV3Operation) Register(registry *ProtocolRegistry, addr common.Address) {
	registry.RegisterProtocolOperation(addr, big.NewInt(c.chainID), c)
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
func (c *CompoundV3Operation) GenerateCalldata(op ContractAction,
	opts GenerateCalldataOptions) (string, error) {

	switch op {
	case LoanSupply:
		return c.supply(opts)
	case LoanWithdraw:
		return c.withdraw(opts)
	default:
		return "", errors.New("unsupported operation")
	}
}

func (c *CompoundV3Operation) withdraw(opts GenerateCalldataOptions) (string, error) {
	calldata, err := c.parsedABI.Pack("withdraw", opts.Asset, opts.Amount)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", "withdraw", err)
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

func (c *CompoundV3Operation) supply(opts GenerateCalldataOptions) (string, error) {

	calldata, err := c.parsedABI.Pack("supply", opts.Asset, opts.Amount)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", "supply", err)
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}
