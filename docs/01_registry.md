# Generalized ProtocolRegistry Interface

The `ProtocolRegistry` interface will define the necessary operations to manage DeFi protocols within the system:

```go
package protocols

const (
 LoanSupply ContractAction = iota
 LoanWithdraw
 NativeStake
 NativeUnStake
 ERC20Stake
 ERC20UnStake
)

const (
 TypeLoan  ProtocolType = "Loan"
 TypeStake ProtocolType = "Stake"
)

// ProtocolRegistry defines methods for managing and accessing DeFi 
type ProtocolRegistry interface {
    // RegisterProtocol adds a new protocol to the registry.
    RegisterProtocol(name string, protocol Protocol) error

    // GetProtocol retrieves a protocol by name.
    GetProtocol(name string) (Protocol, error)

    // ListProtocols returns a list of all registered 
    ListProtocols() []string

    // ListProtocolsByType lists all protocols of a specific type.
    ListProtocolsByType(protocolType ProtocolType) []string
}
```

## Implementation Considerations

1. Flexibility: By providing an `Initialize` method, each protocol can perform necessary setup actions, which might include establishing network connections, fetching contract addresses, or loading ABIs.

2. Dynamic Interaction: The `GetABI` method allows client code to interact dynamically with the protocol’s contract without hardcoding function calls, making the interface adaptable to future changes in the contract.

3. Action and Validation: `GenerateCalldata` and `Validate` take an action and parameters, offering the flexibility to support a wide range of operations without changing the interface. This approach accommodates custom actions that may be specific to certain.

4. Supported Assets: By including a method `GetSupportedAssets` to fetch supported assets per chain, and balance function `GetBalance` to provide the balance of an address per protocol can provide necessary information for front-end applications or other services that need to display or utilize asset data dynamically.

## Usage

This interface can be implemented by any DeFi protocol, with each protocol providing its specific logic for handling transactions, initialization, and other actions. When a new protocol is added to your platform, only the specifics of its configuration and operations need to be defined, adhering to the DeFiProtocol interface. This structure greatly simplifies integrating diverse DeFi functionalities into your system while maintaining robustness and scalability.

### Implementation

```go
package protocols

import (
 "errors"
 "sync"
)

// ProtocolRegistry is a basic implementation of the ProtocolRegistry interface.
type ProtocolRegistry struct {
 mu         sync.RWMutex
 protocols  map[string]Protocol
 protocolByType map[ProtocolType][]string
}

// NewProtocolRegistry creates a new instance of ProtocolRegistry.
func NewProtocolRegistry() *ProtocolRegistry {
 return &ProtocolRegistry{
  protocols: make(map[string]Protocol),
  protocolByType: make(map[ProtocolType][]string),
 }
}

// RegisterProtocol adds a new protocol to the registry.
func (r *ProtocolRegistry) RegisterProtocol(name string, protocol Protocol) error {
 r.mu.Lock()
 defer r.mu.Unlock()

 if _, exists := r.protocols[name]; exists {
  return errors.New("protocol already registered")
 }

 r.protocols[name] = protocol
 protocolType := protocol.GetType()
 r.protocolByType[protocolType] = append(r.protocolByType[protocolType], name)
 return nil
}

// GetProtocol retrieves a protocol by name.
func (r *ProtocolRegistry) GetProtocol(name string) (Protocol, error) {
 r.mu.RLock()
 defer r.mu.RUnlock()

 protocol, exists := r.protocols[name]
 if !exists {
  return nil, errors.New("protocol not found")
 }
 return protocol, nil
}

// ListProtocols returns a list of all registered 
func (r *ProtocolRegistry) ListProtocols() []string {
 r.mu.RLock()
 defer r.mu.RUnlock()

 names := make([]string, 0, len(r.protocols))
 for name := range r.protocols {
  names = append(names, name)
 }
 return names
}

// ListProtocolsByType lists all protocols of a specific type.
func (r *ProtocolRegistry) ListProtocolsByType(protocolType ProtocolType) []string {
 r.mu.RLock()
 defer r.mu.RUnlock()

 return r.protocolByType[protocolType]
}

func setupProtocolOperations(registry ProtocolRegistry, rpcURL string) {
    // Define a helper function to handle protocol creation and registration
    registerProtocol := func(name string, createFunc func() (Protocol, error)) {
        protocol, err := createFunc()
        if err != nil {
            log.Fatalf("Failed to create %s operation: %v", name, err)
        }

        err = registry.RegisterProtocol(name, protocol)
        if err != nil {
            log.Fatalf("Failed to register %s protocol: %v", name, err)
        }
    }

    // Register Lido protocol
    registerProtocol("Lido", func() (Protocol, error) {
        lido := &LidoOperation{}
        config := ProtocolConfig{
            RPCURL:   rpcURL,
            ChainID:  big.NewInt(1), // Mainnet
            Contract: common.HexToAddress("0xYourLidoContractAddress"),
            ABI:      abi.ABI{}, // Set appropriate ABI
            Type:     TypeStake,
        }
        if err := lido.Initialize(context.Background(), config); err != nil {
            return nil, err
        }
        return lido, nil
    })

    // Register Ankr protocol
    registerProtocol("Ankr", func() (Protocol, error) {
        ankr := &AnkrOperation{}
        config := ProtocolConfig{
            RPCURL:   rpcURL,
            ChainID:  big.NewInt(1), // Example chain ID
            Contract: common.HexToAddress("0xYourAnkrContractAddress"),
            ABI:      abi.ABI{}, // Set appropriate ABI
            Type:     TypeStake,
        }
        if err := ankr.Initialize(context.Background(), config); err != nil {
            return nil, err
        }
        return ankr, nil
    })

    // Register Aave protocol (AaveV3)
    registerProtocol("AaveV3", func() (Protocol, error) {
        return NewAaveOperation(AaveProtocolForkAave)
    })

    // Register SparkLend protocol
    registerProtocol("SparkLend", func() (Protocol, error) {
        return NewAaveOperation(AaveProtocolForkSpark)
    })

    // Register RocketPool protocol
    registerProtocol("RocketPool", func() (Protocol, error) {
        return NewRocketPool(rpcURL)
    })

    // Optionally register additional protocols
    // registerCompoundRegistry(registry)
}
```

### Client Usages

Here's how client usages it in your main function

```go
 // Create a new protocol registry
    registry := protocols.NewSimpleProtocolRegistry()

    // Setup protocol operations
    rpcURL := "https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID"
    // List all registered protocols
    protocolsList := registry.ListProtocols()
    log.Println("Registered protocols:", protocolsList)

    // List protocols by type (example)
    stakeProtocols := registry.ListProtocolsByType(protocols.TypeStake)
    log.Println("Stake protocols:", stakeProtocols)
```
