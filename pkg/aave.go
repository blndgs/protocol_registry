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

// ENUM(ethereum,spark,avalon_finance,polygon)
//
// AaveProtocolDeployment matches the numerous deployments of Aave.
// The naming convention here is:
// - If the Aave team themself deploy on a chain, name it the chain,
// - Else if it is a fork deployed by another team, name it that specific protocol
// name
type AaveProtocolDeployment uint8

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
	ethAaveDataProviderContract       = common.HexToAddress("0x7B4EB56E7CD4b454BA8ff71E4518426369a138a3")
	polygonAaveDataProviderContract   = common.HexToAddress("0x5598BbFA2f4fE8151f45bBA0a3edE1b54B51a0a9")
	ethSparklendProviderContract      = common.HexToAddress("0xFc21d6d146E6086B8359705C8b28512a983db0cb")
	bnbAaveDataProviderContract       = common.HexToAddress("0x41585C50524fb8c3899B43D7D797d9486AAc94DB")
	avalonFinanceDataProviderContract = common.HexToAddress("0x672b19DdA450120C505214D149Ee7F7B6DEd8C39")
)

// AaveOperation implements the Protocol interface for Aave
type AaveOperation struct {
	parsedABI       abi.ABI
	dataProviderABI abi.ABI
	contract        common.Address
	chainID         *big.Int
	version         string
	fork            AaveProtocolDeployment
	erc20ABI        abi.ABI

	client *ethclient.Client
}

func isAaveChainSupported(chainID *big.Int, fork AaveProtocolDeployment) error {

	if !IsBnb(chainID) && !IsEth(chainID) {
		return errors.New("only eth and bnb chains are supported")
	}

	if IsBnb(chainID) && fork == AaveProtocolDeploymentSpark {
		return errors.New("spark finance is not supported on Bnb chain. Only Ethereum")
	}

	if IsPolygon(chainID) && fork != AaveProtocolDeploymentPolygon {
		return errors.New("only the official aave deployment on Polygon is supported at the moment")
	}

	return nil
}

func NewAaveOperation(
	client *ethclient.Client,
	chainID *big.Int,
	fork AaveProtocolDeployment,
) (*AaveOperation, error) {

	if err := isAaveChainSupported(chainID, fork); err != nil {
		return nil, err
	}

	if !fork.IsValid() {
		return nil, errors.New("invalid Aave fork")
	}

	networkID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("client.NetworkID: could not fetch network id... %w", err)
	}

	if networkID.Cmp(chainID) != 0 {
		return nil, fmt.Errorf("network id of client(%d) does not match chainID provided (%d)",
			networkID.Int64(), chainID.Int64())
	}

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

	var contract common.Address

	switch fork {
	case AaveProtocolDeploymentEthereum:
		contract = AaveV3ContractAddress
		if chainID.Cmp(BscChainID) == 0 {
			contract = AaveBnbV3ContractAddress
		}
	case AaveProtocolDeploymentAvalonFinance:
		contract = AvalonFinanceContractAddress
	case AaveProtocolDeploymentSpark:
		contract = SparkLendContractAddress
	case AaveProtocolDeploymentPolygon:
		contract = polygonAaveDataProviderContract
	}

	var version string = "3"
	if fork == AaveProtocolDeploymentAvalonFinance {
		version = "2"
	}

	return &AaveOperation{
		dataProviderABI: dataProviderABI,
		parsedABI:       parsedABI,
		erc20ABI:        erc20ABI,
		contract:        contract,
		chainID:         chainID,
		version:         version,
		client:          client,
		fork:            fork,
	}, nil
}

