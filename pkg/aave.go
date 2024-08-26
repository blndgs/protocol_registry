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

// ENUM(aave,spark)
type AaveProtocolFork uint8

const aaveV3ABI = `
 [
   {
     "name": "withdraw",
     "type": "function",
     "inputs": [
       {
         "type": "address"
       },
       {
         "type": "uint256"
       },
       {
         "type": "address"
       }
     ]
   },
   {
     "name": "supply",
     "type": "function",
     "inputs": [
       {
         "type": "address"
       },
       {
         "type": "uint256"
       },
       {
         "type": "address"
       },
       {
         "type": "uint16"
       }
     ]
   }
 ]
	`

const aaveDataProviderABI = `
[
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "asset",
        "type": "address"
      }
    ],
    "name": "getReserveTokensAddresses",
    "outputs": [
      {
        "internalType": "address",
        "name": "aTokenAddress",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "stableDebtTokenAddress",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "variableDebtTokenAddress",
        "type": "address"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]`

var (
	aaveDataProviderContract  = common.HexToAddress("0x7B4EB56E7CD4b454BA8ff71E4518426369a138a3")
	sparklendProviderContract = common.HexToAddress("0xFc21d6d146E6086B8359705C8b28512a983db0cb")
)

// AaveOperation implements the Protocol interface for Aave
type AaveOperation struct {
	parsedABI       abi.ABI
	dataProviderABI abi.ABI
	contract        common.Address
	chainID         *big.Int
	version         string
	fork            AaveProtocolFork
	erc20ABI        abi.ABI

	client *ethclient.Client
}

func NewAaveOperation(client *ethclient.Client, chainID *big.Int, fork AaveProtocolFork) (*AaveOperation, error) {
	parsedABI, err := abi.JSON(strings.NewReader(aaveV3ABI))
	if err != nil {
		return nil, err
	}

	dataProviderABI, err := abi.JSON(strings.NewReader(aaveDataProviderABI))
	if err != nil {
		return nil, err
	}

	erc20ABI, err := abi.JSON(strings.NewReader(erc20BalanceOfABI))
	if err != nil {
		return nil, err
	}

	return &AaveOperation{
		parsedABI:       parsedABI,
		contract:        AaveV3ContractAddress,
		chainID:         chainID,
		version:         "3",
		client:          client,
		erc20ABI:        erc20ABI,
		fork:            fork,
		dataProviderABI: dataProviderABI,
	}, nil
}

// GenerateCalldata creates the necessary blockchain transaction data
func (a *AaveOperation) GenerateCalldata(ctx context.Context, chainID *big.Int,
	action ContractAction, params TransactionParams) (string, error) {
	if chainID.Int64() != 1 {
		return "", ErrChainUnsupported
	}

	var calldata []byte
	var err error

	switch action {
	case LoanSupply:

		referalCode, ok := params.ExtraData["referral_code"].(uint16)
		if !ok {
			return "", errors.New("referal code is not a uint16")
		}

		calldata, err = a.parsedABI.Pack("supply",
			params.Asset, params.Amount, params.GetBeneficiaryOwner(), referalCode)
		if err != nil {
			return "", err
		}

	case LoanWithdraw:

		calldata, err = a.parsedABI.Pack("withdraw",
			params.Asset, params.Amount, params.GetBeneficiaryOwner())
		if err != nil {
			return "", err
		}

	default:
		return "", errors.New("operation not supported")
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

func (l *AaveOperation) getAToken(ctx context.Context, asset common.Address) (common.Address, error) {

	calldata, err := l.dataProviderABI.Pack("getReserveTokensAddresses", asset)
	if err != nil {
		return common.Address{}, err
	}

	var toContract = aaveDataProviderContract
	if l.fork == AaveProtocolForkSpark {
		toContract = sparklendProviderContract
	}

	result, err := l.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &toContract,
		Data: calldata,
	}, nil)
	if err != nil {
		return common.Address{}, err
	}

	addr := common.Address{}
	// dummy value we just need to unpack successfully
	value := common.Address{}
	return addr, l.dataProviderABI.UnpackIntoInterface(&[]interface{}{&addr, &value, &value}, "getReserveTokensAddresses", result)
}

