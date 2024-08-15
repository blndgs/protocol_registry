package pkg

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ChainConfig chain configuration
type ChainConfig struct {
	ChainID *big.Int
	RPCURL  string
}

// ProtocolRegistryImpl is an implementation of the ProtocolRegistryImpl interface.
type ProtocolRegistryImpl struct {
	mu             sync.RWMutex
	protocols      map[string]map[string]Protocol
	protocolByType map[string]map[ProtocolType][]Protocol
	chainConfigs   map[string]ChainConfig
}

// NewProtocolRegistryImpl creates a new instance of ProtocolRegistryImpl.
func NewProtocolRegistry(chainConfigs []ChainConfig) (*ProtocolRegistryImpl, error) {
	r := &ProtocolRegistryImpl{
		protocols:      make(map[string]map[string]Protocol),
		protocolByType: make(map[string]map[ProtocolType][]Protocol),
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

func (r *ProtocolRegistryImpl) GetChainConfig(chainID *big.Int) (ChainConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chainIDStr := chainID.String()
	if config, exists := r.chainConfigs[chainIDStr]; exists {
		return config, nil
	}
	return ChainConfig{}, fmt.Errorf("chain config not found for chainID: %s", chainIDStr)
}

// RegisterProtocol adds a new protocol to the registry by its contract address.
func (r *ProtocolRegistryImpl) RegisterProtocol(chainID *big.Int, address common.Address, protocol Protocol) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	chainIDStr := chainID.String()
	if _, exists := r.chainConfigs[chainIDStr]; !exists {
		return fmt.Errorf("chain config not found for chainID: %s", chainIDStr)
	}

	if _, exists := r.protocolByType[chainIDStr]; !exists {
		r.protocolByType[chainIDStr] = make(map[ProtocolType][]Protocol)
	}

	if _, exists := r.protocols[chainIDStr]; !exists {
		r.protocols[chainIDStr] = make(map[string]Protocol)
	}

	if _, exists := r.protocols[chainIDStr][address.Hex()]; exists {
		return fmt.Errorf("protocol already registered for chainID %s and address %s", chainIDStr, address.Hex())
	}

	r.protocols[chainIDStr][address.Hex()] = protocol
	return nil
}

// GetProtocol retrieves a protocol by its contract address.
func (r *ProtocolRegistryImpl) GetProtocol(chainID *big.Int, address common.Address) (Protocol, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chainIDStr := chainID.String()
	if chainProtocols, exists := r.protocols[chainIDStr]; exists {
		if protocol, exists := chainProtocols[address.Hex()]; exists {
			return protocol, nil
		}
	}

	return nil, fmt.Errorf("protocol not found for chainID %s and address %s", chainIDStr, address.Hex())
}

// ListProtocols returns a list of all registered protocols.
func (r *ProtocolRegistryImpl) ListProtocols(chainID *big.Int) []Protocol {
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
func (r *ProtocolRegistryImpl) ListProtocolsByType(chainID *big.Int, protocolType ProtocolType) []Protocol {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chainIDStr := chainID.String()
	if protocols, exists := r.protocolByType[chainIDStr]; exists {
		return protocols[protocolType]
	}

	return []Protocol{}
}

// setupProtocolOperations initializes and registers various DeFi protocols.
func (r *ProtocolRegistryImpl) setupProtocolOperations() error {

	registerProtocol := func(address common.Address, chainID *big.Int,
		createFunc func(ChainConfig) (Protocol, error)) error {

		chainIDStr := chainID.String()
		config, exists := r.chainConfigs[chainIDStr]

		if !exists {
			return fmt.Errorf("chain configuration not found for chainID: %s", chainIDStr)
		}

		protocol, err := createFunc(config)
		if err != nil {
			return fmt.Errorf("failed to create protocol at address %s: %v", address.Hex(), err)
		}

		err = r.RegisterProtocol(big.NewInt(1), address, protocol)
		if err != nil {
			return fmt.Errorf("failed to register protocol at address %s: %v", address.Hex(), err)
		}

		return nil
	}

	client, err := ethclient.Dial("https://eth.public-rpc.com")
	if err != nil {
		return err
	}

	// Register Lido protocol
	err = registerProtocol(LidoContractAddress, big.NewInt(1), func(config ChainConfig) (Protocol, error) {
		lido, err := NewLidoOperation(client, big.NewInt(1))
		return lido, err
	})
	if err != nil {
		return err
	}

	// Register Aave protocol
	err = registerProtocol(AaveV3ContractAddress, big.NewInt(1), func(config ChainConfig) (Protocol, error) {
		aave, err := NewAaveOperation(client, big.NewInt(1), AaveProtocolForkAave)
		return aave, err
	})
	if err != nil {
		return err
	}

	// Sparklend
	err = registerProtocol(SparkLendContractAddress, big.NewInt(1), func(config ChainConfig) (Protocol, error) {
		aave, err := NewAaveOperation(client, big.NewInt(1), AaveProtocolForkSpark)
		return aave, err
	})
	if err != nil {
		return err
	}

	// ankr
	err = registerProtocol(AnkrContractAddress, big.NewInt(1), func(config ChainConfig) (Protocol, error) {
		ankr, err := NewAnkrOperation(client, big.NewInt(1))
		return ankr, err
	})
	if err != nil {
		return err
	}

	// compound
	registerCompoundRegistry(r)

	// err = registerProtocol(AaveV3ContractAddress, big.NewInt(1), func(config ChainConfig) (Protocol, error) {
	// 	parsedABI, err := abi.JSON(strings.NewReader(aaveV3ABI))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	aave := &AaveOperation{}
	// 	protocolConfig := ProtocolConfig{
	// 		RPCURL:   config.RPCURL,
	// 		ChainID:  config.ChainID,
	// 		Contract: AaveV3ContractAddress,
	// 		ABI:      parsedABI,
	// 		Type:     TypeLoan,
	// 	}
	// 	if err := aave.Initialize(context.Background(), protocolConfig); err != nil {
	// 		return nil, err
	// 	}
	// 	return aave, nil
	// })
	// if err != nil {
	// 	return err
	// }

	return nil
}
