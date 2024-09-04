package pkg

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const compoundv3ABI = `
[
  {
	"name": "withdraw",
	"type": "function",
	"inputs": [
	  {
		"type": "address"
	  },
	  {
		"type": "uint256"
	  }
	]
  },
  {
	"name": "supply",
	"type": "function",
	"inputs": [
	  {
		"type": "address"
	  },
	  {
		"type": "uint256"
	  }
	]
  }
]`

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
			"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", // WETH
		},
	},
}

// dynamically registers all supported pools
func registerCompoundRegistry(registry ProtocolRegistry, client *ethclient.Client) error {
	for chainID, v := range compoundSupportedAssets {
		for poolAddr := range v {
			c, err := NewCompoundOperation(client, big.NewInt(chainID), common.HexToAddress(poolAddr))
			if err != nil {
				return err
			}

			if err := registry.RegisterProtocol(big.NewInt(chainID), common.HexToAddress(poolAddr), c); err != nil {
				return err
			}
		}
	}

	return nil
}

// CompoundOperation implements the Protocol interface for Ankr
type CompoundOperation struct {
	parsedABI abi.ABI
	contract  common.Address
	chainID   *big.Int
	version   string
	erc20ABI  abi.ABI
	// assets that are supported in this pool
	supportedAssets []string

	client *ethclient.Client
}

func NewCompoundOperation(client *ethclient.Client, chainID *big.Int,
	marketPool common.Address) (*CompoundOperation, error) {

	parsedABI, err := abi.JSON(strings.NewReader(compoundv3ABI))
	if err != nil {
		return nil, err
	}

	erc20ABI, err := abi.JSON(strings.NewReader(erc20BalanceOfABI))
	if err != nil {
		return nil, err
	}

	supportedChain, ok := compoundSupportedAssets[chainID.Int64()]
	if !ok {
		return nil, errors.New("unsupported chain for Compound in Protocol registry")
	}

	supportedAssets, ok := supportedChain[strings.ToLower(marketPool.Hex())]
	if !ok {
		return nil, errors.New("unsupported Compound pool address")
	}

	return &CompoundOperation{
		supportedAssets: supportedAssets,
		parsedABI:       parsedABI,
		contract:        marketPool,
		chainID:         chainID,
		version:         "3",
		client:          client,
		erc20ABI:        erc20ABI,
	}, nil
}

// GenerateCalldata creates the necessary blockchain transaction data
func (a *CompoundOperation) GenerateCalldata(ctx context.Context, chainID *big.Int,
	action ContractAction, params TransactionParams) (string, error) {
	if chainID.Int64() != 1 {
		return "", ErrChainUnsupported
	}

	switch action {
	case LoanSupply:
		return a.supply(params)
	case LoanWithdraw:
		return a.withdraw(params)
	default:
		return "", errors.New("unsupported operation")
	}
}

func (c *CompoundOperation) withdraw(opts TransactionParams) (string, error) {
	calldata, err := c.parsedABI.Pack("withdraw", opts.Asset, opts.Amount)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", "withdraw", err)
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

func (c *CompoundOperation) supply(opts TransactionParams) (string, error) {

	calldata, err := c.parsedABI.Pack("supply", opts.Asset, opts.Amount)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", "supply", err)
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

// Validate checks if the provided parameters are valid for the specified action
func (l *CompoundOperation) Validate(ctx context.Context,
	chainID *big.Int, action ContractAction, params TransactionParams) error {

	if chainID.Int64() != 1 {
		return ErrChainUnsupported
	}

	if !l.IsSupportedAsset(ctx, l.chainID, params.Asset) {
		return fmt.Errorf("asset not supported %s", params.Asset)
	}

	if action != LoanSupply && action != LoanWithdraw {
		return errors.New("action not supported")
	}

	if action == LoanSupply {
		return nil
	}

	if params.Amount.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("amount must be greater than zero")
	}

	asset := params.Asset

	if action == LoanWithdraw {
		asset = l.contract
	}

	balance, err := l.GetBalance(ctx, l.chainID, params.Sender, asset)
	if err != nil {
		return err
	}

	if balance.Cmp(params.Amount) == -1 {
		return errors.New("balance not enough")
	}

	return nil
}

// GetBalance retrieves the balance for a specified account and asset
func (l *CompoundOperation) GetBalance(ctx context.Context, chainID *big.Int, account,
	asset common.Address) (*big.Int, error) {

	if chainID.Int64() != 1 {
		return nil, ErrChainUnsupported
	}

	callData, err := l.erc20ABI.Pack("balanceOf", account)
	if err != nil {
		return nil, err
	}

	result, err := l.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &asset,
		Data: callData,
	}, nil)
	if err != nil {
		return nil, err
	}

	balance := new(big.Int)
	err = l.erc20ABI.UnpackIntoInterface(&balance, "balanceOf", result)
	return balance, err
}

// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain
func (c *CompoundOperation) GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error) {
	var addrs = make([]common.Address, 0, len(c.supportedAssets))

	for _, v := range c.supportedAssets {
		addrs = append(addrs, common.HexToAddress(v))
	}

	return addrs, nil
}

// IsSupportedAsset checks if the specified asset is supported on the given chain
func (c *CompoundOperation) IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool {
	if chainID.Int64() != 1 {
		return false
	}

	for _, addr := range c.supportedAssets {
		if strings.EqualFold(strings.ToLower(asset.Hex()), strings.ToLower(addr)) {
			return true
		}
	}

	return false
}

// GetProtocolConfig returns the protocol config for a specific chain
func (l *CompoundOperation) GetProtocolConfig(chainID *big.Int) ProtocolConfig {
	return ProtocolConfig{
		ChainID:  l.chainID,
		Contract: l.contract,
		ABI:      l.parsedABI,
		Type:     TypeStake,
	}
}

// GetABI returns the ABI of the protocol's contract
func (l *CompoundOperation) GetABI(chainID *big.Int) abi.ABI { return l.parsedABI }

// GetType returns the protocol type
func (l *CompoundOperation) GetType() ProtocolType { return TypeLoan }

// GetContractAddress returns the contract address for a specific chain
func (l *CompoundOperation) GetContractAddress(chainID *big.Int) common.Address { return l.contract }

// Name returns the human readable name for the protocol
func (l *CompoundOperation) GetName() string { return Compound }

// GetVersion returns the version of the protocol
func (l *CompoundOperation) GetVersion() string { return l.version }
