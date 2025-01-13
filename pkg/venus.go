package pkg

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// venusABIs contains all abis used by pool registry, vtokens and others
// this allows us only compile and init once
const venusABIs = `
[
{
  "inputs": [],
  "name": "getAllPools",
  "outputs": [
    {
      "components": [
        {
          "internalType": "string",
          "name": "name",
          "type": "string"
        },
        {
          "internalType": "address",
          "name": "creator",
          "type": "address"
        },
        {
          "internalType": "address",
          "name": "comptroller",
          "type": "address"
        },
        {
          "internalType": "uint256",
          "name": "blockPosted",
          "type": "uint256"
        },
        {
          "internalType": "uint256",
          "name": "timestampPosted",
          "type": "uint256"
        }
      ],
      "internalType": "struct PoolRegistryInterface.VenusPool[]",
      "name": "",
      "type": "tuple[]"
    }
  ],
  "stateMutability": "view",
  "type": "function"
},
  {
    "inputs": [],
    "name": "getAllMarkets",
    "outputs": [
      {
        "internalType": "address[]",
        "name": "",
        "type": "address[]"
      }
    ],
    "stateMutability": "view",
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
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "borrower",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "mintAmount",
        "type": "uint256"
      }
    ],
    "name": "mintBehalf",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "mintAmount",
        "type": "uint256"
      }
    ],
    "name": "mint",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "minter",
        "type": "address"
      }
    ],
    "name": "wrapAndSupply",
    "outputs": [],
    "stateMutability": "payable",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "underlying",
    "outputs": [
      {
        "internalType": "address",
        "name": "",
        "type": "address"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "badDebt",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "symbol",
    "outputs": [
      {
        "internalType": "string",
        "name": "",
        "type": "string"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint256",
        "name": "redeemTokens",
        "type": "uint256"
      }
    ],
    "name": "redeem",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "owner",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "redeemTokens",
        "type": "uint256"
      }
    ],
    "name": "redeemBehalf",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
	{
  "inputs": [
    {
      "internalType": "uint256",
      "name": "redeemTokens",
      "type": "uint256"
    }
  ],
  "name": "redeemAndUnwrap",
  "outputs": [],
  "stateMutability": "nonpayable",
  "type": "function"
 },

{
  "name": "actionPaused",
  "type": "function",
  "inputs": [
    {
      "internalType": "address",
      "name": "vToken",
      "type": "address"
    },
    {
      "internalType": "enum Action",
      "name": "action",
      "type": "uint8"
    }
  ],
  "outputs": [
    {
      "internalType": "bool",
      "name": "",
      "type": "bool"
    }
  ]
},
{
  "inputs": [
    {
      "internalType": "address",
      "name": "account",
      "type": "address"
    },
    {
      "internalType": "uint256",
      "name": "amount",
      "type": "uint256"
    }
  ],
  "name": "redeemUnderlyingBehalf",
  "outputs": [],
  "stateMutability": "nonpayable",
  "type": "function"
},
{
  "inputs": [
    {
      "internalType": "uint256",
      "name": "amount",
      "type": "uint256"
    }
  ],
  "name": "redeemUnderlying",
  "outputs": [],
  "stateMutability": "nonpayable",
  "type": "function"
}
]
`

var venusBNBContractAddress ContractAddress = common.HexToAddress("0x9F7b01A536aFA00EF10310A162877fd792cD0666")

// VenusOperation implements the Protocol interface for Lido
type VenusOperation struct {
	parsedABI abi.ABI
	contract  common.Address
	chainID   *big.Int
	version   string

	poolMarkets map[string]struct {
		address string
		// VWBNB can supply BNB or native assets. In this instance, we want to use the NativeGateway instead
		isNative bool
	}

	assetToVTokenMap map[string]string

	client *ethclient.Client
}

