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

// lidoABI is the ABI definition for the Lido protocol
const lidoABI = `
[
  {
    "constant": false,
    "inputs": [
      {
        "name": "_referral",
        "type": "address"
      }
    ],
    "name": "submit",
    "outputs": [
      {
        "name": "",
        "type": "uint256"
      }
    ],
    "payable": true,
    "stateMutability": "payable",
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
]`

// LidoOperation implements the Protocol interface for Lido
type LidoOperation struct {
	parsedABI abi.ABI
	contract  common.Address
	chainID   *big.Int
	version   string

	client *ethclient.Client
}

func NewLidoOperation(client *ethclient.Client, chainID *big.Int) (*LidoOperation, error) {
	parsedABI, err := abi.JSON(strings.NewReader(lidoABI))
	if err != nil {
		return nil, err
	}

	return &LidoOperation{
		parsedABI: parsedABI,
		contract:  LidoContractAddress,
		chainID:   chainID,
		version:   "3",
		client:    client,
	}, nil
}

// GenerateCalldata creates the necessary blockchain transaction data
func (l *LidoOperation) GenerateCalldata(ctx context.Context, chainID *big.Int,
	action ContractAction, params TransactionParams) (string, error) {
	if chainID.Int64() != 1 {
		return "", ErrChainUnsupported
	}

	var calldata []byte
	var err error

	switch action {
	case NativeStake:
		// TODO: change this to Balloondogs referral
		calldata, err = l.parsedABI.Pack("submit", params.GetBeneficiaryOwner())
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("action not supported")
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

// Validate checks if the provided parameters are valid for the specified action
func (l *LidoOperation) Validate(ctx context.Context,
	chainID *big.Int, action ContractAction, params TransactionParams) error {

	if chainID.Int64() != 1 {
		return ErrChainUnsupported
	}

	if !l.IsSupportedAsset(ctx, l.chainID, params.Asset) {
		return fmt.Errorf("asset not supported %s", params.Asset)
	}

	if action != NativeStake {
		return errors.New("action not supported")
	}

	asset := nativeDenomAddress
	if action == NativeUnStake {
		// will default to fetching the balance from the contract
		// not implemented right now as it is recommended for holders to
		// swap their stEth on DEXs or CEXs instead of waiting 3-10 days for lido withdrawal
		asset = ""

		// validate amount only during unstaking
		if params.Amount.Cmp(big.NewInt(0)) <= 0 {
			return errors.New("amount must be greater than zero")
		}
	}

	balance, err := l.GetBalance(ctx, l.chainID, params.Sender, common.HexToAddress(asset))
	if err != nil {
		return err
	}

	if balance.Cmp(params.Amount) == -1 {
		return errors.New("balance not enough")
	}

	return nil
}

// GetBalance retrieves the balance for a specified account and asset
func (l *LidoOperation) GetBalance(ctx context.Context, chainID *big.Int, account, asset common.Address) (*big.Int, error) {
	if chainID.Int64() != 1 {
		return nil, ErrChainUnsupported
	}

	if strings.ToLower(asset.Hex()) == nativeDenomAddress {
		return l.client.BalanceAt(ctx, account, nil)
	}

	callData, err := l.parsedABI.Pack("balanceOf", account)
	if err != nil {
		return nil, err
	}

	result, err := l.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &LidoContractAddress,
		Data: callData,
	}, nil)
	if err != nil {
		return nil, err
	}

	balance := new(big.Int)
	err = l.parsedABI.UnpackIntoInterface(&balance, "balanceOf", result)
	return balance, err
}

// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain
func (l *LidoOperation) GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error) {
	return []common.Address{
		common.HexToAddress(nativeDenomAddress),
	}, nil
}

// IsSupportedAsset checks if the specified asset is supported on the given chain
func (l *LidoOperation) IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool {
	if chainID.Int64() != 1 {
		return false
	}

	return IsNativeToken(asset)
}

// GetProtocolConfig returns the protocol config for a specific chain
func (l *LidoOperation) GetProtocolConfig(chainID *big.Int) ProtocolConfig {
	return ProtocolConfig{
		ChainID:  l.chainID,
		Contract: l.contract,
		ABI:      l.parsedABI,
		Type:     TypeStake,
	}
}

// GetABI returns the ABI of the protocol's contract
func (l *LidoOperation) GetABI(chainID *big.Int) abi.ABI { return l.parsedABI }

// GetType returns the protocol type
func (l *LidoOperation) GetType() ProtocolType { return TypeStake }

// GetContractAddress returns the contract address for a specific chain
func (l *LidoOperation) GetContractAddress(chainID *big.Int) common.Address { return l.contract }

// Name returns the human readable name for the protocol
func (l *LidoOperation) GetName() string { return Lido }

// GetVersion returns the version of the protocol
func (l *LidoOperation) GetVersion() string { return l.version }
