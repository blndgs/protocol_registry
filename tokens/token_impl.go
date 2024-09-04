package tokens

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/blndgs/protocol_registry/pkg"
)

// NewJSONTokenRegistry creates a new JSONTokenRegistry.
func NewJSONTokenRegistry() (*JSONTokenRegistry, error) {
	registry := &JSONTokenRegistry{
		data: make(map[string]*Data),
	}

	chainIDs := []*big.Int{pkg.EthChainID, pkg.BscChainID}
	for _, chainID := range chainIDs {
		fileName := fmt.Sprintf("%d.json", chainID)
		data, err := loadJSONFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("error loading data for chain ID %d: %w", chainID, err)
		}
		registry.data[chainID.String()] = data
	}

	return registry, nil
}

func loadJSONFile(fileName string) (*Data, error) {
	filePath := filepath.Join(".", fileName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", fileName, err)
	}

	var data Data
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON from %s: %w", fileName, err)
	}

	return &data, nil
}

// GetTokens returns all tokens for a given chain ID
func (r *JSONTokenRegistry) GetTokens(chainID *big.Int) ([]Token, error) {
	r.dataLock.RLock()
	defer r.dataLock.RUnlock()

	data, ok := r.data[chainID.String()]
	if !ok {
		return nil, fmt.Errorf("no data available for chain ID %d", chainID)
	}
	return data.Tokens, nil
}

// GetProtocols returns all protocols for a given chain ID
func (r *JSONTokenRegistry) GetProtocols(chainID *big.Int) ([]Protocol, error) {
	r.dataLock.RLock()
	defer r.dataLock.RUnlock()

	data, ok := r.data[chainID.String()]
	if !ok {
		return nil, fmt.Errorf("no data available for chain ID %d", chainID)
	}
	return data.Protocols, nil
}

// GetTokenByAddress returns a token by its address for a given chain ID
func (r *JSONTokenRegistry) GetTokenByAddress(chainID *big.Int, address string) (*Token, error) {
	r.dataLock.RLock()
	defer r.dataLock.RUnlock()

	data, ok := r.data[chainID.String()]
	if !ok {
		return nil, fmt.Errorf("no data available for chain ID %d", chainID)
	}

	for _, token := range data.Tokens {
		if token.TokenAddress == address {
			return &token, nil
		}
	}
	return nil, fmt.Errorf("token not found with address: %s for chain ID %d", address, chainID)
}

// GetProtocolByAddress returns a protocol by its address for a given chain ID
func (r *JSONTokenRegistry) GetProtocolByAddress(chainID *big.Int, address string) (*Protocol, error) {
	r.dataLock.RLock()
	defer r.dataLock.RUnlock()

	data, ok := r.data[chainID.String()]
	if !ok {
		return nil, fmt.Errorf("no data available for chain ID %d", chainID)
	}

	for _, protocol := range data.Protocols {
		if protocol.Address == address {
			return &protocol, nil
		}
	}
	return nil, fmt.Errorf("protocol not found with address: %s for chain ID %d", address, chainID)
}
