# Protocol Registry

![codecov](https://codecov.io/gh/blndgs/protocol_registry/graph/badge.svg?token=O42114OGRQ)
![Go Release](https://img.shields.io/github/v/release/blndgs/protocol_registry?logo=go)

The Protocol Registry is a Go library that provides a flexible and extensible
way to manage and interact with different protocols and their operations.
It allows you to register protocol operations, retrieve them based on protocol
name and action, and generate calldata for specific operations.

## Features

- Support for multiple protocols and their operations
- Easy registration of new protocols and operations
- Retrieval of protocol operations based on protocol name and action
- Generation of calldata for specific operations
- Extensible design to accommodate new protocols and actions

## Installation

To use the Protocol Registry Package in your Go project, you can install it
using the following command:

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
registry := pkg.NewProtocolRegistry()
```

### Registry new Protocol Operation

To register a new protocol operation, you can use the
`RegisterProtocolOperation` function:

```go
registry.RegisterProtocolOperation(protocol.Address, protocol.ChainID, implementationStruct)
```

### Retrieving Protocol Operations

To retrieve a protocol operation, you can use the `GetProtocolOperation` function:

```go
operation, err := registry.GetProtocolOperation(protocol.Address, big.NewInt(1))
if err != nil {
    // Handle the error
}
```

### Generating Calldata

To generate calldata for a specific operation, you can use the
`GenerateCalldata` method of the retrieved operation:

```go
calldata, err := operation.GenerateCalldata(pkg.LoanKind, GenerateCalldataOptions{
  Asset: common.HexToAddress("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984"),
  Recipient: common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
  Amount:    big.NewInt(1e18),
})
if err != nil {
    // Handle the error
}
```

#### Supported protocols

- Lido
- Aave3
- SparkLend
- Rocketpool
- Ankr

### Adding New Protocols and Operations

To add support for a new protocol and its operations, follow these steps:

- Create a struct that implements the following:

```go
GenerateCalldata(ContractAction, GenerateCalldataOptions) (string, error)
Validate(asset common.Address) error
```

- Register the cutom implementation using the `RegisterProtocolOperation` function:

```go
registry.RegisterProtocolOperation(Protocol", pkg.YourAction, big.NewInt(1), &pkg.GenericProtocolOperation{
    DynamicOperation: pkg.DynamicOperation{
        Protocol: "YourProtocol",
        Action:   pkg.YourAction,
        ChainID:  big.NewInt(1),
    },
})
```

- Implement the necessary logic for generating calldata in the `GenerateCalldata` method of the `GenericProtocolOperation` struct, if required.

## Contributing

Contributions to the Protocol Registry Package are welcome! If you find
any issues or have suggestions for improvements, please open an issue or
submit a pull request on the GitHub repository.
