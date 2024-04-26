# Protocol Registry

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
registry := pkg.NewProtocolRegistry()
```

### Registry new Protocol Operation

To register a new protocol operation, you can use the `RegisterProtocolOperation` function:

```go
registry.RegisterProtocolOperation("AaveV3", pkg.SupplyAction, big.NewInt(1), &pkg.GenericProtocolOperation{
    DynamicOperation: pkg.DynamicOperation{
        Protocol: "AaveV3",
        Action:   pkg.SupplyAction,
        ChainID:  big.NewInt(1),
    },
})
```

### Retrieving Protocol Operations

To retrieve a protocol operation, you can use the `GetProtocolOperation` function:

```go
codeoperation, err := registry.GetProtocolOperation("AaveV3", pkg.SupplyAction, big.NewInt(1))
if err != nil {
    // Handle the error
}
```

### Generating Calldata

To generate calldata for a specific operation, you can use the `GenerateCalldata` method of the retrieved operation:

```go
codecalldata, err := operation.GenerateCalldata()
if err != nil {
    // Handle the error
}
```

## Adding New Protocols and Operations

To add support for a new protocol and its operations, follow these steps:

- Define the protocol details in the `SupportedProtocols` map in the `pkg` package:

```go
var SupportedProtocols = map[string]Protocol{
    "YourProtocol": {
        Name:    "YourProtocol",
        Address: "0x...",
        ABI:     `[...]`,
    },
}
```

- Register the protocol operations using the `RegisterProtocolOperation` function:

```go
coderegistry.RegisterProtocolOperation("YourProtocol", pkg.YourAction, big.NewInt(1), &pkg.GenericProtocolOperation{
    DynamicOperation: pkg.DynamicOperation{
        Protocol: "YourProtocol",
        Action:   pkg.YourAction,
        ChainID:  big.NewInt(1),
    },
})
```

- Implement the necessary logic for generating calldata in the `GenerateCalldata` method of the `GenericProtocolOperation` struct, if required.

## Command-Line Tool

The Protocol Registry Package also includes a command-line tool for demonstration and testing purposes. To use the command-line tool, follow these steps:

- Navigate to the cmd directory:

```sh
cd cmd
```

- Build the command-line tool:

```sh
go build -o protocol main.go
```

- Run the command-line tool with the desired flags:

```sh
code./protocol -protocol AaveV3 -action supply
```

Example output:

```sh
Enter the args for the operation (comma-separated):
0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984,1000000000000000000,0x0000000000000000000000000000000000000000,0
Generated calldata: 0x...
```

For more information on the available flags and usage examples, refer to the command-line tool's help information:

```sh
code./protocol -help
```

## Contributing

Contributions to the Protocol Registry Package are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on the GitHub repository.
