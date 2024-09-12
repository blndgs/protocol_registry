# Protocol Registry

![codecov](https://codecov.io/gh/blndgs/protocol_registry/graph/badge.svg?token=O42114OGRQ)
![version](https://img.shields.io/github/v/release/blndgs/protocol_registry?logo=go)
[![release](https://github.com/blndgs/protocol_registry/actions/workflows/release.yml/badge.svg)](https://github.com/blndgs/protocol_registry/actions/workflows/release.yml)

The Protocol Registry is a Go library that provides a flexible and extensible way to manage and interact with different protocols and their operations. It allows you to register protocol operations, retrieve them based on protocol name and action, and generate calldata for specific operations.

## Features

- Support for multiple protocols and their operations
- Easy registration of new protocols and operations
- Retrieval of protocol operations based on protocol name and action
- Generation of calldata for specific operations
- Extensible design to accommodate new protocols and actions
- Whitelisted token support for each blockchain network

## Installation

To use the Protocol Registry Package in your Go project, you can install it using the following command:

```sh
go get github.com/blndgs/protocol_registry
```

## Usage

### Importing the Package

```sh
import "github.com/blndgs/protocol_registry/pkg"
```

## Create a Protocol Registry

```go
    chainConfigs := []protocols.ChainConfig{
        {
            ChainID: big.NewInt(1),
            RPCURL:  "https://mainnet.infura.io/v3/YOUR-PROJECT-ID",
        },
        {
            ChainID: big.NewInt(56),
            RPCURL:  "https://bsc-dataseed.binance.org/",
        },
    }
    registry, err := protocols.NewProtocolRegistry(chainConfigs)
    if err != nil {
        log.Fatalf("Failed to create protocol registry: %v", err)
    }
```

### Registry new Protocol Operation

To register a new protocol operation, you can use the `RegisterProtocol` function:

```go
    err := registry.RegisterProtocol(big.NewInt(1), common.HexToAddress("0xProtocolAddress"), protocolInstance)
    if err != nil {
        log.Fatalf("Failed to register protocol: %v", err)
    }
```

### Retrieving Protocol Operations

To retrieve a protocol operation, you can use the `GetProtocol` function:

```go
protocol, err := registry.GetProtocol(big.NewInt(1), common.HexToAddress("0xProtocolAddress"))
if err != nil {
    // Handle the error
}
```

### Generating Calldata

To generate calldata for a specific operation, you can use the `GenerateCalldata` method of the retrieved operation:

```go
params := protocols.TransactionParams{
    Asset: common.HexToAddress("0xAddress1"),
    Sender:   common.HexToAddress("0xAddress2"),
    Amount:    big.NewInt(1000000000000),
}
calldata, err := protocol.GenerateCalldata(context.Background(), big.NewInt(1), pkg.NativeStake, params)
if err != nil {
    // Handle the error
}
```

## Supported protocols

- Aave V3 ( BSC and ETH )
- Sparklend ( ETH )
- Compound ( ETH )
- Avalon Finance ( BSC )
- Rocketpool ( ETH )
- Lido ( ETH )
- ListaDao ( BSC )
- Ankr ( ETH )

## Protocol Interface

The `Protocol` interface defines the methods that each protocol must implement:

```go
type Protocol interface {
    GenerateCalldata(ctx context.Context, chainID *big.Int, action ContractAction, params TransactionParams) (string, error)
    // Validate ideally should run checks for the balance
    // Currently for LoanSuplly usecase, we do not run any checks
    // because sometimes on the clientside, the usecase might be a multicall that swaps an asset
    // for another one which is then supplied into the protocol hence validation will always fail
    Validate(ctx context.Context, chainID *big.Int, action ContractAction, params TransactionParams) error
    GetBalance(ctx context.Context, chainID *big.Int, account, asset common.Address) (common.Address,*big.Int, error)
    GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error)
    IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool
    GetProtocolConfig(chainID *big.Int) ProtocolConfig
    GetABI(chainID *big.Int) abi.ABI
    GetType() ProtocolType
    GetName() string
    GetVersion() string
    GetContractAddress(chainID *big.Int) common.Address
}
```

For more details on the [Protocol interface and its implementation](./docs/00_protocol.md), refer to the Protocol documentation.

## Registry Interface

The `ProtocolRegistry` interface defines the methods for managing and accessing DeFi protocols:

```go
type ProtocolRegistry interface {
    GetChainConfig(chainID *big.Int) (ChainConfig, error)
    RegisterProtocol(chainID *big.Int, address common.Address, protocol Protocol) error
    GetProtocol(chainID *big.Int, address common.Address) (Protocol, error)
    ListProtocols(chainID *big.Int) []Protocol
    ListProtocolsByType(chainID *big.Int, protocolType ProtocolType) []Protocol
}
```

For more details on the [ProtocolRegistry interface and its implementation](./docs/01_registry.md), refer to the Registry documentation.

## Working with Whitelisted Tokens

The Protocol Registry supports whitelisted tokens for each blockchain network. These tokens are defined in JSON files named after their respective chain IDs (e.g., 1.json for Ethereum mainnet).
These tokens are managed through the `github.com/blndgs/protocol_registry/tokens` package.

For more information on the whitelisted token standard and management, please refer to the [Whitelisted Token documentation](./tokens/README.md) and the [specifications](./docs/02_token.md).

## Using the Token Registry

To use the Token Registry in your application:

### Import the package

```go
import "github.com/blndgs/protocol_registry/tokens"
```

### Create a new JSONTokenRegistry

```go
registry, err := tokens.NewJSONTokenRegistry()
if err != nil {
    log.Fatalf("Failed to create token registry: %v", err)
}
```

### Use the registry methods to access token and protocol data

```go
// Get all tokens for a specific chain
ethTokens, err := registry.GetTokens(pkg.EthChainID)

// Get a specific token by address
token, err := registry.GetTokenByAddress(pkg.EthChainID, "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")

// Get all protocols for a specific chain
ethProtocols, err := registry.GetProtocols(pkg.EthChainID)

// Get a specific protocol by address
protocol, err := registry.GetProtocolByAddress(pkg.EthChainID, "0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2")
```

The Token Registry automatically loads data from JSON files named after their respective chain IDs (e.g., 1.json for Ethereum mainnet, 56.json for Binance Smart Chain) located in the same directory as the executable.

For more detailed information on the Token Registry and its implementation, please refer to the Token Registry documentation.

## Contributing

Contributions to the Protocol Registry Package are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the GitHub repository.

## License

This project is licensed under the terms of the license file in the root directory. See the [LICENSE](./LICENSE) file for details.
