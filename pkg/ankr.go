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

const ankrABI = `
 [
   {
     "name": "stakeAndClaimAethC",
     "type": "function",
     "inputs": []
   },
   {
     "name": "unstakeAETH",
     "type": "function",
     "inputs": [
       {
         "internalType": "uint256",
         "name": "shares",
         "type": "uint256"
       }
     ]
   }
 ]`

var ankrEthER20Account = common.HexToAddress("0xE95A203B1a91a908F9B9CE46459d101078c2c3cb")

// AnkrOperation implements the Protocol interface for Ankr
type AnkrOperation struct {
	parsedABI abi.ABI
	contract  common.Address
	chainID   *big.Int
	version   string
	erc20ABI  abi.ABI

	client *ethclient.Client
}

func NewAnkrOperation(client *ethclient.Client, chainID *big.Int) (*AnkrOperation, error) {
	parsedABI, err := abi.JSON(strings.NewReader(ankrABI))
	if err != nil {
		return nil, err
	}

	erc20ABI, err := abi.JSON(strings.NewReader(erc20BalanceOfABI))
	if err != nil {
		return nil, err
	}

	return &AnkrOperation{
		parsedABI: parsedABI,
		contract:  AnkrContractAddress,
		chainID:   chainID,
		version:   "3",
		client:    client,
		erc20ABI:  erc20ABI,
	}, nil
}

// GenerateCalldata creates the necessary blockchain transaction data
func (a *AnkrOperation) GenerateCalldata(ctx context.Context, chainID *big.Int,
	action ContractAction, params TransactionParams) (string, error) {
	if chainID.Int64() != 1 {
		return "", ErrChainUnsupported
	}

	var calldata []byte
	var err error

	switch action {
	case NativeStake:

		calldata, err = a.parsedABI.Pack("stakeAndClaimAethC")
		if err != nil {
			return "", err
		}

	case NativeUnStake:

		calldata, err = a.parsedABI.Pack("unstakeAETH", params.Amount)
		if err != nil {
			return "", err
		}

	default:

		return "", errors.New("operation not supported")
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

// Validate checks if the provided parameters are valid for the specified action
func (l *AnkrOperation) Validate(ctx context.Context,
	chainID *big.Int, action ContractAction, params TransactionParams) error {

	if chainID.Int64() != 1 {
		return ErrChainUnsupported
	}

	if !l.IsSupportedAsset(ctx, l.chainID, params.Asset) {
		return fmt.Errorf("asset not supported %s", params.Asset)
	}

	var balance = new(big.Int)
	var err error

	switch action {
	case NativeUnStake:

		// only validate amount during withdrawal
		if params.Amount.Cmp(big.NewInt(0)) <= 0 {
			return errors.New("amount to unstake must be greater than zero")
		}

		_, balance, err = l.GetBalance(ctx, l.chainID, params.Sender, params.Asset)

	case NativeStake:

		balance, err = l.client.BalanceAt(ctx, params.Sender, nil)

	default:

		return errors.New("action not supported")

	}
	if err != nil {
		return err
	}

	if balance.Cmp(params.Amount) == -1 {
		return errors.New("your balance is not enough")
	}

	return nil
}

// GetBalance retrieves the balance for a specified account and asset
func (l *AnkrOperation) GetBalance(ctx context.Context, chainID *big.Int,
	account, _ common.Address) (common.Address, *big.Int, error) {

	var address common.Address

	if chainID.Int64() != 1 {
		return address, nil, ErrChainUnsupported
	}

	callData, err := l.erc20ABI.Pack("balanceOf", account)
	if err != nil {
		return address, nil, err
	}

	result, err := l.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &ankrEthER20Account,
		Data: callData,
	}, nil)
	if err != nil {
		return address, nil, err
	}

	balance := new(big.Int)
	err = l.erc20ABI.UnpackIntoInterface(&balance, "balanceOf", result)
	return ankrEthER20Account, balance, err
}

// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain
func (l *AnkrOperation) GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error) {
	return []common.Address{
		common.HexToAddress(nativeDenomAddress),
	}, nil
}

// IsSupportedAsset checks if the specified asset is supported on the given chain
func (l *AnkrOperation) IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool {
	if chainID.Int64() != 1 {
		return false
	}

	return IsNativeToken(asset) || asset.Hex() == ankrEthER20Account.Hex()
}

// GetProtocolConfig returns the protocol config for a specific chain
func (l *AnkrOperation) GetProtocolConfig(chainID *big.Int) ProtocolConfig {
	return ProtocolConfig{
		ChainID:  l.chainID,
		Contract: l.contract,
		ABI:      l.parsedABI,
		Type:     TypeStake,
	}
}

// GetABI returns the ABI of the protocol's contract
func (l *AnkrOperation) GetABI(chainID *big.Int) abi.ABI { return l.parsedABI }

// GetType returns the protocol type
func (l *AnkrOperation) GetType() ProtocolType { return TypeStake }

// GetContractAddress returns the contract address for a specific chain
func (l *AnkrOperation) GetContractAddress(chainID *big.Int) common.Address { return l.contract }

// Name returns the human readable name for the protocol
func (l *AnkrOperation) GetName() string { return Ankr }

// GetVersion returns the version of the protocol
func (l *AnkrOperation) GetVersion() string { return l.version }
