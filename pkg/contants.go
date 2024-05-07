package pkg

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// Hex prefix
const HexPrefix = "0x"

type ContractAction string

type Protocol struct {
	ChainID   *big.Int
	Name      string
	Address   string
	ABI       string
	ParsedABI abi.ABI
}

const (
	SupplyAction   ContractAction = "supply"
	WithdrawAction ContractAction = "withdraw"
	StakingAction  ContractAction = "submit"
)

// Predefined protocols
var SupportedProtocols = map[string]Protocol{
	"AaveV3": {
		ChainID: big.NewInt(1),
		Name:    "AaveV3",
		Address: "0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2",
		ABI:     `[{"name":"supply","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"},{"type":"uint16"}]},{"name":"withdraw","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"}]}]`,
	},
	"SparkLend": {
		ChainID: big.NewInt(1),
		Name:    "SparkLend",
		Address: "0xc13e21b648a5ee794902342038ff3adab66be987",
		ABI:     `[{"name":"supply","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"},{"type":"uint16"}]},{"name":"withdraw","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"}]}]`,
	},
	"Lido": {
		ChainID: big.NewInt(1),
		Name:    "Lido",
		Address: "0xae7ab96520de3a18e5e111b5eaab095312d7fe84",
		ABI:     `[{"name": "submit", "type": "function","inputs": [{"type": "address"}]}]`,
	},
}
