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
	protocols map[string]map[ContractAction]map[int64]ProtocolOperation
}

// NewProtocolRegistry creates a new instance of ProtocolRegistry.
func NewProtocolRegistry() *ProtocolRegistry {
	return &ProtocolRegistry{
		protocols: make(map[string]map[ContractAction]map[int64]ProtocolOperation),
	}
}

// RegisterProtocolOperation registers a new operation for a protocol on a specific chain.
func (pr *ProtocolRegistry) RegisterProtocolOperation(protocol string, action ContractAction, chainID *big.Int, operation ProtocolOperation) {
	pr.lock.Lock()
	defer pr.lock.Unlock()

	// Validate protocol - it must be supported
	if _, ok := SupportedProtocols[protocol]; !ok {
		panic("unsupported protocol: " + protocol)
	}

	// Validate chainID - it should not be nil or negative
	if chainID == nil || chainID.Sign() != 1 {
		panic("invalid chain ID: " + chainID.String())
	}

	// Validate operation - it must not be nil
	if operation == nil {
		panic("nil operation not allowed")
	}

	if pr.protocols[protocol] == nil {
		pr.protocols[protocol] = make(map[ContractAction]map[int64]ProtocolOperation)
	}
	if pr.protocols[protocol][action] == nil {
		pr.protocols[protocol][action] = make(map[int64]ProtocolOperation)
	}
	pr.protocols[protocol][action][chainID.Int64()] = operation
}

// GetProtocolOperation retrieves an operation for a given protocol and chain.
func (pr *ProtocolRegistry) GetProtocolOperation(protocol string, action ContractAction, chainID *big.Int) (ProtocolOperation, error) {
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
func SetupProtocolOperations(registry *ProtocolRegistry) {
	for name, details := range SupportedProtocols {
		parsedABI, err := abi.JSON(strings.NewReader(details.ABI))
		if err != nil {
			panic(fmt.Sprintf("failed to parse ABI for %s: %v", name, err))
		}
		updatedProtocol := details
		updatedProtocol.ParsedABI = parsedABI
		SupportedProtocols[name] = updatedProtocol

		for _, method := range parsedABI.Methods {
			action := ContractAction(method.Name)
			registry.RegisterProtocolOperation(name, action, details.ChainID, &GenericProtocolOperation{
				DynamicOperation: DynamicOperation{
					Protocol: name,
					Action:   action,
					ChainID:  details.ChainID,
				},
			})
		}
	}
}
