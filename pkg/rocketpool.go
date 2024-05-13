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

var _ ProtocolOperation = (*RocketPoolOperation)(nil)

const (
	storageAddress             = "0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46"
	rocketPoolForwarderAddress = "rocketDepositPool"

	rocketPoolABI = `
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

type RocketPoolOperation struct {
	protocol    ProtocolName
	action      ContractAction
	chainID     *big.Int
	poolAddress common.Address
	rethAddress common.Address

	contract     *rocketpool.Contract
	rethContract *rocketpool.Contract
	parsedABI    abi.ABI
}

func NewRocketPool(rpcURL string, action ContractAction) (*RocketPoolOperation, error) {

	if action != SubmitAction && action != WithdrawAction {
		return nil, errors.New("unsupported action type")
	}

	parsedABI, err := abi.JSON(strings.NewReader(rocketPoolABI))
	if err != nil {
		return nil, err
	}

	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	rp, err := rocketpool.NewRocketPool(ethClient, common.HexToAddress(storageAddress))
	if err != nil {
		return nil, err
	}

	addr, err := rp.GetAddress(rocketPoolForwarderAddress, &bind.CallOpts{})
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

	p := &RocketPoolOperation{
		poolAddress:  *contract.Address,
		rethAddress:  *rethContract.Address,
		protocol:     RocketPool,
		action:       action,
		chainID:      big.NewInt(1),
		contract:     contract,
		rethContract: rethContract,
		parsedABI:    parsedABI,
	}

	// Use time.Ticker to always update the address?
	// Might be better that way as we essentially have a form of cache as against
	// reaching out to the network on every call

	return p, nil
}

func (r *RocketPoolOperation) Register(registry *ProtocolRegistry) {
	registry.RegisterProtocolOperation(r.protocol, r.action, r.chainID, r)
}

func (r *RocketPoolOperation) GetContractAddress(
	_ context.Context) (common.Address, error) {
	switch r.action {
	case SubmitAction:
		return r.poolAddress, nil

	default:
		return r.rethAddress, nil
	}
}

func (r *RocketPoolOperation) GenerateCalldata(kind AssetKind, args []interface{}) (string, error) {

	switch r.action {
	case SubmitAction:
		return r.deposit(args)
	case WithdrawAction:
		return r.withdraw(args)
	}

	return "", errors.New("unsupported action")
}

func (r *RocketPoolOperation) withdraw(args []interface{}) (string, error) {

	_, exists := r.parsedABI.Methods["transfer"]
	if !exists {
		return "", errors.New("unsupported action")
	}

	calldata, err := r.parsedABI.Pack("transfer", args...)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", r.action, err)
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

	// no args here
	calldata, err := r.parsedABI.Pack("deposit")
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", r.action, err)
	}

	calldataHex := HexPrefix + hex.EncodeToString(calldata)
	return calldataHex, nil
}
