# Generalized Chain-Aware DeFi Protocol Interface

Here's a comprehensive interface that can adapt to any DeFi protocol, incorporating initialization, transaction generation, validation, and data retrieval.
This interface also includes a method to handle the supported assets, which can vary per chain.

```go
package protocols

// DeFiProtocol defines a generic interface for all types of DeFi protocols.
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

// ProtocolConfig contains configuration data for initializing a protocol.
type ProtocolConfig struct {
    RPCURL   string
    Name     string
    Version  string
    ChainID  *big.Int
    Contract common.Address
    ABI      abi.ABI
    Type     ProtocolType
}

// TransactionParams encapsulates parameters needed to generate calldata for transactions.
type TransactionParams struct {
    FromAddress  common.Address
    ToAddress    common.Address
    AmountIn     *big.Int
    AmountOut    *big.Int
    Sender       common.Address
    Recipient    common.Address
    ReferralCode  any
    ExtraData    map[string]interface{}
}
```

## Implementation Considerations

1. Flexibility: each protocol can perform necessary setup actions in their constructors as they wish, which might
   include establishing network connections, fetching contract addresses, or loading ABIs.

2. Dynamic Interaction: The `GetABI` method allows client code to interact dynamically with the protocol’s contract
   without hardcoding function calls, making the interface adaptable to future changes in the contract.

3. Action and Validation: `GenerateCalldata` and `Validate` take an action and parameters, offering the flexibility to
   support a wide range of operations without changing the interface. This approach accommodates custom actions that may be specific to certain protocols.

4. Supported Assets: By including a method `GetSupportedAssets` to fetch supported assets per chain, `IsSupportedAsset` validates if the asset is supported and balance function `GetBalance` to provide the balance of an address per protocol can provide necessary information for client applications or other services that need to display or utilize asset data dynamically.

## Usage

This interface can be implemented by any DeFi protocol, with each protocol providing its specific logic for handling transactions, initialization, and other actions. When a new protocol is added to your platform, only the specifics of its configuration and operations need to be defined, adhering to the DeFiProtocol interface. This structure greatly simplifies integrating diverse DeFi functionalities into your system while maintaining robustness and scalability.

## Creating New Protocol Instances

Here's how you would create new instances of the `Lido` protocol according to the defined interface.

### Lido Protocol Initialization

```go
package protocols

import (
 "context"
 "encoding/hex"
 "errors"
 "fmt"
 "math/big"
 "strings"

 "github.com/ethereum/go-ethereum/accounts/abi"
 "github.com/ethereum/go-ethereum/common"
)

// LidoABI is the ABI definition for the Lido protocol
const LidoABI = `
[
    {
        "inputs": [
            {
                "internalType": "uint256",
                "name": "amount",
                "type": "uint256"
            }
        ],
        "name": "submit",
        "outputs": [],
        "stateMutability": "payable",
        "type": "function"
    }
]
`
// LidoOperation implements the Protocol interface for Lido
type LidoOperation struct {
    parsedABI abi.ABI
    contract  common.Address
    chainID   *big.Int
    version   string
    // additional fields
}

// Initialize prepares the Lido protocol with necessary configurations and network connections
func (l *LidoOperation) Initialize(ctx context.Context, config ProtocolConfig) error {
    parsedABI, err := abi.JSON(strings.NewReader(LidoABI))
    if err != nil {
    return err
    }

    l.parsedABI = parsedABI
    l.contract = config.Contract
    l.chainID = config.ChainID
    l.name = config.Name
    l.version = config.Version

    // Perform any additional initialization tasks here, if needed

    return nil
}

// GenerateCalldata creates the necessary blockchain transaction data
func (l *LidoOperation) GenerateCalldata(ctx context.Context, chainID *big.Int, action ContractAction, params TransactionParams) (string, error) {
    var calldata []byte
    var err error

    switch action {
    case NativeStake:
    calldata, err = l.parsedABI.Pack("submit", params.AmountIn)
    if err != nil {
    return "", err
    }
    default:
    return "", errors.New("action not supported")
    }

    return "0x" + hex.EncodeToString(calldata), nil
}

// Validate checks if the provided parameters are valid for the specified action
func (l *LidoOperation) Validate(ctx context.Context, chainID *big.Int, action ContractAction, params TransactionParams) error {
    if action == NativeStake {
    if params.AmountIn.Cmp(big.NewInt(0)) <= 0 {
    return errors.New("amount must be greater than zero")
    }
    return nil
    }
    return errors.New("action not supported")
}

// GetBalance retrieves the balance for a specified account and asset
func (l *LidoOperation) GetBalance(ctx context.Context, chainID *big.Int, account, asset common.Address) (*big.Int, error) {
    return nil, errors.New("not implemented")
}

// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain
func (l *LidoOperation) GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error) {
    return []common.Address{}, nil
}

// IsSupportedAsset checks if the specified asset is supported on the given chain
func (l *LidoOperation) IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool {
    return asset == common.HexToAddress("0x0000000000000000000000000000000000000000")
}

// GetProtocolConfig returns the protocol config for a specific chain
func (l *LidoOperation) GetProtocolConfig(chainID *big.Int) ProtocolConfig {
    return ProtocolConfig{
    RPCURL:   "",
    ChainID:  chainID,
    Contract: l.contract,
    ABI:      l.parsedABI,
    Type:     TypeStake,
    }
}

// GetABI returns the ABI of the protocol's contract
func (l *LidoOperation) GetABI(chainID *big.Int) abi.ABI {
    return l.parsedABI
}

// GetType returns the protocol type
func (l *LidoOperation) GetType() ProtocolType {
    return TypeStake
}

// GetContractAddress returns the contract address for a specific chain
func (l *LidoOperation) GetContractAddress(chainID *big.Int) common.Address {
    return l.contract
}

// Name returns the human readable name for the protocol
func (l *LidoOperation) Name() string {
    return l.name
}

// GetVersion returns the version of the protocol
func (l *LidoOperation) GetVersion() string {
    return l.version
}

// GetBeneficiaryOwner determines the ultimate beneficiary owner for the token to be minted.
func (l *LidoOperation) GetBeneficiaryOwner(params TransactionParams) common.Address {
    if params.Recipient.Hex() == "0x0000000000000000000000000000000000000000" {
        return params.Sender
    }
    return params.Recipient
}
```
