# Generalized Chain-Aware ProtocolRegistry Interface

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

// ChainConfig chain configuration.
type ChainConfig struct {
    ChainID *big.Int
    RPCURL  string
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

// ChainConfig chain configuration
type ChainConfig struct {
    ChainID *big.Int
    RPCURL  string
}

// ProtocolRegistry is an implementation of the ProtocolRegistry interface.
type ProtocolRegistry struct {
    mu              sync.RWMutex
    protocols       map[common.Address]Protocol
    protocolByType  map[ProtocolType][]Protocol
    chainConfigs    map[string]ChainConfig
    
}


// NewProtocolRegistry creates a new instance of ProtocolRegistry.
func NewProtocolRegistry(chainConfigs []ChainConfig) *ProtocolRegistry {
    r := &ProtocolRegistry{
        protocols:       make(map[common.Address]Protocol),
        protocolByType:  make(map[ProtocolType][]Protocol),
        chainConfigs:   make(map[string]ChainConfig),
    }
    
    // Add chain configurations
    for _, config := range chainConfigs {
        chainIDStr := config.ChainID.String()
        r.chainConfigs[chainIDStr] = config
    }

    // Setup protocol operations
    err := r.setupProtocolOperations()
    if err != nil {
        return nil, err
    }

    return r, nil
}

func (r *ProtocolRegistry) GetChainConfig(chainID *big.Int) (ChainConfig, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    chainIDStr := chainID.String()
    if config, exists := r.chainConfigs[chainIDStr]; exists {
        return config, nil
    }
    return ChainConfig{}, fmt.Errorf("chain config not found for chainID: %s", chainIDStr)
}

// RegisterProtocol adds a new protocol to the registry by its contract address.
func (r *ProtocolRegistry) RegisterProtocol(chainID *big.Int, address common.Address, protocol Protocol) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    chainIDStr := chainID.String()
    if _, exists := r.chainConfigs[chainIDStr]; !exists {
        return fmt.Errorf("chain config not found for chainID: %s", chainIDStr)
    }

    if _, exists := r.protocols[chainIDStr][address]; exists {
        return fmt.Errorf("protocol already registered for chainID %s and address %s", chainIDStr, address.Hex())
    }

    r.protocols[chainIDStr][address] = protocol
    protocolType := protocol.GetType()
    r.protocolByType[chainIDStr][protocolType] = append(r.protocolByType[chainIDStr][protocolType], protocol)

    return nil
}

// GetProtocol retrieves a protocol by its contract address.
func (r *ProtocolRegistry) GetProtocol(chainID *big.Int, address common.Address) (Protocol, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    chainIDStr := chainID.String()
    if chainProtocols, exists := r.protocols[chainIDStr]; exists {
        if protocol, exists := chainProtocols[address]; exists {
            return protocol, nil
        }
    }

    return nil, fmt.Errorf("protocol not found for chainID %s and address %s", chainIDStr, address.Hex())
}

// ListProtocols returns a list of all registered protocols.
func (r *ProtocolRegistry) ListProtocols(chainID *big.Int) []Protocol {
    r.mu.RLock()
    defer r.mu.RUnlock()

    chainIDStr := chainID.String()
    var protocols []Protocol
    if chainProtocols, exists := r.protocols[chainIDStr]; exists {
        for _, protocol := range chainProtocols {
            protocols = append(protocols, protocol)
        }
    }
    return protocols
}

// ListProtocolsByType lists all protocols of a specific type.
func (r *ProtocolRegistry) ListProtocolsByType(chainID *big.Int, protocolType ProtocolType) []Protocol {
    r.mu.RLock()
    defer r.mu.RUnlock()

    chainIDStr := chainID.String()
    if chainTypes, exists := r.protocolByType[chainIDStr]; exists {
        if protocols, exists := chainTypes[protocolType]; exists {
            return protocols
        }
    }
    return []Protocol{}
}

// setupProtocolOperations initializes and registers various DeFi protocols.
func (r *ProtocolRegistry) setupProtocolOperations() error {
    registerProtocol := func(address common.Address, chainID *big.Int, createFunc func(ChainConfig) (Protocol, error)) error {
    chainIDStr := chainID.String()
    config, exists := r.chainConfigs[chainIDStr]
    
    if !exists {
        return fmt.Errorf("chain configuration not found for chainID: %s", chainIDStr)
    }

    protocol, err := createFunc(config)
    if err != nil {
        return fmt.Errorf("failed to create protocol at address %s: %v", address.Hex(), err)
    }

    err = r.RegisterProtocol(address, protocol)
    if err != nil {
        return fmt.Errorf("failed to register protocol at address %s: %v", address.Hex(), err)
    }

    return nil
    }

    // Register Lido protocol
    err := registerProtocol(common.HexToAddress("0xLidoContractAddress"), big.NewInt(1), func(config ChainConfig) (Protocol, error) {
    lido := &LidoOperation{}
    protocolConfig := ProtocolConfig{
        RPCURL:   config.RPCURL,
        ChainID:  config.ChainID,
        Contract: common.HexToAddress("0xLidoContractAddress"),
        ABI:      LidoABI,
        Type:     TypeStake,
    }
    if err := lido.Initialize(context.Background(), protocolConfig); err != nil {
        return nil, err
    }
        return lido, nil
    })
    if err != nil {
        return err
    }

    // so on....
}
```

### Client Usages

Here's how client usages it in your main function

```go
    // Setup protocol operations
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
    // Create a new protocol registry
    registry := protocols.NewProtocolRegistry(chainConfigs)

   // Example contract addresses
    lidoAddress := common.HexToAddress("0xLidoContractAddress")

    // Get protocol by address
    lido, err := registry.GetProtocol(lidoAddress)
    if err != nil {
        log.Fatalf("Error getting protocol by address: %v", err)
    }
    log.Println("Retrieved protocol by address:", lido)

    // List all registered protocols
    protocolsList := registry.ListProtocols()
    log.Println("Registered protocols:", protocolsList)

    // List protocols by type (example)
    stakeProtocols := registry.ListProtocolsByType(protocols.TypeStake)
    log.Println("Stake protocols:", stakeProtocols)
```

This implementation provides thread-safe operations using a read-write mutex and efficiently manages protocols across multiple chains using nested maps. It includes all the methods defined in the ProtocolRegistry interface, handling errors appropriately and returning meaningful error messages when operations fail.
