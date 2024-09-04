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
	EthChainStr = "1"
	BscChainStr = "56"
)

var (
	EthChainID = big.NewInt(1)
	BscChainID = big.NewInt(56)
)

// Hex prefix
const HexPrefix = "0x"

var ErrChainUnsupported = errors.New("chain not supported")

const (
	erc20BalanceOfABI = `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`
)

type (
	ProtocolName    = string
	ProtocolMethod  = string
	ContractAddress = common.Address
)

type ContractAction int64

type ProtocolType string

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

const (
	TypeLoan  ProtocolType = "Loan"
	TypeStake ProtocolType = "Stake"
)

var (
	AaveV3ContractAddress        ContractAddress = common.HexToAddress("0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2")
	AaveBnbV3ContractAddress     ContractAddress = common.HexToAddress("0x6807dc923806fE8Fd134338EABCA509979a7e0cB")
	SparkLendContractAddress     ContractAddress = common.HexToAddress("0xC13e21B648A5Ee794902342038FF3aDAB66BE987")
	LidoContractAddress          ContractAddress = common.HexToAddress("0xae7ab96520de3a18e5e111b5eaab095312d7fe84")
	RocketPoolStorageAddress     ContractAddress = common.HexToAddress("0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46")
	AnkrContractAddress          ContractAddress = common.HexToAddress("0x84db6ee82b7cf3b47e8f19270abde5718b936670")
	RenzoManagerAddress          ContractAddress = common.HexToAddress("0x74a09653A083691711cF8215a6ab074BB4e99ef5")
	AvalonFinanceContractAddress ContractAddress = common.HexToAddress("0xf9278C7c4AEfAC4dDfd0D496f7a1C39cA6BCA6d4")
	ListaDaoContractAddress      ContractAddress = common.HexToAddress("0x1adB950d8bB3dA4bE104211D5AB038628e477fE6")
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
type Protocol interface {
	// GenerateCalldata creates the necessary blockchain transaction data.
	GenerateCalldata(ctx context.Context, chainID *big.Int, action ContractAction, params TransactionParams) (string, error)

	// Validate checks if the provided parameters are valid for the specified action.
	Validate(ctx context.Context, chainID *big.Int, action ContractAction, params TransactionParams) error

	// GetBalance retrieves the balance for a specified account and asset.
	GetBalance(ctx context.Context, chainID *big.Int, account, asset common.Address) (*big.Int, error)

	// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain.
	GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error)

	// IsSupportedAsset checks if the specified asset is supported on the given chain.
	IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool

	// GetProtocolConfig returns the protocol config for a specific chain.
	GetProtocolConfig(chainID *big.Int) ProtocolConfig

	// GetABI returns the ABI of the protocol's contract, allowing dynamic interaction.
	GetABI(chainID *big.Int) abi.ABI

	// GetType returns the protocol type.
	GetType() ProtocolType

	// GetName returns the human-readable name of the protocol.
	GetName() string

	// GetVersion returns the version of the protocol.
	GetVersion() string

	// GetContractAddress returns the contract address for a specific chain.
	GetContractAddress(chainID *big.Int) common.Address
}

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

// GetBeneficiaryOwner return the beneficiary address.
func (params TransactionParams) GetBeneficiaryOwner() common.Address {
	if params.Recipient.Hex() == "0x0000000000000000000000000000000000000000" {
		return params.Sender
	}

	return params.Recipient
}

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
