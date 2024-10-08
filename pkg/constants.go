package pkg

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const (
	ReferralAddress = "0x000000000000000000000000000000000000dEaD"
)

const (
	EthChainStr     = "1"
	BscChainStr     = "56"
	PolygonChainStr = "137"
)

var (
	EthChainID     = big.NewInt(1)
	BscChainID     = big.NewInt(56)
	PolygonChainID = big.NewInt(137)
)

// Hex prefix
const HexPrefix = "0x"

var ErrChainUnsupported = errors.New("chain not supported")

type (
	ProtocolName    = string
	ProtocolMethod  = string
	ContractAddress = common.Address
)

type ContractAction int64

type ProtocolType string

type Protocol interface {
	// Initialize(ctx context.Context, config ProtocolConfig) error
	GenerateCalldata(ctx context.Context, chainID *big.Int, action ContractAction, params TransactionParams) (string, error)
	Validate(ctx context.Context, chainID *big.Int, action ContractAction, params TransactionParams) error
	GetBalance(ctx context.Context, chainID *big.Int, account, asset common.Address) (common.Address, *big.Int, error)
	GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error)
	IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool
	GetProtocolConfig(chainID *big.Int) ProtocolConfig
	GetABI(chainID *big.Int) abi.ABI
	GetType() ProtocolType
	GetName() string
	GetVersion() string
	GetContractAddress(chainID *big.Int) common.Address
}

const (
	AaveV3        ProtocolName = "aave_v3"
	SparkLend     ProtocolName = "spark_lend"
	Lido          ProtocolName = "lido"
	RocketPool    ProtocolName = "rocket_pool"
	Ankr          ProtocolName = "ankr"
	Renzo         ProtocolName = "renzo"
	Compound      ProtocolName = "compound"
	ListaDao      ProtocolName = "lista_dao"
	AvalonFinance ProtocolName = "avalon_finance"
)

var (
	AaveEthereumV3ContractAddress ContractAddress = common.HexToAddress("0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2")
	AaveBnbV3ContractAddress      ContractAddress = common.HexToAddress("0x6807dc923806fE8Fd134338EABCA509979a7e0cB")
	AavePolygonV3ContractAddress  ContractAddress = common.HexToAddress("0x794a61358D6845594F94dc1DB02A252b5b4814aD")
	SparkLendContractAddress      ContractAddress = common.HexToAddress("0xC13e21B648A5Ee794902342038FF3aDAB66BE987")
	LidoContractAddress           ContractAddress = common.HexToAddress("0xae7ab96520de3a18e5e111b5eaab095312d7fe84")
	RocketPoolStorageAddress      ContractAddress = common.HexToAddress("0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46")
	AnkrContractAddress           ContractAddress = common.HexToAddress("0x84db6ee82b7cf3b47e8f19270abde5718b936670")
	RenzoManagerAddress           ContractAddress = common.HexToAddress("0x74a09653A083691711cF8215a6ab074BB4e99ef5")
	AvalonFinanceContractAddress  ContractAddress = common.HexToAddress("0xf9278C7c4AEfAC4dDfd0D496f7a1C39cA6BCA6d4")
	ListaDaoContractAddress       ContractAddress = common.HexToAddress("0x1adB950d8bB3dA4bE104211D5AB038628e477fE6")
)

const (
	erc20BalanceOfABI = `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`
)

// ProtocolConfig contains configuration data for initializing a protocol.
type ProtocolConfig struct {
	RPCURL  string
	ChainID *big.Int

	Name     string
	Version  string
	Contract common.Address
	ABI      abi.ABI
	Type     ProtocolType
}

// TransactionParams encapsulates parameters needed to generate calldata for transactions.
type TransactionParams struct {
	Amount       *big.Int
	Sender       common.Address
	Recipient    common.Address
	Asset        common.Address
	ReferralCode any
	ExtraData    map[string]interface{}
}

func (params TransactionParams) GetBeneficiaryOwner() common.Address {
	if params.Recipient.Hex() == "0x0000000000000000000000000000000000000000" {
		return params.Sender
	}

	return params.Recipient
}

const (
	LoanSupply ContractAction = iota
	LoanWithdraw
	NativeStake
	NativeUnStake
	ERC20Stake
	ERC20UnStake
	LoanBorrow
	LoanRepay
)

func (a ContractAction) String() string {
	switch a {
	case LoanSupply:
		return "loan_supply"
	case LoanWithdraw:
		return "loan_withdraw"
	case NativeStake:
		return "native_stake"
	case NativeUnStake:
		return "native_unstake"
	default:
		return ""
	}
}

const (
	TypeLoan  ProtocolType = "Loan"
	TypeStake ProtocolType = "Stake"
)

// ProtocolRegistry defines methods for managing and accessing DeFi
type ProtocolRegistry interface {
	// GetChainConfig retrieves the configuration for a specific chain
	GetChainConfig(chainID *big.Int) (ChainConfig, error)

	// RegisterProtocol adds a new protocol to the registry for a specific chain
	RegisterProtocol(chainID *big.Int, address common.Address, protocol Protocol) error

	// GetProtocol retrieves a protocol by its contract address and chain ID
	GetProtocol(chainID *big.Int, address common.Address) (Protocol, error)

	// ListProtocols returns a list of all registered protocols for a specific chain
	ListProtocols(chainID *big.Int) []Protocol

	// ListProtocolsByType lists all protocols of a specific type for a given chain
	ListProtocolsByType(chainID *big.Int, protocolType ProtocolType) []Protocol
}

// IsBnb checks if the provided chain matches the BSC chain id
func IsBnb(chainID *big.Int) bool { return chainID.Cmp(BscChainID) == 0 }

// IsEth checks if the provided chain matches the ethereum chain id
func IsEth(chainID *big.Int) bool { return chainID.Cmp(EthChainID) == 0 }

// IsPolygon checks if the provided chain maches polygon id
func IsPolygon(chainID *big.Int) bool { return chainID.Cmp(PolygonChainID) == 0 }
