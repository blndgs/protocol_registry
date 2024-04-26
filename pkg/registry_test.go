package pkg

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
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

func TestGenericProtocolOperation_GenerateCalldata(t *testing.T) {
	operation := &GenericProtocolOperation{
		DynamicOperation: DynamicOperation{
			Protocol: "AaveV3",
			Action:   SupplyAction,
			Args: []interface{}{
				common.HexToAddress("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984"),
				big.NewInt(1000000000000000000),
				common.HexToAddress("0x0000000000000000000000000000000000000000"),
				uint16(0),
			},
			ChainID: big.NewInt(1),
		},
	}

	calldata, err := operation.GenerateCalldata()

	// sample cast generate
	// cast calldata "supply(address,uint256,address,uint16)" 0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984 1000000000000000000 0x0000000000000000000000000000000000000000 0
	//
	// 0x617ba0370000000000000000000000001f9840a85d5af5bf1d1762f925bdaddc4201f9840000000000000000000000000000000000000000000000000de0b6b3a764000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
	expectedOutput := "0x617ba0370000000000000000000000001f9840a85d5af5bf1d1762f925bdaddc4201f9840000000000000000000000000000000000000000000000000de0b6b3a764000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	require.NoError(t, err)
	require.NotEmpty(t, calldata)
	require.Equal(t, expectedOutput, calldata)
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
