package pkg_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type TokenRegistry struct {
	Tokens    []Token    `json:"tokens"`
	Protocols []Protocol `json:"protocols"`
}

type Token struct {
	TokenAddress string `json:"token_address"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Decimals     int    `json:"decimals"`
}

type Protocol struct {
	Address     string   `json:"address"`
	Name        string   `json:"name"`
	Source      bool     `json:"source"`
	Destination bool     `json:"destination"`
	Tokens      []string `json:"tokens"`
}

func validateTokenRegistry(t *testing.T, path string) error {
	// Read the JSON file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var registry TokenRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", path, err)
	}

	// Create a map of valid token addresses for O(1) lookup
	validTokens := make(map[string]bool)
	for _, token := range registry.Tokens {
		validTokens[token.TokenAddress] = true
	}

	// Check each protocol's tokens
	for _, protocol := range registry.Protocols {
		for _, tokenAddr := range protocol.Tokens {
			if !validTokens[tokenAddr] {
				return fmt.Errorf("protocol %s (%s) references non-existent token: %s",
					protocol.Name, protocol.Address, tokenAddr)
			}
		}
	}

	return nil
}

func TestValidateTokenRegistry(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Ethereum mainnet",
			path:    "../tokens/1.json",
			wantErr: false,
		},
		{
			name:    "BSC mainnet",
			path:    "../tokens/56.json",
			wantErr: false,
		},
		{
			name:    "Polygon mainnet",
			path:    "../tokens/137.json",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTokenRegistry(t, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTokenRegistry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAllRegistries(t *testing.T) {
	registryFiles := []string{
		"../tokens/1.json",
		"../tokens/56.json",
		"../tokens/137.json",
	}

	for _, file := range registryFiles {
		if err := validateTokenRegistry(t, file); err != nil {
			t.Errorf("validation failed for %s: %v", file, err)
		}
	}
}
