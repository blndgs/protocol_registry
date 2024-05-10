package pkg

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

var _ ProtocolOperation = (*RocketPoolOperation)(nil)

const (
	storageAddress             = "0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46"
	rocketPoolForwarderAddress = "rocketDepositPool"
)

type RocketPoolOperation struct {
	protocol    ProtocolName
	action      ContractAction
	chainID     *big.Int
	poolAddress common.Address

	contract *rocketpool.Contract
}

func NewRocketPool(rpcURL string) (*RocketPoolOperation, error) {

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

	contract, err := rp.MakeContract("rocketDepositPool", *addr, &bind.CallOpts{})
	if err != nil {
		return nil, err
	}

	p := &RocketPoolOperation{
		poolAddress: *contract.Address,
		protocol:    RocketPool,
		action:      SubmitAction,
		chainID:     big.NewInt(1),
		contract:    contract,
	}

	return p, nil
}

func (r *RocketPoolOperation) GetContractAddress(
	_ context.Context) (common.Address, error) {
	return r.poolAddress, nil
}

func (r *RocketPoolOperation) GenerateCalldata(kind AssetKind, args []interface{}) (string, error) {

	var amount = big.NewInt(0)

	if err := r.contract.Call(&bind.CallOpts{}, &amount, "getMaximumDepositAmount"); err != nil {
		return "", err
	}

	return "", nil
}
