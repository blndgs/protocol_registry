package pkg

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// chainID -> Contract address -> ERC20s that can be used as collateral
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
		},
	},
}

type CompoundOperation struct {
	chainID int64

	// Compound has different deployments for each single market
	proxyContract string

	supportedAssets []string
}

func NewCompound(chainID *big.Int,
	proxyContractAddress common.Address) (*CompoundOperation, error) {

	supportedChain, ok := compoundSupportedAssets[chainID.Int64()]
	if !ok {
		return nil, errors.New("unsupported chain for Compound in Protocol registry")
	}

	supportedAssets, ok := supportedChain[strings.ToLower(proxyContractAddress.Hex())]
	if !ok {
		return nil, errors.New("unsupported Compound pool address")
	}

	return &CompoundOperation{
		proxyContract:   strings.ToLower(proxyContractAddress.Hex()),
		chainID:         chainID.Int64(),
		supportedAssets: supportedAssets,
	}, nil
}

func (c *CompoundOperation) GetContractAddress(_ context.Context) (common.Address, error) {
	return common.HexToAddress(c.proxyContract), nil
}

func (c *CompoundOperation) Validate(asset common.Address) error {

	for _, addr := range c.supportedAssets {
		if strings.EqualFold(strings.ToLower(asset.Hex()), strings.ToLower(addr)) {
			return nil
		}
	}

	return fmt.Errorf("unsupported asset for %s ( %s )", Compound, asset)
}
