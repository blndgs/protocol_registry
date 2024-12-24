package tokens

import (
	"math/big"
	"sync"
)

// Token represents a token with its properties
type Token struct {
	TokenAddress string `json:"token_address"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Decimals     int    `json:"decimals"`
}

// Protocol represents a protocol with its properties
type Protocol struct {
	Address     string   `json:"address,omitempty"`
	Name        string   `json:"name,omitempty"`
	Source      bool     `json:"source,omitempty"`
	Destination bool     `json:"destination,omitempty"`
	Tokens      []string `json:"tokens,omitempty"`
	Type        string   `json:"type,omitempty"`
}

// Data represents the entire structure of the JSON file
type Data struct {
	Tokens    []Token    `json:"tokens"`
	Protocols []Protocol `json:"protocols"`
}

// TokenRegistry is an interface for retrieving token and protocol data
type TokenRegistry interface {
	// GetTokens retrieves all tokens for a given chain ID.
	GetTokens(chainID *big.Int) ([]Token, error)

	// GetProtocols retrieves all protocols for a given chain ID.
	GetProtocols(chainID *big.Int) ([]Protocol, error)

	// GetTokenByAddress retrieves a specific token by its address for a given chain ID.
	GetTokenByAddress(chainID *big.Int, address string) (*Token, error)

	// GetProtocolByAddress retrieves a specific protocol by its address for a given chain ID.
	GetProtocolByAddress(chainID *big.Int, address string) (*Protocol, error)
}

// JSONTokenRegistry implements TokenRegistry for JSON files
type JSONTokenRegistry struct {
	data     map[string]*Data
	dataLock sync.RWMutex
}
