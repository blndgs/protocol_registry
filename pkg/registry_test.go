package pkg

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/require"
)

func TestProtocolRegistry(t *testing.T) {
	registry := NewProtocolRegistry()
	SetupProtocolOperations(registry)

	t.Run("GetProtocolOperation_Exists", func(t *testing.T) {
		operation, err := registry.GetProtocolOperation("AaveV3", SupplyAction, big.NewInt(1))
		require.NoError(t, err)
		require.NotNil(t, operation)
		require.IsType(t, &GenericProtocolOperation{}, operation)
	})

	t.Run("GetProtocolOperation_NotExists", func(t *testing.T) {
		operation, err := registry.GetProtocolOperation("AaveV3", SupplyAction, big.NewInt(2))
		require.Error(t, err)
		require.Nil(t, operation)
	})

	t.Run("RegisterProtocolOperation_UnsupportedProtocol", func(t *testing.T) {
		require.Panics(t, func() {
			registry.RegisterProtocolOperation("UnsupportedProtocol", SupplyAction, big.NewInt(1), &GenericProtocolOperation{})
		})
	})

	t.Run("RegisterProtocolOperation_InvalidChainID", func(t *testing.T) {
		require.Panics(t, func() {
			registry.RegisterProtocolOperation("AaveV3", SupplyAction, big.NewInt(-1), &GenericProtocolOperation{})
		})
	})

	t.Run("RegisterProtocolOperation_NilOperation", func(t *testing.T) {
		require.Panics(t, func() {
			registry.RegisterProtocolOperation("AaveV3", SupplyAction, big.NewInt(1), nil)
		})
	})
}

func TestProtocolOperations(t *testing.T) {
	registry := NewProtocolRegistry()
	SetupProtocolOperations(registry)

	tests := []struct {
		name     string
		protocol string
		action   ContractAction
		args     []interface{}
		expected string
	}{
		{
			name:     "AaveV3 Supply",
			protocol: "AaveV3",
			action:   SupplyAction,
			args: []interface{}{
				"0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984",
				"1000000000000000000",
				"0x0000000000000000000000000000000000000000",
				"0",
			},
			// cast calldata "supply(address,uint256,address,uint16)" 0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984 1000000000000000000 0x0000000000000000000000000000000000000000 0
			// 0x617ba0370000000000000000000000001f9840a85d5af5bf1d1762f925bdaddc4201f9840000000000000000000000000000000000000000000000000de0b6b3a764000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
			expected: "0x617ba0370000000000000000000000001f9840a85d5af5bf1d1762f925bdaddc4201f9840000000000000000000000000000000000000000000000000de0b6b3a764000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:     "SparkLend Withdraw",
			protocol: "SparkLend",
			action:   WithdrawAction,
			args: []interface{}{
				"0xc0ffee254729296a45a3885639AC7E10F9d54979",
				"500000000000000000",
				"0x0000000000000000000000000000000000000000",
			},
			// cast calldata "withdraw(address,uint256,address)" 0xc0ffee254729296a45a3885639AC7E10F9d54979 500000000000000000 0x0000000000000000000000000000000000000000
			// 0x69328dec000000000000000000000000c0ffee254729296a45a3885639ac7e10f9d5497900000000000000000000000000000000000000000000000006f05b59d3b200000000000000000000000000000000000000000000000000000000000000000000
			expected: "0x69328dec000000000000000000000000c0ffee254729296a45a3885639ac7e10f9d5497900000000000000000000000000000000000000000000000006f05b59d3b200000000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:     "Lido Stake",
			protocol: "Lido",
			action:   StakingAction,
			args: []interface{}{
				"0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6",
			},
			// cast calldata "submit(address)" 0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6
			// 0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6
			expected: "0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation, err := registry.GetProtocolOperation(tt.protocol, tt.action, big.NewInt(1))
			require.NoError(t, err)
			require.NotNil(t, operation)

			// Cast the operation to *GenericProtocolOperation to access GenerateCalldata
			genOp, ok := operation.(*GenericProtocolOperation)
			require.True(t, ok)

			// Set the Args field in the GenericProtocolOperation
			genOp.Args = tt.args

			calldata, err := genOp.GenerateCalldata()
			require.NoError(t, err)
			require.Equal(t, tt.expected, calldata)
		})
	}
}
func TestSetupProtocolOperations(t *testing.T) {
	registry := NewProtocolRegistry()
	SetupProtocolOperations(registry)

	for protocol, protocolDetails := range SupportedProtocols {
		parsedABI, _ := abi.JSON(strings.NewReader(protocolDetails.ABI))
		for _, method := range parsedABI.Methods {
			action := ContractAction(method.Name)
			operation, err := registry.GetProtocolOperation(protocol, action, big.NewInt(1))
			require.NoError(t, err)
			require.NotNil(t, operation)
			require.IsType(t, &GenericProtocolOperation{}, operation)
		}
	}
}
