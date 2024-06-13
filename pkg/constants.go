package pkg

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Hex prefix
const HexPrefix = "0x"

type (
	ProtocolName    = string
	ProtocolType    = string
	ContractAction  = int64
	ProtocolMethod  = string
	ContractAddress = common.Address
)
type Protocol struct {
	Name      ProtocolName
	Action    ContractAction
	Method    ProtocolMethod
	ChainID   *big.Int
	Address   ContractAddress
	ABI       string
	ParsedABI abi.ABI
}

const (
	TypeLoan  ProtocolType = "Loan"
	TypeStake ProtocolType = "Stake"
)

const (
	AaveV3     ProtocolName = "aave_v3"
	SparkLend  ProtocolName = "spark_lend"
	Lido       ProtocolName = "lido"
	RocketPool ProtocolName = "rocket_pool"
	Ankr       ProtocolName = "ankr"
	Renzo      ProtocolName = "renzo"
	Compound   ProtocolName = "compound"
)

const (
	LoanSupply ContractAction = iota
	LoanWithdraw
	NativeStake
	NativeUnStake
	// ERC20Stake
	// ERC20UnStake
)

const (
	aaveSupply        ProtocolMethod = "supply"
	aaveWithdraw      ProtocolMethod = "withdraw"
	sparkLendSupply   ProtocolMethod = "supply"
	sparkLendWithdraw ProtocolMethod = "withdraw"
	lidoStake         ProtocolMethod = "submit"
	ankrStake         ProtocolMethod = "stakeAndClaimAethC"
	ankrUnstake       ProtocolMethod = "unstakeAETH"
	rocketPoolStake   ProtocolMethod = "deposit"
	rocketPoolUnStake ProtocolMethod = "transfer"
	renzoStakeETH     ProtocolMethod = "depositETH"
	// renzoStakeERC20   ProtocolMethod = "deposit"
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
	AaveV3SupplyABI       = `[{"name":"supply","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"},{"type":"uint16"}]}]`
	AaveV3WithdrawABI     = `[{"name":"withdraw","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"}]}]`
	SparkSupplyABI        = AaveV3SupplyABI
	SparkWithdrawABI      = AaveV3WithdrawABI
	LidoSubmitABI         = `[{"name": "submit", "type": "function","inputs": [{"type": "address"}]}]`
	AnkrSupplyABI         = `[{"name":"stakeAndClaimAethC","type":"function","inputs":[]}]`
	AnkrWithdrawABI       = `[{"name":"unstakeAETH","type":"function","inputs":[{"internalType":"uint256","name":"shares","type":"uint256"}]}]`
	RenzoDepositETHABI    = `[{"name":"depositETH","type":"function","inputs":[]}]`
	RenzoDepositERC20ABI  = `[{"name":"deposit","type":"function","inputs":[{"type":"address"},{"type":"uint256"}]}]`
	CompoundV3SupplyABI   = `[{"name":"supply","type":"function","inputs":[{"type":"address"},{"type":"uint256"}]}]`
	CompoundV3WithdrawABI = `[{"name":"withdraw","type":"function","inputs":[{"type":"address"},{"type":"uint256"}]}]`
)

// Predefined protocols
var SupportedProtocols = map[ProtocolType][]Protocol{
	TypeLoan: {
		{
			Name:    AaveV3,
			Action:  LoanSupply,
			Method:  aaveSupply,
			ChainID: big.NewInt(1),
			Address: AaveV3ContractAddress,
			ABI:     AaveV3SupplyABI,
		},
		{
			Name:    AaveV3,
			Action:  LoanWithdraw,
			Method:  aaveWithdraw,
			ChainID: big.NewInt(1),
			Address: AaveV3ContractAddress,
			ABI:     AaveV3WithdrawABI,
		},
		{
			Name:    SparkLend,
			Action:  LoanSupply,
			Method:  sparkLendSupply,
			ChainID: big.NewInt(1),
			Address: SparkLendContractAddress,
			ABI:     SparkSupplyABI,
		},
		{
			Name:    SparkLend,
			Action:  LoanWithdraw,
			Method:  sparkLendWithdraw,
			ChainID: big.NewInt(1),
			Address: SparkLendContractAddress,
			ABI:     SparkWithdrawABI,
		},
	},
	TypeStake: {
		{
			Name:    Lido,
			Action:  NativeStake,
			Method:  lidoStake,
			ChainID: big.NewInt(1),
			Address: LidoContractAddress,
			ABI:     LidoSubmitABI,
		},
		{
			Name:    Ankr,
			Action:  NativeStake,
			Method:  ankrStake,
			ChainID: big.NewInt(1),
			Address: AnkrContractAddress,
			ABI:     AnkrSupplyABI,
		},
		{
			Name:    Ankr,
			Action:  NativeUnStake,
			Method:  ankrUnstake,
			ChainID: big.NewInt(1),
			Address: AnkrContractAddress,
			ABI:     AnkrWithdrawABI,
		},
		{
			Name:    Renzo,
			Action:  NativeStake,
			Method:  renzoStakeETH,
			ChainID: big.NewInt(1),
			Address: RenzoManagerAddress,
			ABI:     RenzoDepositETHABI,
		},
		// TODO:: later (not supporting for now)
		// {
		// 	Name:    Renzo,
		// 	Action:  ERC20Stake,
		// 	Method:  renzoStakeERC20,
		// 	ChainID: big.NewInt(1),
		// 	Address: RenzoManagerAddress,
		// 	ABI:     RenzoDepositERC20ABI,
		// },
	},
}
