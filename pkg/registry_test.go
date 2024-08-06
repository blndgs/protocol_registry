package pkg

import (
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// getTestRPCURL helper function that gets the rpc url from env.
func getTestRPCURL(t *testing.T) string {
	t.Helper()
	u := os.Getenv("TEST_ETH_RPC_URL")
	if len(strings.TrimSpace(u)) == 0 {
		u = "https://eth.public-rpc.com"
	}
	require.NotEmpty(t, u)
	return u
}

// TestProtocolRegistry_Validate test protocol registry validation.
func TestProtocolRegistry_Validate(t *testing.T) {
	registry := NewProtocolRegistry(getTestRPCURL(t))

	t.Run("ValidateAave", func(t *testing.T) {
		operation, err := registry.GetProtocolOperation(AaveV3ContractAddress, big.NewInt(1))
		require.NoError(t, err)

		require.Nil(t, operation.Validate(common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")))
	})

	t.Run("GetSpark", func(t *testing.T) {
		_, err := registry.GetProtocolOperation(SparkLendContractAddress, big.NewInt(1))
		require.NoError(t, err)
	})

	t.Run("GetRocketpool", func(t *testing.T) {
		_, err := registry.GetProtocolOperation(RocketPoolStorageAddress, big.NewInt(1))
		require.NoError(t, err)
	})

	t.Run("GetLido", func(t *testing.T) {
		_, err := registry.GetProtocolOperation(LidoContractAddress, big.NewInt(1))
		require.NoError(t, err)
	})

	t.Run("GetAnkr", func(t *testing.T) {
		_, err := registry.GetProtocolOperation(AnkrContractAddress, big.NewInt(1))
		require.NoError(t, err)
	})

	t.Run("GetCompound", func(t *testing.T) {
		_, err := registry.GetProtocolOperation(common.HexToAddress("0xc3d688b66703497daa19211eedff47f25384cdc3"),
			big.NewInt(1))
		require.NoError(t, err)

		_, err = registry.GetProtocolOperation(common.HexToAddress("0xa17581a9e3356d9a858b789d68b4d866e593ae94"),
			big.NewInt(1))
		require.NoError(t, err)
	})
}