// dynamically registers all supported pools
func registerVenusPools(registry ProtocolRegistry, client *ethclient.Client, chainID int64) error {
	parsedABI, err := abi.JSON(strings.NewReader(venusABIs))
	if err != nil {
		return err
	}

	data, err := parsedABI.Pack("getAllPools")
	if err != nil {
		log.Fatalf("Failed to pack contract call: %v", err)
	}

	callMsg := ethereum.CallMsg{
		To:   &venusBNBContractAddress,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return err
	}

	var pools []struct {
		Name            string
		Creator         common.Address
		Comptroller     common.Address
		BlockPosted     *big.Int
		TimestampPosted *big.Int
	}
	err = parsedABI.UnpackIntoInterface(&pools, "getAllPools", result)
	if err != nil {
		return err
	}

	for _, poolAddr := range pools {
		c, err := NewVenusOperation(client, big.NewInt(chainID), poolAddr.Comptroller)
		if err != nil {
			return err
		}

		if err := registry.RegisterProtocol(big.NewInt(chainID), poolAddr.Comptroller, c); err != nil {
			return err
		}
	}

	return nil
}

func NewVenusOperation(client *ethclient.Client,
	chainID *big.Int,
	marketPool common.Address) (*VenusOperation, error) {
	if !IsBnb(chainID) {
		return nil, fmt.Errorf("unsupported chain ID (%d)", chainID.Int64())
	}

	if client == nil {
		return nil, errors.New("ethclient cannot be nil")
	}

	parsedABI, err := abi.JSON(strings.NewReader(venusABIs))
	if err != nil {
		return nil, err
	}

	v := &VenusOperation{
		parsedABI: parsedABI,
		contract:  marketPool,
		chainID:   chainID,
		version:   "4",
		client:    client,
		poolMarkets: map[string]struct {
			address  string
			isNative bool
		}{},
		assetToVTokenMap: make(map[string]string),
	}

	if err := v.getSupportedAssets(); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *VenusOperation) getSupportedAssets() error {

	getAllMarketsCalldata, err := v.parsedABI.Pack("getAllMarkets")
	if err != nil {
		return err
	}

	msg := ethereum.CallMsg{
		To:   &v.contract,
		Data: getAllMarketsCalldata,
	}

	result, err := v.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return err
	}

	var marketInfo []common.Address

	err = v.parsedABI.UnpackIntoInterface(&marketInfo, "getAllMarkets", result)
	if err != nil {
		return fmt.Errorf("failed to unpack output: %v", err)
	}

	// get underlying token to add to map
	for _, addr := range marketInfo {

		getSymbol, err := v.parsedABI.Pack("symbol")
		if err != nil {
			return err
		}

		msg := ethereum.CallMsg{
			To:   &addr,
			Data: getSymbol,
		}

		result, err := v.client.CallContract(context.Background(), msg, nil)
		if err != nil {
			return err
		}

		var symbol string

		err = v.parsedABI.UnpackIntoInterface(&symbol, "symbol", result)
		if err != nil {
			return fmt.Errorf("failed to unpack symbol output: %v", err)
		}

		getUnderlying, err := v.parsedABI.Pack("underlying")
		if err != nil {
			return err
		}

		msg = ethereum.CallMsg{
			To:   &addr,
			Data: getUnderlying,
		}

		result, err = v.client.CallContract(context.Background(), msg, nil)
		if err != nil {
			return err
		}

		var underlyingAddr common.Address

		err = v.parsedABI.UnpackIntoInterface(&underlyingAddr, "underlying", result)
		if err != nil {
			return fmt.Errorf("failed to unpack underlying output: %v", err)
		}

		v.poolMarkets[addr.Hex()] = struct {
			address  string
			isNative bool
		}{
			address:  underlyingAddr.Hex(),
			isNative: symbol == "vWBNB_LiquidStakedBNB",
		}

		v.assetToVTokenMap[underlyingAddr.Hex()] = addr.Hex()
		if v.poolMarkets[addr.Hex()].isNative {
			v.assetToVTokenMap[nativeDenomAddress] = addr.Hex()
		}
	}

	return nil
}

