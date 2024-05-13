package pkg

import (
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// ProtocolRegistry maintains a registry of supported protocols and their operations.
type ProtocolRegistry struct {
	lock      sync.RWMutex
	protocols map[ProtocolName]map[ContractAction]map[int64]ProtocolOperation
}

// NewProtocolRegistry creates a new instance of ProtocolRegistry.
func NewProtocolRegistry() *ProtocolRegistry {
	return &ProtocolRegistry{
		protocols: make(map[ProtocolName]map[ContractAction]map[int64]ProtocolOperation),
	}
}

// RegisterProtocolOperation registers a new operation for a protocol on a specific chain.
func (pr *ProtocolRegistry) RegisterProtocolOperation(protocol ProtocolName, action ContractAction, chainID *big.Int, operation ProtocolOperation) {
	pr.lock.Lock()
	defer pr.lock.Unlock()
	// Check if protocol is supported
	if !isProtocolSupported(protocol, action) {
		panic(fmt.Sprintf("unsupported protocol: %s", protocol))
	}

	// Check chainID validity
	if chainID == nil || chainID.Sign() != 1 {
		panic(fmt.Sprintf("invalid chain ID: %s", chainID))
	}

	// Check if operation is non-nil
	if operation == nil {
		panic("nil operation not allowed")
	}

	// Initialize maps if necessary
	if pr.protocols[protocol] == nil {
		pr.protocols[protocol] = make(map[ContractAction]map[int64]ProtocolOperation)
	}
	if pr.protocols[protocol][action] == nil {
		pr.protocols[protocol][action] = make(map[int64]ProtocolOperation)
	}
	pr.protocols[protocol][action][chainID.Int64()] = operation
}

// isProtocolSupported checks if the protocol and action are defined in SupportedProtocols
func isProtocolSupported(protocol ProtocolName, action ContractAction) bool {
	for _, protocols := range SupportedProtocols {
		for _, proto := range protocols {
			if proto.Name == protocol && proto.Action == action {
				return true
			}
		}
	}
	return false
}

// GetProtocolOperation retrieves an operation for a given protocol and chain.
func (pr *ProtocolRegistry) GetProtocolOperation(protocol ProtocolName, action ContractAction, chainID *big.Int) (ProtocolOperation, error) {
	pr.lock.RLock()
	defer pr.lock.RUnlock()

	if ops, exists := pr.protocols[protocol]; exists {
		if actionOps, ok := ops[action]; ok {
			if op, operationExists := actionOps[chainID.Int64()]; operationExists {
				return op, nil
			}
		}
	}

	return nil, fmt.Errorf("operation not found for action %s on chain %d for protocol %s", action, chainID, protocol)
}

// SetupProtocolOperations automatically sets up protocol operations based on the SupportedProtocols map.
func SetupProtocolOperations(rpcURL string, registry *ProtocolRegistry) {
	for assetKind, protocols := range SupportedProtocols {
		for i, protocol := range protocols {
			parsedABI, err := abi.JSON(strings.NewReader(protocol.ABI))
			if err != nil {
				panic(fmt.Sprintf("Failed to parse ABI for %s: %v", protocol.Name, err))
			}

			// Correctly updating the protocol entry with parsed ABI
			protocol.ParsedABI = parsedABI
			SupportedProtocols[assetKind][i] = protocol

			// Register each action of the protocol in the registry
			registry.RegisterProtocolOperation(protocol.Name, protocol.Action, protocol.ChainID, &GenericProtocolOperation{
				DynamicOperation: DynamicOperation{
					Protocol: protocol.Name,
					Action:   protocol.Action,
					ChainID:  protocol.ChainID,
				},
			})
		}
	}
	rocketPoolSubmit, err := NewRocketPool(rpcURL, RocketPoolStorageAddress, SubmitAction)
	if err != nil {
		panic(fmt.Sprintf("Failed to create RocketPool submit operation: %v", err))
	}
	rocketPoolSubmit.Register(registry)

	rocketPoolWithdraw, err := NewRocketPool(rpcURL, RocketPoolStorageAddress, WithdrawAction)
	if err != nil {
		panic(fmt.Sprintf("Failed to create RocketPool withdraw operation: %v", err))
	}
	rocketPoolWithdraw.Register(registry)
}
