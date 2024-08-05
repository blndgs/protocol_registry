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
	registry := NewProtocolRegistry()
	SetupProtocolOperations(getTestRPCURL(t), registry)

	t.Run("ValidateAave", func(t *testing.T) {
		operation, err := registry.GetProtocolOperation(AaveV3ContractAddress, LoanSupply, big.NewInt(1))
		require.NoError(t, err)

		require.Nil(t, operation.Validate(common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")))
	})

	t.Run("ValidateAave_NativeAsset", func(t *testing.T) {
		operation, err := registry.GetProtocolOperation(AaveV3ContractAddress, LoanSupply, big.NewInt(1))
		require.NoError(t, err)

		// native token not supported
		require.Error(t, operation.Validate(common.HexToAddress(nativeDenomAddress)))
	})

	t.Run("ValidateAave_UnsupportedAsset", func(t *testing.T) {
		operation, err := registry.GetProtocolOperation(AaveV3ContractAddress, LoanSupply, big.NewInt(1))
		require.NoError(t, err)

		require.Error(t, operation.Validate(common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb49")))
	})

	t.Run("ValidateLido_NativeAsset", func(t *testing.T) {
		operation, err := registry.GetProtocolOperation(LidoContractAddress, NativeStake, big.NewInt(1))
		require.NoError(t, err)

		require.Nil(t, operation.Validate(common.HexToAddress(nativeDenomAddress)))
		require.Error(t, operation.Validate(common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb49")))
	})
}

// TestProtocolRegistry test protocol registry.
func TestProtocolRegistry(t *testing.T) {
	registry := NewProtocolRegistry()
	SetupProtocolOperations(getTestRPCURL(t), registry)

	t.Run("GetProtocolOperation_Exists", func(t *testing.T) {
		operation, err := registry.GetProtocolOperation(AaveV3ContractAddress, LoanSupply, big.NewInt(1))
		require.NoError(t, err)
		require.NotNil(t, operation)
		require.IsType(t, &GenericProtocolOperation{}, operation)
	})

	t.Run("GetProtocolOperation_NotExists", func(t *testing.T) {
		operation, err := registry.GetProtocolOperation(LidoContractAddress, NativeStake, big.NewInt(2))
		require.Error(t, err)
		require.Nil(t, operation)
	})

	t.Run("RegisterProtocolOperation_InvalidChainID", func(t *testing.T) {
		require.Panics(t, func() {
			registry.RegisterProtocolOperation(AaveV3ContractAddress, LoanSupply, big.NewInt(-1), &GenericProtocolOperation{})
		})
	})

	t.Run("RegisterProtocolOperation_NilOperation", func(t *testing.T) {
		require.Panics(t, func() {
			registry.RegisterProtocolOperation(AaveV3ContractAddress, LoanSupply, big.NewInt(1), nil)
		})
	})
}

// TestProtocolOperations test protocol operations.
func TestProtocolOperations(t *testing.T) {
	registry := NewProtocolRegistry()
	SetupProtocolOperations(getTestRPCURL(t), registry)
	tests := []struct {
		name     string
		protocol ContractAddress
		action   ContractAction
		args     []interface{}
		expected string
	}{
		{
			name:     "AaveV3 Supply",
			protocol: AaveV3ContractAddress,
			action:   LoanSupply,
			args: []interface{}{
				common.HexToAddress("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984"),
				big.NewInt(1000000000000000000),
				common.HexToAddress("0x0000000000000000000000000000000000000000"),
				uint16(10),
			},
			// cast calldata "supply(address,uint256,address,uint16)" 0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984 1000000000000000000 0x0000000000000000000000000000000000000000 10
			// 0x617ba0370000000000000000000000001f9840a85d5af5bf1d1762f925bdaddc4201f9840000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a
			expected: "0x617ba0370000000000000000000000001f9840a85d5af5bf1d1762f925bdaddc4201f9840000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a",
		},
		// {
		// 	name:     "SparkLend Withdraw",
		// 	protocol: SparkLendContractAddress,
		// 	action:   LoanWithdraw,
		// 	args: []interface{}{
		// 		common.HexToAddress("0xc0ffee254729296a45a3885639AC7E10F9d54979"),
		// 		big.NewInt(500000000000000000),
		// 		common.HexToAddress("0x0000000000000000000000000000000000000000"),
		// 	},
		// 	// cast calldata "withdraw(address,uint256,address)" 0xc0ffee254729296a45a3885639AC7E10F9d54979 500000000000000000 0x0000000000000000000000000000000000000000
		// 	// 0x69328dec000000000000000000000000c0ffee254729296a45a3885639ac7e10f9d5497900000000000000000000000000000000000000000000000006f05b59d3b200000000000000000000000000000000000000000000000000000000000000000000
		// 	expected: "0x69328dec000000000000000000000000c0ffee254729296a45a3885639ac7e10f9d5497900000000000000000000000000000000000000000000000006f05b59d3b200000000000000000000000000000000000000000000000000000000000000000000",
		// },
		// {
		// 	name:     "Lido Stake",
		// 	protocol: LidoContractAddress,
		// 	action:   NativeStake,
		// 	args: []interface{}{
		// 		common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
		// 	},
		// 	// cast calldata "submit(address)" 0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6
		// 	// 0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6
		// 	expected: "0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6",
		// },
		// {
		// 	name:     "Ankr staking ( deposit )",
		// 	protocol: AnkrContractAddress,
		// 	action:   NativeStake,
		// 	args:     []interface{}{},
		// 	// cast calldata "stakeAndClaimAethC()"
		// 	// 0x9fa65c56
		// 	expected: "0x9fa65c56",
		// },
		// {
		// 	name:     "Ankr staking ( withdrawal )",
		// 	protocol: AnkrContractAddress,
		// 	action:   NativeUnStake,
		// 	args: []interface{}{
		// 		big.NewInt(3987509938965136896),
		// 	},
		// 	// cast calldata "unstakeAETH(uint256)" 3987509938965136896
		// 	// 0xc957619d00000000000000000000000000000000000000000000000037567b29aa5b4600
		// 	expected: "0xc957619d00000000000000000000000000000000000000000000000037567b29aa5b4600",
		// },
		// {
		// 	name:     "Renzo ETH Stake",
		// 	protocol: RenzoManagerAddress,
		// 	action:   NativeStake,
		// 	args:     []interface{}{},
		// 	// cast calldata "depositETH()"
		// 	// 0xf6326fb3
		// 	expected: "0xf6326fb3",
		// },
		// {
		// 	name:     "Renzo ERC20 Stake",
		// 	protocol: RenzoManagerAddress,
		// 	action:   ERC20Stake,
		// 	args: []interface{}{
		// 		common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
		// 		big.NewInt(1 * 1e18)},
		// 	// cast calldata "deposit(address,uint256)" 0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6 1000000000000000000
		// 	// 0x47e7ef24000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d60000000000000000000000000000000000000000000000000de0b6b3a7640000
		// 	expected: "0x47e7ef24000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d60000000000000000000000000000000000000000000000000de0b6b3a7640000",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation, err := registry.GetProtocolOperation(tt.protocol, tt.action, big.NewInt(1))
			require.NoError(t, err)
			require.NotNil(t, operation)

			calldata, err := operation.GenerateCalldata(tt.args)
			require.NoError(t, err)
			require.Equal(t, tt.expected, calldata)
		})
	}
}
