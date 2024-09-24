package pkg

import (
	"errors"
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

// setupProtocolOperations initializes and registers various DeFi protocols for both ETH and BNB.
func (r *ProtocolRegistryImpl) setupProtocolOperations() error {
	val, ok := r.chainConfigs[EthChainStr]
	if !ok {
		return errors.New("please provide ETH chain config")
	}

	bscConfig, ok := r.chainConfigs[BscChainStr]
	if !ok {
		return errors.New("please provide BSC chain config")
	}

	polygonConfig, ok := r.chainConfigs[PolygonChainStr]
	if !ok {
		return errors.New("please provide Polygon chain config")
	}

	// Initialize ETH client
	client, err := ethclient.Dial(val.RPCURL)
	if err != nil {
		return err
	}

	// Initialize BSC client
	bscClient, err := ethclient.Dial(bscConfig.RPCURL)
	if err != nil {
		return err
	}

	// Initialize Polygon client
	polygonClient, err := ethclient.Dial(polygonConfig.RPCURL)
	if err != nil {
		return err
	}

	// Setup protocols for ETH
	err = r.setupEthProtocols(client)
	if err != nil {
		return err
	}

	// Setup protocols for BNB
	err = r.setupBnbProtocols(bscClient)
	if err != nil {
		return err
	}

	// Setup protocols for Polygon
	return r.setupPolygonProtocols(polygonClient)
}

// setupPolygonProtocols initializes and registers various DeFi protocols on the Polygon chain.
func (r *ProtocolRegistryImpl) setupPolygonProtocols(client *ethclient.Client) error {

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

		err = r.RegisterProtocol(chainID, address, protocol)
		if err != nil {
			return fmt.Errorf("failed to register protocol at address %s: %v", address.Hex(), err)
		}

		return nil
	}

	// Register Aave protocol on Polygon
	return registerProtocol(
		AavePolygonV3ContractAddress,
		PolygonChainID,
		func(config ChainConfig) (Protocol, error) {
			return NewAaveOperation(client, PolygonChainID, AaveProtocolDeploymentPolygon)
		})
}

// setupEthProtocols initializes and registers various DeFi protocols on the Ethereum chain.
func (r *ProtocolRegistryImpl) setupEthProtocols(client *ethclient.Client) error {

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

		err = r.RegisterProtocol(chainID, address, protocol)
		if err != nil {
			return fmt.Errorf("failed to register protocol at address %s: %v", address.Hex(), err)
		}

		return nil
	}

	// Register Lido protocol on Ethereum
	err := registerProtocol(LidoContractAddress, EthChainID, func(config ChainConfig) (Protocol, error) {
		return NewLidoOperation(client, EthChainID)
	})
	if err != nil {
		return err
	}

	// Register Aave protocol on Ethereum
	err = registerProtocol(AaveEthereumV3ContractAddress, EthChainID, func(config ChainConfig) (Protocol, error) {
		return NewAaveOperation(client, EthChainID, AaveProtocolDeploymentEthereum)
	})
	if err != nil {
		return err
	}

	// Register Sparklend protocol on Ethereum
	err = registerProtocol(SparkLendContractAddress, EthChainID, func(config ChainConfig) (Protocol, error) {
		return NewAaveOperation(client, EthChainID, AaveProtocolDeploymentSpark)
	})
	if err != nil {
		return err
	}

	// Register Ankr protocol on Ethereum
	err = registerProtocol(AnkrContractAddress, EthChainID, func(config ChainConfig) (Protocol, error) {
		return NewAnkrOperation(client, EthChainID)
	})
	if err != nil {
		return err
	}

	// Register Rocketpool protocol on Ethereum
	err = registerProtocol(RocketPoolStorageAddress, EthChainID, func(config ChainConfig) (Protocol, error) {
		return NewRocketpoolOperation(client, EthChainID)
	})
	if err != nil {
		return err
	}

	// Register Compound protocol on Ethereum
	return registerCompoundRegistry(r, client)
}

// setupBnbProtocols initializes and registers various DeFi protocols on the Binance Smart Chain.
func (r *ProtocolRegistryImpl) setupBnbProtocols(client *ethclient.Client) error {

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

		err = r.RegisterProtocol(chainID, address, protocol)
		if err != nil {
			return fmt.Errorf("failed to register protocol at address %s: %v", address.Hex(), err)
		}

		return nil
	}

	// Register Aave protocol on BNB
	err := registerProtocol(AaveBnbV3ContractAddress, BscChainID, func(config ChainConfig) (Protocol, error) {
		return NewAaveOperation(client, BscChainID, AaveProtocolDeploymentEthereum)
	})
	if err != nil {
		return err
	}

	// Register Avalon Finance protocol on BNB
	err = registerProtocol(AvalonFinanceContractAddress, BscChainID, func(config ChainConfig) (Protocol, error) {
		return NewAaveOperation(client, BscChainID, AaveProtocolDeploymentAvalonFinance)
	})
	if err != nil {
		return err
	}

	// Register Lista Dao protocol on BNB
	err = registerProtocol(ListaDaoContractAddress, BscChainID, func(config ChainConfig) (Protocol, error) {
		return NewListaStakingOperation(client, BscChainID)
	})
	if err != nil {
		return err
	}

	return nil
}
