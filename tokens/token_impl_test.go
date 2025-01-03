//go:build integration
// +build integration

package tokens

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/blndgs/protocol_registry/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sample data for testing
const sampleEthData = `{
	"tokens": [
		{"token_address": "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "name": "USD Coin", "symbol": "USDC", "decimals": 6}
	],
	"protocols": [
		{"address": "0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2", "name": "AaveV3", "source": true, "destination": true, "tokens": ["0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"]}
	]
}`

const sampleBscData = `{
	"tokens": [
		{"token_address": "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d", "name": "USD Coin", "symbol": "USDC", "decimals": 18}
	],
	"protocols": [
		{"address": "0x6807dc923806fE8Fd134338EABCA509979a7e0cB", "name": "AaveV3", "source": true, "destination": true, "tokens": ["0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d"]}
	]
}`

const samplePolygonData = `{
	"tokens": [
		{"token_address": "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359", "name": "USD Coin", "symbol": "USDC", "decimals": 6}
	],
	"protocols": [
		{"address": "0x794a61358D6845594F94dc1DB02A252b5b4814aD", "name": "AaveV3", "source": true, "destination": true, "tokens": ["0x8f3Cf7ad23Cd3CaDbD9735AFf958023239c6A063"]}
	]
}`

func TestNewJSONTokenRegistry(t *testing.T) {
	// Setup: Create temporary JSON files for testing
	tmpDir := t.TempDir()
	createTempJSONFile(t, tmpDir, "1.json", sampleEthData)
	createTempJSONFile(t, tmpDir, "56.json", sampleBscData)
	createTempJSONFile(t, tmpDir, "137.json", samplePolygonData)

	// Change working directory to temp directory
	oldWd, _ := os.Getwd()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	// Test
	registry, err := NewJSONTokenRegistry()
	require.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Len(t, registry.data, 3)
}

func TestGetTokens(t *testing.T) {
	registry, err := NewJSONTokenRegistry()
	require.NoError(t, err)
	tests := []struct {
		name    string
		chainID *big.Int
		want    int
		wantErr bool
	}{
		{"Ethereum chain", pkg.EthChainID, 17, false},
		{"BSC chain", pkg.BscChainID, 9, false},
		{"Polyhon chain", pkg.PolygonChainID, 12, false},
		{"Unknown chain", big.NewInt(999), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := registry.GetTokens(tt.chainID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, tokens, tt.want)
			}
		})
	}
}

func TestGetProtocols(t *testing.T) {
	registry, err := NewJSONTokenRegistry()
	require.NoError(t, err)

	tests := []struct {
		name    string
		chainID *big.Int
		want    int
		wantErr bool
	}{
		{"Ethereum chain", pkg.EthChainID, 7, false},
		{"BSC chain", pkg.BscChainID, 3, false},
		{"Polygon chain", pkg.PolygonChainID, 1, false},
		{"Unknown chain", big.NewInt(999), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protocols, err := registry.GetProtocols(tt.chainID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, protocols, tt.want)
			}
		})
	}
}

func TestGetTokenByAddress(t *testing.T) {
	registry, err := NewJSONTokenRegistry()
	require.NoError(t, err)

	tests := []struct {
		name    string
		chainID *big.Int
		address string
		want    string
		wantErr bool
	}{
		{"Ethereum USDC", pkg.EthChainID, "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "USDC", false},
		{"BSC USDC", pkg.BscChainID, "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d", "USDC", false},
		{"Polygon USDC", pkg.PolygonChainID, "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359", "USDC", false},
		{"Unknown token", pkg.EthChainID, "0x1234567890123456789012345678901234567890", "", true},
		{"Unknown chain", big.NewInt(999), "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := registry.GetTokenByAddress(tt.chainID, tt.address)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, token.Symbol)
			}
		})
	}
}

func TestGetProtocolByAddress(t *testing.T) {
	registry, err := NewJSONTokenRegistry()
	require.NoError(t, err)

	tests := []struct {
		name    string
		chainID *big.Int
		address string
		want    string
		wantErr bool
	}{
		{"Ethereum AaveV3", pkg.EthChainID, "0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2", "AaveV3", false},
		{"BSC AaveV3", pkg.BscChainID, "0x6807dc923806fE8Fd134338EABCA509979a7e0cB", "AaveV3", false},
		{"Polygon AaveV3", pkg.PolygonChainID, "0x794a61358D6845594F94dc1DB02A252b5b4814aD", "AaveV3", false},
		{"Unknown protocol", pkg.EthChainID, "0x1234567890123456789012345678901234567890", "", true},
		{"Unknown chain", big.NewInt(999), "0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protocol, err := registry.GetProtocolByAddress(tt.chainID, tt.address)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, protocol.Name)
			}
		})
	}
}

func createTempJSONFile(t *testing.T, dir, filename, content string) {
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}
