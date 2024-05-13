package pkg

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func getTestRPCURL(t *testing.T) string {
	t.Helper()

	// u := os.Getenv("TEST_ETH_RPC_URL")
	u := "https://eth.public-rpc.com"

	require.NotEmpty(t, u)
	return u
}

func TestProtocolRegistry(t *testing.T) {
	registry := NewProtocolRegistry()
	SetupProtocolOperations(getTestRPCURL(t), registry)

	t.Run("GetProtocolOperation_Exists", func(t *testing.T) {
		operation, err := registry.GetProtocolOperation(AaveV3, SupplyAction, big.NewInt(1))
		require.NoError(t, err)
		require.NotNil(t, operation)
		require.IsType(t, &GenericProtocolOperation{}, operation)
	})

	t.Run("GetProtocolOperation_NotExists", func(t *testing.T) {
		operation, err := registry.GetProtocolOperation(AaveV3, SupplyAction, big.NewInt(2))
		require.Error(t, err)
		require.Nil(t, operation)
	})

	t.Run("RegisterProtocolOperation_InvalidChainID", func(t *testing.T) {
		require.Panics(t, func() {
			registry.RegisterProtocolOperation(AaveV3, SupplyAction, big.NewInt(-1), &GenericProtocolOperation{})
		})
	})

	t.Run("RegisterProtocolOperation_NilOperation", func(t *testing.T) {
		require.Panics(t, func() {
			registry.RegisterProtocolOperation(AaveV3, SupplyAction, big.NewInt(1), nil)
		})
	})
}
func TestProtocolOperations(t *testing.T) {
	registry := NewProtocolRegistry()
	SetupProtocolOperations(getTestRPCURL(t), registry)
	tests := []struct {
		name     string
		protocol ProtocolName
		action   ContractAction
		kind     AssetKind
		args     []interface{}
		expected string
	}{
		{
			name:     "AaveV3 Supply",
			protocol: AaveV3,
			action:   SupplyAction,
			kind:     LoanKind,
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
		{
			name:     "SparkLend Withdraw",
			protocol: SparkLend,
			action:   WithdrawAction,
			kind:     LoanKind,
			args: []interface{}{
				common.HexToAddress("0xc0ffee254729296a45a3885639AC7E10F9d54979"),
				big.NewInt(500000000000000000),
				common.HexToAddress("0x0000000000000000000000000000000000000000"),
			},
			// cast calldata "withdraw(address,uint256,address)" 0xc0ffee254729296a45a3885639AC7E10F9d54979 500000000000000000 0x0000000000000000000000000000000000000000
			// 0x69328dec000000000000000000000000c0ffee254729296a45a3885639ac7e10f9d5497900000000000000000000000000000000000000000000000006f05b59d3b200000000000000000000000000000000000000000000000000000000000000000000
			expected: "0x69328dec000000000000000000000000c0ffee254729296a45a3885639ac7e10f9d5497900000000000000000000000000000000000000000000000006f05b59d3b200000000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:     "Lido Stake",
			protocol: Lido,
			action:   SubmitAction,
			kind:     StakeKind,
			args: []interface{}{
				common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
			},
			// cast calldata "submit(address)" 0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6
			// 0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6
			expected: "0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6",
		},
		{
			name:     "Rocket Pool Stake",
			protocol: RocketPool,
			action:   SubmitAction,
			kind:     StakeKind,
			args: []interface{}{
				big.NewInt(1 * 1e6),
			},
			// cast calldata "deposit()"
			// 0xd0e30db0
			expected: "0xd0e30db0",
		},
		{
			name:     "Rocket Pool UnStake (withdraw)",
			protocol: RocketPool,
			action:   WithdrawAction,
			kind:     StakeKind,
			args: []interface{}{
				common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
				big.NewInt(1 * 1e18),
			},
			// cast calldata "transfer(address,uint256)" 0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6 1000000000000000000
			// 0xa9059cbb000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d60000000000000000000000000000000000000000000000000de0b6b3a7640000
			expected: "0xa9059cbb000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d60000000000000000000000000000000000000000000000000de0b6b3a7640000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation, err := registry.GetProtocolOperation(tt.protocol, tt.action, big.NewInt(1))
			require.NoError(t, err)
			require.NotNil(t, operation)

			calldata, err := operation.GenerateCalldata(tt.kind, tt.args)
			require.NoError(t, err)
			require.Equal(t, tt.expected, calldata)
		})
	}
}

func TestSetupProtocolOperations(t *testing.T) {
	registry := NewProtocolRegistry()
	SetupProtocolOperations(getTestRPCURL(t), registry)

	// Iterate over each asset kind and their associated protocols
	for _, protocols := range SupportedProtocols {
		for _, proto := range protocols {
			// Parse the ABI once per protocol for testing
			parsedABI, err := abi.JSON(strings.NewReader(proto.ABI))
			require.NoError(t, err, "ABI should be correctly parsed without error")

			// Test each method in the parsed ABI
			for _, method := range parsedABI.Methods {
				action := ContractAction(method.Name)
				operation, err := registry.GetProtocolOperation(proto.Name, action, big.NewInt(1))

				require.NoError(t, err, "Should not have error retrieving operation for protocol %s and action %s", proto.Name, action)
				require.NotNil(t, operation, "Operation should not be nil for protocol %s and action %s", proto.Name, action)
				require.IsType(t, &GenericProtocolOperation{}, operation, "Operation should be of type *GenericProtocolOperation")
			}
		}
	}
}
