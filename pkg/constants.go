package pkg

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// Hex prefix
const HexPrefix = "0x"

type (
	ProtocolName   string
	ContractAction string
	// AssetKind describes the way to process an intent
	// TODO:: replace with model after protobuf
	AssetKind string
)
type Protocol struct {
	Name      ProtocolName
	Action    ContractAction
	ChainID   *big.Int
	Address   string
	ABI       string
	ParsedABI abi.ABI
}

const (
	// TokenKind describes how to swap an asset onchain
	TokenKind AssetKind = "TOKEN"
	// StakeKind describes an inent to stake an asset onchain
	StakeKind AssetKind = "STAKE"
	// LoanKind describes how to supply an asset to a defi protocol onchain
	LoanKind AssetKind = "LOAN"
)
const (
	AaveV3     ProtocolName = "aave_v3"
	SparkLend  ProtocolName = "spark_lend"
	Lido       ProtocolName = "lido"
	RocketPool ProtocolName = "rocket_pool"
)

const (
	SupplyAction   ContractAction = "supply"
	WithdrawAction ContractAction = "withdraw"
	SubmitAction   ContractAction = "submit"
)

const (
	AaveV3ContractAddress    = "0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2"
	AaveV3SupplyABI          = `[{"name":"supply","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"},{"type":"uint16"}]}]`
	AaveV3WithdrawABI        = `[{"name":"withdraw","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"}]}]`
	SparkLendContractAddress = "0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2" //TODO:: Larne please check and fix the correct address
	SparkSupplyABI           = AaveV3SupplyABI
	SparkWithdrawABI         = AaveV3WithdrawABI
	LidoContractAddress      = "0xae7ab96520de3a18e5e111b5eaab095312d7fe84"
	LidoSubmitABI            = `[{"name": "submit", "type": "function","inputs": [{"type": "address"}]}]`
	RocketPoolStorageAddress = "0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46"
)

// Predefined protocols
var SupportedProtocols = map[AssetKind][]Protocol{
	LoanKind: {
		{
			Name:    AaveV3,
			Action:  SupplyAction,
			ChainID: big.NewInt(1),
			Address: AaveV3ContractAddress,
			ABI:     AaveV3SupplyABI,
		},
		{
			Name:    AaveV3,
			Action:  WithdrawAction,
			ChainID: big.NewInt(1),
			Address: AaveV3ContractAddress,
			ABI:     AaveV3WithdrawABI,
		},
		{
			Name:    SparkLend,
			Action:  SupplyAction,
			ChainID: big.NewInt(1),
			Address: SparkLendContractAddress,
			ABI:     SparkSupplyABI,
		},
		{
			Name:    SparkLend,
			Action:  WithdrawAction,
			ChainID: big.NewInt(1),
			Address: SparkLendContractAddress,
			ABI:     SparkWithdrawABI,
		},
	},
	StakeKind: {
		{
			Name:    Lido,
			Action:  SubmitAction,
			ChainID: big.NewInt(1),
			Address: LidoContractAddress,
			ABI:     LidoSubmitABI,
		},
		{
			Name:    RocketPool,
			Action:  SubmitAction,
			ChainID: big.NewInt(1),
			Address: RocketPoolStorageAddress,
			ABI:     RocketPoolABI,
		},
		{
			Name:    RocketPool,
			Action:  WithdrawAction,
			ChainID: big.NewInt(1),
			Address: RocketPoolStorageAddress,
			ABI:     RocketPoolABI,
		},
	},
}
