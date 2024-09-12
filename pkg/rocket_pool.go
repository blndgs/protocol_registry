package pkg

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
)

const RocketPoolABI = `
	[
       {
         "inputs": [],
         "name": "deposit",
         "outputs": [],
         "stateMutability": "payable",
         "type": "function"
       },
       {
         "inputs": [
           {
             "internalType": "address",
             "name": "recipient",
             "type": "address"
           },
           {
             "internalType": "uint256",
             "name": "amount",
             "type": "uint256"
           }
         ],
         "name": "transfer",
         "outputs": [
           {
             "internalType": "bool",
             "name": "",
             "type": "bool"
           }
         ],
         "stateMutability": "nonpayable",
         "type": "function"
       }
     ]`

// RocketpoolOperation implements the Protocol interface for Ankr
type RocketpoolOperation struct {
	parsedABI abi.ABI
	chainID   *big.Int
	version   string

	client *ethclient.Client

	// main deposit pool. this contract takes in the ETH
	contract *rocketpool.Contract
	// for unstaking
	rethContract *rocketpool.Contract
	// the deposit pool contract only checks for the maximum allowed amount
	// but the settings contract allows us check for the minimum amount of eth that
	// can be staked
	depositSettingsContract *rocketpool.Contract

	rp *rocketpool.RocketPool
}

func NewRocketpoolOperation(client *ethclient.Client, chainID *big.Int) (*RocketpoolOperation, error) {
	rp, err := rocketpool.NewRocketPool(client, RocketPoolStorageAddress)
	if err != nil {
		return nil, err
	}

	addr, err := rp.GetAddress("rocketDepositPool", &bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	if addr == nil {
		return nil, errors.New("could not fetch rocketpool address pool")
	}

	rethAddr, err := rp.GetAddress("rocketTokenRETH", &bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	if rethAddr == nil {
		return nil, errors.New("could not fetch rocketpool address pool")
	}

	contract, err := rp.MakeContract("rocketDepositPool", *addr, &bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	rethContract, err := rp.MakeContract("rocketTokenRETH", *rethAddr, &bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	settingsForDeposits, err := rp.GetAddress("rocketDAOProtocolSettingsDeposit", &bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	depositSettingsContract, err := rp.MakeContract("rocketDAOProtocolSettingsDeposit", *settingsForDeposits, &bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	parsedABI, err := abi.JSON(strings.NewReader(RocketPoolABI))
	if err != nil {
		return nil, err
	}

	return &RocketpoolOperation{
		client:                  client,
		chainID:                 big.NewInt(1),
		version:                 "1",
		contract:                contract,
		rethContract:            rethContract,
		depositSettingsContract: depositSettingsContract,
		parsedABI:               parsedABI,
		rp:                      rp,
	}, nil
}

// GenerateCalldata creates the necessary blockchain transaction data
func (a *RocketpoolOperation) GenerateCalldata(ctx context.Context, chainID *big.Int,
	action ContractAction, params TransactionParams) (string, error) {
	if chainID.Int64() != 1 {
		return "", ErrChainUnsupported
	}

	switch action {
	case NativeStake:
		return a.deposit(params)
	case NativeUnStake:
		return a.withdraw(params)

	default:
		return "", errors.New("unsupported action")
	}
}

func (r *RocketpoolOperation) withdraw(opts TransactionParams) (string, error) {

	calldata, err := r.parsedABI.Pack("transfer", opts.GetBeneficiaryOwner(), opts.Amount)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", "withdraw", err)
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

func (r *RocketpoolOperation) deposit(_ TransactionParams) (string, error) {

	calldata, err := r.parsedABI.Pack("deposit")
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", "deposit", err)
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

// Validate checks if the provided parameters are valid for the specified action
func (l *RocketpoolOperation) Validate(ctx context.Context,
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
	case NativeStake:

		amount := big.NewInt(0)

		if err := l.contract.Call(&bind.CallOpts{}, &amount, "getMaximumDepositAmount"); err != nil {
			return err
		}

		if val := amount.Cmp(params.Amount); val == -1 {
			return errors.New("rocketpool not accepting this much eth deposit at this time")
		}

		amount = big.NewInt(0)

		if err := l.depositSettingsContract.Call(&bind.CallOpts{}, &amount, "getMinimumDeposit"); err != nil {
			return err
		}

		if val := amount.Cmp(params.Amount); val == 1 {
			return errors.New("eth value too low to deposit to Rocketpool at this time")
		}

		balance, err = l.client.BalanceAt(ctx, params.Sender, nil)

	case NativeUnStake:

		// validate amount only during unstaking
		if params.Amount.Cmp(big.NewInt(0)) <= 0 {
			return errors.New("amount must be greater than zero")
		}

		_, balance, err = l.GetBalance(ctx, l.chainID, params.Sender, params.Asset)

	default:

		return errors.New("action not supported")
	}

	if err != nil {
		return err
	}

	if balance.Cmp(params.Amount) == -1 {
		return errors.New("balance not enough")
	}

	return nil
}

// GetBalance retrieves the balance for a specified account and asset
func (l *RocketpoolOperation) GetBalance(ctx context.Context,
	chainID *big.Int, account, _ common.Address) (common.Address, *big.Int, error) {

	var address common.Address

	if chainID.Int64() != 1 {
		return address, nil, ErrChainUnsupported
	}

	bal, err := tokens.GetRETHBalance(l.rp, account, nil)
	return *l.rethContract.Address, bal, err
}

// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain
func (l *RocketpoolOperation) GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error) {
	return []common.Address{
		common.HexToAddress(nativeDenomAddress),
	}, nil
}

// IsSupportedAsset checks if the specified asset is supported on the given chain
func (l *RocketpoolOperation) IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool {
	if chainID.Int64() != 1 {
		return false
	}

	// native token or Reth
	return IsNativeToken(asset) || asset.Hex() == common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393").Hex()
}

// GetProtocolConfig returns the protocol config for a specific chain
func (l *RocketpoolOperation) GetProtocolConfig(chainID *big.Int) ProtocolConfig {
	return ProtocolConfig{
		ChainID:  l.chainID,
		Contract: *l.contract.Address,
		ABI:      l.parsedABI,
		Type:     TypeStake,
	}
}

// GetABI returns the ABI of the protocol's contract
func (l *RocketpoolOperation) GetABI(chainID *big.Int) abi.ABI { return l.parsedABI }

// GetType returns the protocol type
func (l *RocketpoolOperation) GetType() ProtocolType { return TypeStake }

// GetContractAddress returns the contract address for a specific chain
func (l *RocketpoolOperation) GetContractAddress(chainID *big.Int) common.Address {
	return *l.contract.Address
}

// Name returns the human readable name for the protocol
func (l *RocketpoolOperation) GetName() string { return RocketPool }

// GetVersion returns the version of the protocol
func (l *RocketpoolOperation) GetVersion() string { return l.version }
