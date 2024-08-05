package pkg

import (
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
)

const (
	RocketPoolABI = `
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
     ]
     `
)

// RocketPoolOperation implements an implementation for generating calldata for staking and unstaking with Rocketpool
// It also implements dynamic retrival of Rocketpool's dynamic deposit and token contract addresses
type RocketPoolOperation struct {
	DynamicOperation
	// main deposit pool. this contract takes in the ETH
	contract *rocketpool.Contract
	// for unstaking
	rethContract *rocketpool.Contract
	// the deposit pool contract only checks for the maximum allowed amount
	// but the settings contract allows us check for the minimum amount of eth that
	// can be staked
	depositSettingsContract *rocketpool.Contract
	action                  ContractAction
	parsedABI               abi.ABI
}

// NewRocketPool initializes a RocketPool client
func NewRocketPool(rpcURL string) (*RocketPoolOperation, error) {

	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	rp, err := rocketpool.NewRocketPool(ethClient, RocketPoolStorageAddress)
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

	p := &RocketPoolOperation{
		DynamicOperation: DynamicOperation{
			Protocol: RocketPool,
			ChainID:  big.NewInt(1),
			Address:  RocketPoolStorageAddress,
		},
		contract:                contract,
		rethContract:            rethContract,
		depositSettingsContract: depositSettingsContract,
		parsedABI:               parsedABI,
	}

	return p, nil
}

// Register registers the RocketPoolOperation client into the protocol registry so it can be used by any user of
// the registry library
func (r *RocketPoolOperation) Register(registry *ProtocolRegistry) {
	// registry.RegisterProtocolOperation(r.Address, r.action, r.ChainID, r)
}

// GenerateCalldata dynamically generates the calldata for deposit and withdrawal actions
func (r *RocketPoolOperation) GenerateCalldata(op ContractAction, opts GenerateCalldataOptions) (string, error) {
	switch op {
	case NativeStake:
		return r.deposit(opts)
	case NativeUnStake:
		return r.withdraw(opts)

	default:
		return "", errors.New("unsupported action")
	}
}

func (r *RocketPoolOperation) withdraw(opts GenerateCalldataOptions) (string, error) {

	calldata, err := r.parsedABI.Pack("transfer", opts.Amount)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", r.Method, err)
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

func (r *RocketPoolOperation) deposit(opts GenerateCalldataOptions) (string, error) {

	amount := big.NewInt(0)

	if err := r.contract.Call(&bind.CallOpts{}, &amount, "getMaximumDepositAmount"); err != nil {
		return "", err
	}

	if val := amount.Cmp(opts.Amount); val == -1 {
		return "", errors.New("rocketpool not accepting this much eth deposit at this time")
	}

	amount = big.NewInt(0)

	if err := r.depositSettingsContract.Call(&bind.CallOpts{}, &amount, "getMinimumDeposit"); err != nil {
		return "", err
	}

	if val := amount.Cmp(opts.Amount); val == 1 {
		return "", errors.New("eth value too low to deposit to Rocketpool at this time")
	}

	calldata, err := r.parsedABI.Pack("deposit")
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", r.Method, err)
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

func (r *RocketPoolOperation) Validate(asset common.Address) error {
	if IsNativeToken(asset) {
		return nil
	}

	return fmt.Errorf("unsupported asset for rocket pool staking (%s)", asset.Hex())
}