// Validate checks if the provided parameters are valid for the specified action
func (l *AaveOperation) Validate(ctx context.Context,
	chainID *big.Int, action ContractAction, params TransactionParams) error {

	if chainID.Int64() != 1 {
		return ErrChainUnsupported
	}

	if !l.IsSupportedAsset(ctx, l.chainID, params.Asset) {
		return fmt.Errorf("asset not supported %s", params.Asset)
	}

	if action != LoanSupply && action != LoanWithdraw {
		return errors.New("unsupported action")
	}

	if params.Amount.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("amount must be greater than zero")
	}

	if action == LoanSupply {
		return nil
	}

	asset, err := l.getAToken(ctx, params.Asset)
	if err != nil {
		return err
	}

	balance, err := l.GetBalance(ctx, l.chainID, params.Sender, asset)
	if err != nil {
		return err
	}

	if balance.Cmp(params.Amount) == -1 {
		return errors.New("balance not enough")
	}

	return nil
}

// GetBalance retrieves the balance for a specified account and asset
func (l *AaveOperation) GetBalance(ctx context.Context, chainID *big.Int, account, asset common.Address) (*big.Int, error) {
	if chainID.Int64() != 1 {
		return nil, ErrChainUnsupported
	}

	callData, err := l.erc20ABI.Pack("balanceOf", account)
	if err != nil {
		return nil, err
	}

	result, err := l.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &asset,
		Data: callData,
	}, nil)
	if err != nil {
		return nil, err
	}

	balance := new(big.Int)
	err = l.erc20ABI.UnpackIntoInterface(&balance, "balanceOf", result)
	return balance, err
}

// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain
func (l *AaveOperation) GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error) {
	assets := make([]common.Address, 0, len(tokenSupportedMap[1][AaveV3]))

	for _, v := range tokenSupportedMap[1][AaveV3] {
		assets = append(assets, common.HexToAddress(v))
	}

	return assets, nil
}

// IsSupportedAsset checks if the specified asset is supported on the given chain
func (l *AaveOperation) IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool {
	if chainID.Int64() != 1 {
		return false
	}
	protocols, ok := tokenSupportedMap[1]
	if !ok {
		return false
	}

	var protocol = AaveV3
	if l.fork == AaveProtocolForkSpark {
		protocol = SparkLend
	}

	addrs, ok := protocols[protocol]
	if !ok {
		return false
	}

	if len(addrs) == 0 {
		if strings.EqualFold(strings.ToLower(asset.Hex()), nativeDenomAddress) {
			return false
		}

		return false
	}

	for _, addr := range addrs {
		if strings.EqualFold(strings.ToLower(asset.Hex()), strings.ToLower(addr)) {
			return true
		}
	}

	return false
}

// GetProtocolConfig returns the protocol config for a specific chain
func (l *AaveOperation) GetProtocolConfig(chainID *big.Int) ProtocolConfig {
	return ProtocolConfig{
		ChainID:  l.chainID,
		Contract: l.contract,
		ABI:      l.parsedABI,
		Type:     TypeStake,
	}
}

// GetABI returns the ABI of the protocol's contract
func (l *AaveOperation) GetABI(chainID *big.Int) abi.ABI { return l.parsedABI }

// GetType returns the protocol type
func (l *AaveOperation) GetType() ProtocolType { return TypeLoan }

// GetContractAddress returns the contract address for a specific chain
func (l *AaveOperation) GetContractAddress(chainID *big.Int) common.Address { return l.contract }

// Name returns the human readable name for the protocol
func (l *AaveOperation) GetName() string {

	if l.fork == AaveProtocolForkAave {
		return AaveV3
	}

	return SparkLend
}

// GetVersion returns the version of the protocol
func (l *AaveOperation) GetVersion() string { return l.version }