// GenerateCalldata creates the necessary blockchain transaction data
func (a *AaveOperation) GenerateCalldata(ctx context.Context, chainID *big.Int,
	action ContractAction, params TransactionParams) (string, error) {

	if err := isAaveChainSupported(a.chainID, a.fork); err != nil {
		return "", err
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

	var toContract common.Address
	switch {
	case IsEth(l.chainID):
		toContract = ethAaveDataProviderContract
		if l.fork == AaveProtocolDeploymentSpark {
			toContract = ethSparklendProviderContract
		}

	case IsBnb(l.chainID):
		if l.fork == AaveProtocolDeploymentSpark {
			return common.HexToAddress(""), errors.New("BSC: spark finance is not supported on Aave")
		}

		toContract = bnbAaveDataProviderContract
		if l.fork == AaveProtocolDeploymentAvalonFinance {
			toContract = avalonFinanceDataProviderContract
		}
	default:
		return common.HexToAddress(""), errors.New("unsupported chain")
	}

	result, err := l.client.CallContract(ctx, ethereum.CallMsg{
		To:   &toContract,
		Data: calldata,
	}, nil)
	if err != nil {
		return common.Address{}, err
	}

	addr := common.Address{}
	// dummy value we just need to unpack successfully
	value := common.Address{}
	err = l.dataProviderABI.UnpackIntoInterface(&[]interface{}{&addr, &value, &value}, "getReserveTokensAddresses", result)
	if err != nil {
		return common.Address{}, err
	}

	if addr.Hex() == zeroAddress {
		return common.Address{}, errors.New("asset not supported")
	}

	return addr, nil
}

// Validate checks if the provided parameters are valid for the specified action
func (l *AaveOperation) Validate(ctx context.Context,
	chainID *big.Int, action ContractAction, params TransactionParams) error {

	if err := isAaveChainSupported(l.chainID, l.fork); err != nil {
		return err
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

	_, balance, err := l.GetBalance(ctx, l.chainID, params.Sender, asset)
	if err != nil {
		return err
	}

	if balance.Cmp(params.Amount) == -1 {
		return errors.New("balance not enough")
	}

	return nil
}

// GetBalance retrieves the balance for a specified account and asset
func (l *AaveOperation) GetBalance(ctx context.Context,
	chainID *big.Int, account,
	asset common.Address) (common.Address, *big.Int, error) {

	var address common.Address

	if err := isAaveChainSupported(l.chainID, l.fork); err != nil {
		return address, nil, err
	}

	callData, err := l.erc20ABI.Pack("balanceOf", account)
	if err != nil {
		return address, nil, err
	}

	aToken, err := l.getAToken(ctx, asset)
	if err != nil {
		return address, nil, err
	}

	result, err := l.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &aToken,
		Data: callData,
	}, nil)
	if err != nil {
		return address, nil, err
	}

	balance := new(big.Int)
	err = l.erc20ABI.UnpackIntoInterface(&balance, "balanceOf", result)
	return aToken, balance, err
}

// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain
func (l *AaveOperation) GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error) {

	if err := isAaveChainSupported(l.chainID, l.fork); err != nil {
		return []common.Address{}, err
	}

	var protocol = AaveV3

	switch l.fork {
	case AaveProtocolDeploymentEthereum:
		protocol = AaveV3
	case AaveProtocolDeploymentAvalonFinance:
		protocol = AvalonFinance
	case AaveProtocolDeploymentSpark:
		protocol = SparkLend
	}

	assets := make([]common.Address, 0, len(tokenSupportedMap[l.chainID.Int64()][protocol]))

	for _, v := range tokenSupportedMap[l.chainID.Int64()][protocol] {
		assets = append(assets, common.HexToAddress(v))
	}

	return assets, nil
}

// IsSupportedAsset checks if the specified asset is supported on the given chain
func (l *AaveOperation) IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool {
	if err := isAaveChainSupported(l.chainID, l.fork); err != nil {
		return false
	}

	protocols, ok := tokenSupportedMap[l.chainID.Int64()]
	if !ok {
		return false
	}

	var protocol = AaveV3

	switch l.fork {
	case AaveProtocolDeploymentEthereum:
		protocol = AaveV3
	case AaveProtocolDeploymentAvalonFinance:
		protocol = AvalonFinance
	case AaveProtocolDeploymentSpark:
		protocol = SparkLend
	}

	addrs, ok := protocols[protocol]
	if !ok {
		return false
	}

	if len(addrs) == 0 {
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

	if l.fork == AaveProtocolDeploymentEthereum {
		return AaveV3
	}

	return SparkLend
}

// GetVersion returns the version of the protocol
func (l *AaveOperation) GetVersion() string { return l.version }
