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
	"github.com/ethereum/go-ethereum/ethclient"
)

const listaABI = `
	[
       {
         "inputs": [],
         "name": "deposit",
         "outputs": [],
         "stateMutability": "payable",
         "type": "function"
       }
     ]`

// ListaStakingOperation implements staking, lending and supply for the lista dao project
// https://lista.org
type ListaStakingOperation struct {
	contract  common.Address
	parsedABI abi.ABI
	chainID   *big.Int
	client    *ethclient.Client
}

func NewListaStakingOperation(client *ethclient.Client,
	chainID *big.Int) (*ListaStakingOperation, error) {
	parsedABI, err := abi.JSON(strings.NewReader(listaABI))
	if err != nil {
		return nil, err
	}

	if chainID.Cmp(BscChainID) != 0 {
		return nil, ErrChainUnsupported
	}

	networkID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("client.NetworkID: could not fetch network id.. %w", err)
	}

	if networkID.Cmp(chainID) != 0 {
		return nil, fmt.Errorf("network id does not match")
	}

	return &ListaStakingOperation{
		parsedABI: parsedABI,
		chainID:   chainID,
		client:    client,
		contract:  ListaDaoContractAddress,
	}, nil
}

// GenerateCalldata creates the necessary blockchain transaction data
func (a *ListaStakingOperation) GenerateCalldata(ctx context.Context, chainID *big.Int,
	action ContractAction, params TransactionParams) (string, error) {
	if !a.isSupportedChain(chainID) {
		return "", ErrChainUnsupported
	}

	var calldata []byte
	var err error

	switch action {
	case NativeStake:
		calldata, err = a.parsedABI.Pack("deposit")
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("action not supported")
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

// Validate checks if the provided parameters are valid for the specified action
func (l *ListaStakingOperation) Validate(ctx context.Context,
	chainID *big.Int, action ContractAction, params TransactionParams) error {

	if !l.isSupportedChain(chainID) {
		return ErrChainUnsupported
	}

	if action != NativeStake {
		return errors.New("unsupported action")
	}

	balance, err := l.GetBalance(ctx, l.chainID, params.Sender, common.HexToAddress(nativeDenomAddress))
	if err != nil {
		return err
	}

	if params.Amount.Cmp(big.NewInt(0)) == 0 {
		return errors.New("please provide a non zero amount")
	}

	if balance.Cmp(params.Amount) == -1 {
		return errors.New("balance not enough")
	}

	return nil
}

// GetBalance retrieves the balance for a specified account and asset
func (l *ListaStakingOperation) GetBalance(ctx context.Context, chainID *big.Int,
	account, asset common.Address) (*big.Int, error) {

	if !l.isSupportedChain(chainID) {
		return nil, ErrChainUnsupported
	}

	return l.client.BalanceAt(ctx, account, nil)
}

// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain
func (l *ListaStakingOperation) GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error) {
	if !l.isSupportedChain(chainID) {
		return nil, ErrChainUnsupported
	}

	return []common.Address{
		common.HexToAddress(nativeDenomAddress),
	}, nil
}

func (l *ListaStakingOperation) isSupportedChain(chain *big.Int) bool {
	return l.chainID.Cmp(chain) == 0
}

// IsSupportedAsset checks if the specified asset is supported on the given chain
func (l *ListaStakingOperation) IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool {
	if !l.isSupportedChain(chainID) {
		return false
	}

	return IsNativeToken(asset)
}

// GetProtocolConfig returns the protocol config for a specific chain
func (l *ListaStakingOperation) GetProtocolConfig(chainID *big.Int) ProtocolConfig {
	return ProtocolConfig{
		ChainID:  l.chainID,
		ABI:      l.parsedABI,
		Type:     TypeStake,
		Contract: l.contract,
	}
}

// GetABI returns the ABI of the protocol's contract
func (l *ListaStakingOperation) GetABI(chainID *big.Int) abi.ABI { return l.parsedABI }

// GetType returns the protocol type
func (l *ListaStakingOperation) GetType() ProtocolType { return TypeStake }

// GetContractAddress returns the contract address for a specific chain
func (l *ListaStakingOperation) GetContractAddress(chainID *big.Int) common.Address {
	return l.contract
}

// Name returns the human readable name for the protocol
func (l *ListaStakingOperation) GetName() string { return ListaDao }

// GetVersion returns the version of the protocol
func (l *ListaStakingOperation) GetVersion() string { return "1" }