// GenerateCalldata creates the necessary blockchain transaction data to
// supply or withdraw your asset
func (l *VenusOperation) GenerateCalldata(ctx context.Context, chainID *big.Int,
	action ContractAction, params TransactionParams) (string, error) {
	if !IsBnb(chainID) {
		return "", ErrChainUnsupported
	}

	var calldata []byte
	var err error

	switch action {
	case LoanSupply:

		calldata, err = l.parsedABI.Pack("mint", params.Amount)
		if err != nil {
			return "", err
		}

		if !IsZeroAddress(params.Recipient) {
			calldata, err = l.parsedABI.Pack("mintBehalf", params.Recipient, params.Amount)
			if err != nil {
				return "", err
			}
		}

		if IsNativeToken(params.Asset) {
			calldata, err = l.parsedABI.Pack("wrapAndSupply", params.Sender)
			if err != nil {
				return "", err
			}
		}

	case LoanWithdraw:

		calldata, err = l.parsedABI.Pack("redeemUnderlying", params.Amount)
		if err != nil {
			return "", err
		}

		if !IsZeroAddress(params.Recipient) {
			calldata, err = l.parsedABI.Pack("redeemUnderlyingBehalf", params.Recipient, params.Amount)
			if err != nil {
				return "", err
			}
		}

		if IsNativeToken(params.Asset) {
			calldata, err = l.parsedABI.Pack("redeemAndUnwrap", params.Amount)
			if err != nil {
				return "", err
			}
		}

	default:
		return "", errors.New("action not supported")
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

// Validate checks if the provided parameters are valid for the specified action
func (l *VenusOperation) Validate(ctx context.Context,
	chainID *big.Int, action ContractAction, params TransactionParams) error {

	if !IsBnb(chainID) {
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

	_, balance, err := l.GetBalance(ctx, l.chainID, params.Sender, params.Asset)
	if err != nil {
		return err
	}

	if balance.Cmp(params.Amount) == -1 {
		return errors.New("balance not enough")
	}

	return nil
}

// GetBalance retrieves the balance for a specified account and asset
func (l *VenusOperation) GetBalance(ctx context.Context,
	chainID *big.Int, account, asset common.Address) (common.Address, *big.Int, error) {

	var address common.Address
	if !IsBnb(chainID) {
		return address, nil, ErrChainUnsupported
	}

	contract, ok := l.assetToVTokenMap[asset.Hex()]
	if !ok {
		return common.Address{}, big.NewInt(0), errors.New("asset not supported")
	}

	callData, err := l.parsedABI.Pack("balanceOf", account)
	if err != nil {
		return address, nil, err
	}

	c := common.HexToAddress(contract)

	result, err := l.client.CallContract(ctx, ethereum.CallMsg{
		To:   &c,
		Data: callData,
	}, nil)
	if err != nil {
		return address, nil, err
	}

	balance := new(big.Int)
	err = l.parsedABI.UnpackIntoInterface(&balance, "balanceOf", result)
	return c, balance, err
}

// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain
func (l *VenusOperation) GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error) {
	supportedAssets := make([]common.Address, 0)

	// only add native token once. Ideally this might not be needed as vBNB pool is the only one
	// right now and most realistic one ever but doesn't hurt to be defensive here
	var nativeTokenAdded bool

	for _, v := range l.poolMarkets {
		supportedAssets = append(supportedAssets, common.HexToAddress(v.address))

		if v.isNative && !nativeTokenAdded {
			supportedAssets = append(supportedAssets, common.HexToAddress(nativeDenomAddress))
			nativeTokenAdded = true
		}
	}

	return supportedAssets, nil
}

// IsSupportedAsset checks if the specified asset is supported on the given chain
func (l *VenusOperation) IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool {
	if !IsBnb(chainID) {
		return false
	}

	assets, err := l.GetSupportedAssets(ctx, chainID)
	if err != nil {
		return false
	}

	for _, a := range assets {
		if asset.Hex() == a.Hex() {
			return true
		}
	}

	return false
}

// GetProtocolConfig returns the protocol config for a specific chain
func (l *VenusOperation) GetProtocolConfig(chainID *big.Int) ProtocolConfig {
	return ProtocolConfig{
		ChainID:  l.chainID,
		Contract: l.contract,
		ABI:      l.parsedABI,
		Type:     TypeLoan,
	}
}

// GetABI returns the ABI of the protocol's contract
func (l *VenusOperation) GetABI(chainID *big.Int) abi.ABI { return l.parsedABI }

// GetType returns the protocol type
func (l *VenusOperation) GetType() ProtocolType { return TypeLoan }

// GetContractAddress returns the contract address for a specific chain
func (l *VenusOperation) GetContractAddress(chainID *big.Int) common.Address { return l.contract }

// Name returns the human readable name for the protocol
func (l *VenusOperation) GetName() string { return VenusProtocolIsolatedPool }

// GetVersion returns the version of the protocol
func (l *VenusOperation) GetVersion() string { return l.version }
