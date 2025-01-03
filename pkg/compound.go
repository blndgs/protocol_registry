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
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "",
        "type": "address"
      }
    ],
    "name": "userCollateral",
    "outputs": [
      {
        "internalType": "uint128",
        "name": "balance",
        "type": "uint128"
      },
      {
        "internalType": "uint128",
        "name": "_reserved",
        "type": "uint128"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "numAssets",
    "outputs": [
      {
        "internalType": "uint8",
        "name": "",
        "type": "uint8"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint8",
        "name": "",
        "type": "uint8"
      }
    ],
    "name": "getAssetInfo",
    "outputs": [
      {
        "internalType": "uint8",
        "name": "offset",
        "type": "uint8"
      },
      {
        "internalType": "address",
        "name": "asset",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "priceFeed",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "scale",
        "type": "uint64"
      },
      {
        "internalType": "uint64",
        "name": "borrowCollateralFactor",
        "type": "uint64"
      },
      {
        "internalType": "uint64",
        "name": "liquidateCollateralFactor",
        "type": "uint64"
      },
      {
        "internalType": "uint64",
        "name": "liquidationFactor",
        "type": "uint64"
      },
      {
        "internalType": "uint128",
        "name": "supplyCap",
        "type": "uint128"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]
`

const (
	CompoundV3USDCPool        = "0xc3d688b66703497daa19211eedff47f25384cdc3"
	CompoundV3ETHPool         = "0xa17581a9e3356d9a858b789d68b4d866e593ae94"
	CompoundV3PolygonUSDCPool = "0xF25212E676D1F7F89Cd72fFEe66158f541246445"
	CompoundV3PolygonUSDTPool = "0xaeB318360f27748Acb200CE616E389A6C9409a07"
)

var poolMaps = map[int64][]string{
	1:   {CompoundV3ETHPool, CompoundV3USDCPool},
	137: {CompoundV3PolygonUSDCPool, CompoundV3PolygonUSDTPool},
}

// dynamically registers all supported pools
func registerCompoundRegistry(registry ProtocolRegistry, client *ethclient.Client, chainID int64) error {
	protocols, ok := poolMaps[chainID]
	if !ok {
		return nil
	}

	for _, poolAddr := range protocols {
		c, err := NewCompoundOperation(client, big.NewInt(chainID), common.HexToAddress(poolAddr))
		if err != nil {
			return err
		}

		if err := registry.RegisterProtocol(big.NewInt(chainID), common.HexToAddress(poolAddr), c); err != nil {
			return err
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

	// assets that are supported in this pool
	supportedAssets []common.Address

	client *ethclient.Client
}

func NewCompoundOperation(client *ethclient.Client, chainID *big.Int,
	marketPool common.Address) (*CompoundOperation, error) {

	parsedABI, err := abi.JSON(strings.NewReader(compoundv3ABI))
	if err != nil {
		return nil, err
	}

	supportedAssets, err := getSupportedAssets(parsedABI, client, marketPool)
	if err != nil {
		return nil, err
	}

	if chainID.Int64() != EthChainID.Int64() && chainID.Int64() != PolygonChainID.Int64() {
		return nil, errors.New("unsupported chain id")
	}

	return &CompoundOperation{
		supportedAssets: supportedAssets,
		parsedABI:       parsedABI,
		contract:        marketPool,
		chainID:         chainID,
		version:         "3",
		client:          client,
	}, nil
}

func getSupportedAssets(parsedPoolABI abi.ABI,
	client *ethclient.Client, marketPool common.Address) ([]common.Address, error) {

	numAssetsCallData, err := parsedPoolABI.Pack("numAssets")
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		To:   &marketPool,
		Data: numAssetsCallData,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, err
	}

	var numAssets uint8

	err = parsedPoolABI.UnpackIntoInterface(&numAssets, "numAssets", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack output: %v", err)
	}

	var supportedTokens = make([]common.Address, 0, numAssets)

	// Fetch info for each collateral asset
	for i := uint8(0); i < numAssets; i++ {
		var assetInfo struct {
			Offset                    uint8
			Asset                     common.Address
			PriceFeed                 common.Address
			Scale                     uint64
			BorrowCollateralFactor    uint64
			LiquidateCollateralFactor uint64
			LiquidationFactor         uint64
			SupplyCap                 *big.Int
		}

		assetInfoCalldata, err := parsedPoolABI.Pack("getAssetInfo", i)
		if err != nil {
			return nil, err
		}

		msg := ethereum.CallMsg{
			To:   &marketPool,
			Data: assetInfoCalldata,
		}

		result, err := client.CallContract(context.Background(), msg, nil)
		if err != nil {
			return nil, err
		}

		err = parsedPoolABI.UnpackIntoInterface(&assetInfo, "getAssetInfo", result)
		if err != nil {
			return nil, fmt.Errorf("failed to unpack output: %v", err)
		}

		supportedTokens = append(supportedTokens, assetInfo.Asset)
	}

	return supportedTokens, nil
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

	return nil
}

// GetBalance retrieves the balance for a specified account and asset
func (l *CompoundOperation) GetBalance(ctx context.Context,
	chainID *big.Int,
	account, asset common.Address) (common.Address, *big.Int, error) {

	var address common.Address

	if !l.IsSupportedAsset(ctx, chainID, asset) {
		return address, nil, errors.New("unsupported asset. cannot fetch it's balance")
	}

	if chainID.Int64() != 1 {
		return address, nil, ErrChainUnsupported
	}

	callData, err := l.parsedABI.Pack("userCollateral", account, asset)
	if err != nil {
		return address, nil, err
	}

	result, err := l.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &l.contract,
		Data: callData,
	}, nil)
	if err != nil {
		return address, nil, err
	}

	balance := new(big.Int)
	res := new(big.Int)
	err = l.parsedABI.UnpackIntoInterface(&[]interface{}{&balance, &res}, "userCollateral", result)
	return l.contract, balance, err
}

// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain
func (c *CompoundOperation) GetSupportedAssets(ctx context.Context,
	chainID *big.Int) ([]common.Address, error) {
	return c.supportedAssets, nil
}

// IsSupportedAsset checks if the specified asset is supported on the given chain
func (c *CompoundOperation) IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool {
	if chainID.Int64() != 1 {
		return false
	}

	for _, addr := range c.supportedAssets {
		if addr.Hex() == asset.Hex() {
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
