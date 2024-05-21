package pkg

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Hex prefix
const HexPrefix = "0x"

type (
	ProtocolName   = string
	ContractAction = string
	// AssetKind describes the way to process an intent
	// TODO:: replace with model after protobuf
	AssetKind       = string
	ContractAddress = common.Address
)
type Protocol struct {
	Name      ProtocolName
	Action    ContractAction
	ChainID   *big.Int
	Address   ContractAddress
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
	Ankr       ProtocolName = "ankr"
	Renzo      ProtocolName = "renzo"
)

const (
	SupplyAction            ContractAction = "supply"
	WithdrawAction          ContractAction = "withdraw"
	LidoStakeAction         ContractAction = "submit"
	StakeAndClaimEthC       ContractAction = "stakeAndClaimAethC"
	UnstakeAndClaimEthC     ContractAction = "unstakeAETH"
	RocketPoolStakeAction   ContractAction = "deposit"
	RocketPoolUnStakeAction ContractAction = "transfer"
	RenzoStakeETHAction     ContractAction = "depositETH"
	RenzoStakeERC20Action   ContractAction = "deposit"
)

var (
	AaveV3ContractAddress    ContractAddress = common.HexToAddress("0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2")
	SparkLendContractAddress ContractAddress = common.HexToAddress("0xC13e21B648A5Ee794902342038FF3aDAB66BE987")
	LidoContractAddress      ContractAddress = common.HexToAddress("0xae7ab96520de3a18e5e111b5eaab095312d7fe84")
	RocketPoolStorageAddress ContractAddress = common.HexToAddress("0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46")
	AnkrContractAddress      ContractAddress = common.HexToAddress("0x84db6ee82b7cf3b47e8f19270abde5718b936670")
	RenzoManagerAddress      ContractAddress = common.HexToAddress("0x74a09653A083691711cF8215a6ab074BB4e99ef5")
)

const (
	AaveV3SupplyABI      = `[{"name":"supply","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"},{"type":"uint16"}]}]`
	AaveV3WithdrawABI    = `[{"name":"withdraw","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"}]}]`
	SparkSupplyABI       = AaveV3SupplyABI
	SparkWithdrawABI     = AaveV3WithdrawABI
	LidoSubmitABI        = `[{"name": "submit", "type": "function","inputs": [{"type": "address"}]}]`
	AnkrSupplyABI        = `[{"name":"stakeAndClaimAethC","type":"function","inputs":[]}]`
	AnkrWithdrawABI      = `[{"name":"unstakeAETH","type":"function","inputs":[{"internalType":"uint256","name":"shares","type":"uint256"}]}]`
	RenzoDepositETHABI   = `[{"name":"depositETH","type":"function","inputs":[]}]`
	RenzoDepositERC20ABI = `[{"name":"deposit","type":"function","inputs":[{"type":"address"},{"type":"uint256"}]}]`
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
			Action:  LidoStakeAction,
			ChainID: big.NewInt(1),
			Address: LidoContractAddress,
			ABI:     LidoSubmitABI,
		},
		{
			Name:    RocketPool,
			Action:  RocketPoolStakeAction,
			ChainID: big.NewInt(1),
			Address: RocketPoolStorageAddress,
			ABI:     RocketPoolABI,
		},
		{
			Name:    RocketPool,
			Action:  RocketPoolUnStakeAction,
			ChainID: big.NewInt(1),
			Address: RocketPoolStorageAddress,
			ABI:     RocketPoolABI,
		},
		{
			Name:    Ankr,
			Action:  StakeAndClaimEthC,
			ChainID: big.NewInt(1),
			Address: AnkrContractAddress,
			ABI:     AnkrSupplyABI,
		},
		{
			Name:    Ankr,
			Action:  UnstakeAndClaimEthC,
			ChainID: big.NewInt(1),
			Address: AnkrContractAddress,
			ABI:     AnkrWithdrawABI,
		},
		{
			Name:    Renzo,
			Action:  RenzoStakeETHAction,
			ChainID: big.NewInt(1),
			Address: RenzoManagerAddress,
			ABI:     RenzoDepositETHABI,
		},
		{
			Name:    Renzo,
			Action:  RenzoStakeERC20Action,
			ChainID: big.NewInt(1),
			Address: RenzoManagerAddress,
			ABI:     RenzoDepositERC20ABI,
		},
	},
}
