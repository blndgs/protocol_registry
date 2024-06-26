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
	protocols map[ContractAddress]map[ContractAction]map[int64]ProtocolOperation
}

// NewProtocolRegistry creates a new instance of ProtocolRegistry.
func NewProtocolRegistry() *ProtocolRegistry {
	return &ProtocolRegistry{
		protocols: make(map[ContractAddress]map[ContractAction]map[int64]ProtocolOperation),
	}
}

// RegisterProtocolOperation registers a new operation for a protocol on a specific chain.
func (pr *ProtocolRegistry) RegisterProtocolOperation(protocol ContractAddress, action ContractAction, chainID *big.Int, operation ProtocolOperation) {
	pr.lock.Lock()
	defer pr.lock.Unlock()

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

// GetProtocolOperation retrieves an operation for a given protocol and chain.
func (pr *ProtocolRegistry) GetProtocolOperation(protocol ContractAddress, action ContractAction, chainID *big.Int) (ProtocolOperation, error) {
	pr.lock.RLock()
	defer pr.lock.RUnlock()

	if ops, exists := pr.protocols[protocol]; exists {
		if actionOps, ok := ops[action]; ok {
			if op, operationExists := actionOps[chainID.Int64()]; operationExists {
				return op, nil
			}
		}
	}

	return nil, fmt.Errorf("operation not found for action %d on chain %d for protocol %s", action, chainID, protocol)
}

// SetupProtocolOperations automatically sets up protocol operations based on the SupportedProtocols map.
func SetupProtocolOperations(rpcURL string, registry *ProtocolRegistry) {

	for protocolType, protocols := range staticProtocols {
		for i, protocol := range protocols {
			parsedABI, err := abi.JSON(strings.NewReader(protocol.ABI))
			if err != nil {
				panic(fmt.Sprintf("Failed to parse ABI for %s: %v", protocol.Name, err))
			}

			// Correctly updating the protocol entry with parsed ABI
			protocol.ParsedABI = parsedABI
			staticProtocols[protocolType][i] = protocol

			// Register each action of the protocol in the registry
			registry.RegisterProtocolOperation(protocol.Address, protocol.Action, protocol.ChainID, &GenericProtocolOperation{
				DynamicOperation: DynamicOperation{
					Protocol: protocol.Name,
					Method:   protocol.Method,
					ChainID:  protocol.ChainID,
					Address:  protocol.Address,
				},
			})
		}
	}

	rocketPoolSubmit, err := NewRocketPool(rpcURL, RocketPoolStorageAddress, NativeStake, rocketPoolStake)
	if err != nil {
		panic(fmt.Sprintf("Failed to create RocketPool submit operation: %v", err))
	}
	rocketPoolSubmit.Register(registry)

	rocketPoolWithdraw, err := NewRocketPool(rpcURL, RocketPoolStorageAddress, NativeUnStake, rocketPoolUnStake)
	if err != nil {
		panic(fmt.Sprintf("Failed to create RocketPool withdraw operation: %v", err))
	}
	rocketPoolWithdraw.Register(registry)

	registerCompoundRegistry(registry)
}
