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

const binanceStakedETHABI = `
[
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      },
      {
        "internalType": "address",
        "name": "referral",
        "type": "address"
      }
    ],
    "name": "deposit",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "account",
        "type": "address"
      }
    ],
    "name": "balanceOf",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]
	`

var bnbBinanceWrappedETHOperation = common.HexToAddress("0x2170Ed0880ac9A755fd29B2688956BD959F933F8")

// BinanceWrappedEthOperation implements the Protocol interface for Lido
type BinanceWrappedEthOperation struct {
	parsedABI abi.ABI
	contract  common.Address
	chainID   *big.Int
	version   string

	client *ethclient.Client
}

func NewBinanceWrappedEthOperation(client *ethclient.Client, chainID *big.Int) (*BinanceWrappedEthOperation, error) {
	if !IsBnb(chainID) {
		return nil, fmt.Errorf("unsupported chain ID (%d)", chainID.Int64())
	}

	chainIDfromClient, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	if !IsBnb(chainIDfromClient) {
		return nil, fmt.Errorf("unsupported chain ID (%d)", chainIDfromClient.Int64())
	}

	parsedABI, err := abi.JSON(strings.NewReader(binanceStakedETHABI))
	if err != nil {
		return nil, err
	}

	return &BinanceWrappedEthOperation{
		parsedABI: parsedABI,
		contract:  BinanceStakedETHBNBContractAddress,
		chainID:   chainID,
		version:   "2",
		client:    client,
	}, nil
}

// GenerateCalldata creates the necessary blockchain transaction data to stake
// your eth with Binance
func (l *BinanceWrappedEthOperation) GenerateCalldata(ctx context.Context, chainID *big.Int,
	action ContractAction, params TransactionParams) (string, error) {
	if !IsBnb(chainID) {
		return "", ErrChainUnsupported
	}

	var calldata []byte
	var err error

	switch action {
	case NativeStake:
		referralAddress, ok := params.ExtraData["referral_address"].(common.Address)
		if !ok {
			return "", errors.New("referral code is not a valid address")
		}

		calldata, err = l.parsedABI.Pack("deposit", params.Amount, referralAddress)
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("action not supported")
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

// Validate checks if the provided parameters are valid for the specified action
func (l *BinanceWrappedEthOperation) Validate(ctx context.Context,
	chainID *big.Int, action ContractAction, params TransactionParams) error {

	if !IsBnb(chainID) {
		return ErrChainUnsupported
	}

	if !l.IsSupportedAsset(ctx, l.chainID, params.Asset) {
		return fmt.Errorf("asset not supported %s", params.Asset)
	}

	if action != NativeStake {
		return ErrActionUnsupported
	}

	return nil
}

// GetBalance retrieves the balance for a specified account and asset
func (l *BinanceWrappedEthOperation) GetBalance(ctx context.Context,
	chainID *big.Int, account, _ common.Address) (common.Address, *big.Int, error) {

	var address common.Address
	if !IsBnb(chainID) {
		return address, nil, ErrChainUnsupported
	}

	callData, err := l.parsedABI.Pack("balanceOf", account)
	if err != nil {
		return address, nil, err
	}

	result, err := l.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &BinanceStakedETHBNBContractAddress,
		Data: callData,
	}, nil)
	if err != nil {
		return address, nil, err
	}

	balance := new(big.Int)
	err = l.parsedABI.UnpackIntoInterface(&balance, "balanceOf", result)
	return BinanceStakedETHBNBContractAddress, balance, err
}

// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain
func (l *BinanceWrappedEthOperation) GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error) {
	return []common.Address{bnbBinanceWrappedETHOperation}, nil
}

// IsSupportedAsset checks if the specified asset is supported on the given chain
func (l *BinanceWrappedEthOperation) IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool {
	if !IsBnb(chainID) {
		return false
	}

	return asset.Hex() == bnbBinanceWrappedETHOperation.Hex()
}

// GetProtocolConfig returns the protocol config for a specific chain
func (l *BinanceWrappedEthOperation) GetProtocolConfig(chainID *big.Int) ProtocolConfig {
	return ProtocolConfig{
		ChainID:  l.chainID,
		Contract: l.contract,
		ABI:      l.parsedABI,
		Type:     TypeStake,
	}
}

// GetABI returns the ABI of the protocol's contract
func (l *BinanceWrappedEthOperation) GetABI(chainID *big.Int) abi.ABI { return l.parsedABI }

// GetType returns the protocol type
func (l *BinanceWrappedEthOperation) GetType() ProtocolType { return TypeStake }

// GetContractAddress returns the contract address for a specific chain
func (l *BinanceWrappedEthOperation) GetContractAddress(chainID *big.Int) common.Address {
	return l.contract
}

// Name returns the human readable name for the protocol
func (l *BinanceWrappedEthOperation) GetName() string { return BinanceStakedETH }

// GetVersion returns the version of the protocol
func (l *BinanceWrappedEthOperation) GetVersion() string { return l.version }
