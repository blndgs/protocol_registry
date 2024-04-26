package pkg

// Hex prefix
const HexPrefix = "0x"

type ContractAction string

type Protocol struct {
	Name    string
	Address string
	ABI     string
}

const (
	SupplyAction   ContractAction = "supply"
	WithdrawAction ContractAction = "withdraw"
)

var SupportedProtocols = map[string]Protocol{
	"AaveV3": {
		Name:    "AaveV3",
		Address: "0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2",
		ABI:     `[{"name":"supply","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"},{"type":"uint16"}]},{"name":"withdraw","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"}]}]`,
	},
	"SparkLend": {
		Name:    "SparkLend",
		Address: "0xc13e21b648a5ee794902342038ff3adab66be987",
		ABI:     `[{"name":"supply","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"},{"type":"uint16"}]},{"name":"withdraw","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"address"}]}]`,
	},
}
