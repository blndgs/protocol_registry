# Protocol Registry

![codecov](https://codecov.io/gh/blndgs/protocol_registry/graph/badge.svg?token=O42114OGRQ)
![Go Release](https://img.shields.io/github/v/release/blndgs/protocol_registry?logo=go)

The Protocol Registry is a Go library that provides a flexible and extensible way to manage and interact with different protocols and their operations. It allows you to register protocol operations, retrieve them based on protocol name and action, and generate calldata for specific operations.

## Features

- Support for multiple protocols and their operations
- Easy registration of new protocols and operations
- Retrieval of protocol operations based on protocol name and action
- Generation of calldata for specific operations
- Extensible design to accommodate new protocols and actions

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
    FromAddress: common.HexToAddress("0xAddress1"),
    ToAddress:   common.HexToAddress("0xAddress2"),
    AmountIn:    big.NewInt(1000000000000),
    // Set other necessary parameters
}
calldata, err := protocol.GenerateCalldata(context.Background(), big.NewInt(1), protocols.NativeStake, params)
if err != nil {
    // Handle the error
}
```

## Protocol Interface

The `Protocol` interface defines the methods that each protocol must implement:

```go
type Protocol interface {
    Initialize(ctx context.Context, config ProtocolConfig) error
    GenerateCalldata(ctx context.Context, chainID *big.Int, action ContractAction, params TransactionParams) (string, error)
    Validate(ctx context.Context, chainID *big.Int, action ContractAction, params TransactionParams) error
    GetBalance(ctx context.Context, chainID *big.Int, account, asset common.Address) (*big.Int, error)
    GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error)
    IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool
    GetProtocolConfig(chainID *big.Int) ProtocolConfig
    GetABI(chainID *big.Int) abi.ABI
    GetType() ProtocolType
    GetName() string
    GetVersion() string
    GetContractAddress(chainID *big.Int) common.Address
    GetBeneficiaryOwner(params TransactionParams) common.Address
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

## Contributing

Contributions to the Protocol Registry Package are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the GitHub repository.

## License

This project is licensed under the terms of the license file in the root directory. See the [LICENSE](./LICENSE) file for details.
