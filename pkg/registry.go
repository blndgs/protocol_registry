package pkg

import (
	"fmt"
	"math/big"
	"sync"
)

// ProtocolRegistry maintains a registry of supported protocols and their operations.
type ProtocolRegistry struct {
	lock      sync.RWMutex
	protocols map[ContractAddress]map[int64]ProtocolOperation
}

// NewProtocolRegistry creates a new instance of ProtocolRegistry.
func NewProtocolRegistry(rpcURL string) *ProtocolRegistry {
	r := &ProtocolRegistry{
		protocols: make(map[ContractAddress]map[int64]ProtocolOperation),
	}

	r.setupProtocolOperations(rpcURL)
	return r
}

// RegisterProtocolOperation registers a new operation for a protocol on a specific chain.
func (pr *ProtocolRegistry) RegisterProtocolOperation(protocol ContractAddress,
	chainID *big.Int, operation ProtocolOperation) {

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
		pr.protocols[protocol] = make(map[int64]ProtocolOperation)
	}

	pr.protocols[protocol][chainID.Int64()] = operation
}

// GetProtocolOperation retrieves an operation for a given protocol and chain.
func (pr *ProtocolRegistry) GetProtocolOperation(protocol ContractAddress, chainID *big.Int) (ProtocolOperation, error) {
	pr.lock.RLock()
	defer pr.lock.RUnlock()

	if ops, exists := pr.protocols[protocol]; exists {
		if op, ok := ops[chainID.Int64()]; ok {
			return op, nil
		}
	}

	return nil, fmt.Errorf("operation not found for chain %d for protocol %s", chainID, protocol)
}

// SetupProtocolOperations automatically sets up protocol operations based on the SupportedProtocols map.
func (pr *ProtocolRegistry) setupProtocolOperations(rpcURL string) {

	lido, err := NewLidoOperation()
	if err != nil {
		panic(fmt.Sprintf("Failed to create Aave operation: %v", err))
	}

	lido.Register(pr)

	ankr, err := NewAnkrOperation()
	if err != nil {
		panic(fmt.Sprintf("Failed to create Aave operation: %v", err))
	}

	ankr.Register(pr)

	aaveOperation, err := NewAaveOperation()
	if err != nil {
		panic(fmt.Sprintf("Failed to create Aave operation: %v", err))
	}

	aaveOperation.Register(pr, AaveV3ContractAddress)
	aaveOperation.Register(pr, SparkLendContractAddress)

	rocketPool, err := NewRocketPool(rpcURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to create RocketPool submit operation: %v", err))
	}
	rocketPool.Register(pr)
}
