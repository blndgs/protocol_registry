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

type RocketPoolOperation struct {
	DynamicOperation
	contract     *rocketpool.Contract
	rethContract *rocketpool.Contract
}

func (r *RocketPoolOperation) GenerateCalldata(kind AssetKind, args []interface{}) (string, error) {
	switch r.Action {
	case SubmitAction:
		return r.deposit(args)
	case WithdrawAction:
		return r.withdraw(args)
	}
	return "", errors.New("unsupported action")
}

func (r *RocketPoolOperation) Register(registry *ProtocolRegistry) {
	registry.RegisterProtocolOperation(r.Protocol, r.Action, r.ChainID, r)
}

func NewRocketPool(rpcURL, contractAddress string, action ContractAction) (*RocketPoolOperation, error) {
	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	rp, err := rocketpool.NewRocketPool(ethClient, common.HexToAddress(contractAddress))
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

	p := &RocketPoolOperation{
		DynamicOperation: DynamicOperation{
			Protocol: RocketPool,
			Action:   action,
			ChainID:  big.NewInt(1),
		},
		contract:     contract,
		rethContract: rethContract,
	}

	return p, nil
}

func (r *RocketPoolOperation) withdraw(args []interface{}) (string, error) {
	parsedABI, err := abi.JSON(strings.NewReader(RocketPoolABI))
	if err != nil {
		return "", err
	}

	_, exists := parsedABI.Methods["transfer"]
	if !exists {
		return "", errors.New("unsupported action")
	}

	calldata, err := parsedABI.Pack("transfer", args...)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", r.Action, err)
	}

	calldataHex := HexPrefix + hex.EncodeToString(calldata)
	return calldataHex, nil
}

func (r *RocketPoolOperation) deposit(args []interface{}) (string, error) {
	parsedABI, err := abi.JSON(strings.NewReader(RocketPoolABI))
	if err != nil {
		return "", err
	}

	_, exists := parsedABI.Methods["deposit"]
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

	calldata, err := parsedABI.Pack("deposit")
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", r.Action, err)
	}

	calldataHex := HexPrefix + hex.EncodeToString(calldata)
	return calldataHex, nil
}
