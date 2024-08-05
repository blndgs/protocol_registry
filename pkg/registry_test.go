package pkg

import (
	"os"
	"strings"
	"testing"

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

//
// // TestProtocolRegistry_Validate test protocol registry validation.
// func TestProtocolRegistry_Validate(t *testing.T) {
// 	registry := NewProtocolRegistry()
// 	SetupProtocolOperations(getTestRPCURL(t), registry)
//
// 	t.Run("ValidateAave", func(t *testing.T) {
// 		operation, err := registry.GetProtocolOperation(AaveV3ContractAddress, LoanSupply, big.NewInt(1))
// 		require.NoError(t, err)
//
// 		require.Nil(t, operation.Validate(common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")))
// 	})
//
// 	t.Run("ValidateAave_NativeAsset", func(t *testing.T) {
// 		operation, err := registry.GetProtocolOperation(AaveV3ContractAddress, LoanSupply, big.NewInt(1))
// 		require.NoError(t, err)
//
// 		// native token not supported
// 		require.Error(t, operation.Validate(common.HexToAddress(nativeDenomAddress)))
// 	})
//
// 	t.Run("ValidateAave_UnsupportedAsset", func(t *testing.T) {
// 		operation, err := registry.GetProtocolOperation(AaveV3ContractAddress, LoanSupply, big.NewInt(1))
// 		require.NoError(t, err)
//
// 		require.Error(t, operation.Validate(common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb49")))
// 	})
//
// 	t.Run("ValidateLido_NativeAsset", func(t *testing.T) {
// 		operation, err := registry.GetProtocolOperation(LidoContractAddress, NativeStake, big.NewInt(1))
// 		require.NoError(t, err)
//
// 		require.Nil(t, operation.Validate(common.HexToAddress(nativeDenomAddress)))
// 		require.Error(t, operation.Validate(common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb49")))
// 	})
// }
