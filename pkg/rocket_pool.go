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

// GenerateCalldata dynamically generates the calldata for deposit and withdrawal actions
func (r *RocketPoolOperation) GenerateCalldata(args []interface{}) (string, error) {
	switch r.Method {
	case rocketPoolStake:
		return r.deposit(args)
	case rocketPoolUnStake:
		return r.withdraw(args)
	}
	return "", errors.New("unsupported action")
}

// Register registers the RocketPoolOperation client into the protocol registry so it can be used by any user of
// the registry library
func (r *RocketPoolOperation) Register(registry *ProtocolRegistry) {
	registry.RegisterProtocolOperation(r.Address, r.action, r.ChainID, r)
}

// NewRocketPool initializes a RocketPool client
func NewRocketPool(rpcURL string, contractAddress ContractAddress, action ContractAction, method ProtocolMethod) (*RocketPoolOperation, error) {
	if method != rocketPoolStake && method != rocketPoolUnStake {
		return nil, errors.New("unsupported action")
	}

	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	rp, err := rocketpool.NewRocketPool(ethClient, contractAddress)
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
			Method:   method,
			ChainID:  big.NewInt(1),
		},
		contract:                contract,
		rethContract:            rethContract,
		depositSettingsContract: depositSettingsContract,
		action:                  action,
		parsedABI:               parsedABI,
	}

	return p, nil
}

func (r *RocketPoolOperation) withdraw(args []interface{}) (string, error) {

	_, exists := r.parsedABI.Methods["transfer"]
	if !exists {
		return "", errors.New("unsupported action")
	}

	calldata, err := r.parsedABI.Pack("transfer", args...)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", r.Method, err)
	}

	calldataHex := HexPrefix + hex.EncodeToString(calldata)
	return calldataHex, nil
}

func (r *RocketPoolOperation) deposit(args []interface{}) (string, error) {

	_, exists := r.parsedABI.Methods["deposit"]
	if !exists {
		return "", errors.New("unsupported action")
	}

	amount := big.NewInt(0)

	if err := r.contract.Call(&bind.CallOpts{}, &amount, "getMaximumDepositAmount"); err != nil {
		return "", err
	}

	amountToDeposit, ok := args[0].(*big.Int)
	if !ok {
		return "", errors.New("arg is not of type *big.Int")
	}

	if val := amount.Cmp(amountToDeposit); val == -1 {
		return "", errors.New("rocketpool not accepting this much eth deposit at this time")
	}

	amount = big.NewInt(0)

	if err := r.depositSettingsContract.Call(&bind.CallOpts{}, &amount, "getMinimumDeposit"); err != nil {
		return "", err
	}

	if val := amount.Cmp(amountToDeposit); val == 1 {
		return "", errors.New("eth value too low to deposit to Rocketpool at this time")
	}

	calldata, err := r.parsedABI.Pack("deposit")
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", r.Method, err)
	}

	calldataHex := HexPrefix + hex.EncodeToString(calldata)
	return calldataHex, nil
}

// GetContractAddress dynamically returns the correct contract address for the operation
func (r *RocketPoolOperation) GetContractAddress(
	_ context.Context) (common.Address, error) {
	switch r.action {
	case NativeStake:
		return *r.contract.Address, nil
	case NativeUnStake:
		return *r.rethContract.Address, nil
	default:
		return common.Address{}, fmt.Errorf("action %d not supported for the rocket pool protocol", r.action)
	}
}

func (r *RocketPoolOperation) Validate(asset common.Address) error {
	if IsNativeToken(asset) {
		return nil
	}

	return fmt.Errorf("unsupported asset for rocket pool staking (%s)", asset.Hex())
}
