//go:build integration
// +build integration

package pkg

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// getTestRPCURL helper function that gets the rpc url from env.
func getTestRPCURL(t *testing.T, c Chain) string {
	t.Helper()

	u := os.Getenv(fmt.Sprintf("TEST_%s_RPC_URL", c.String()))

	if len(strings.TrimSpace(u)) == 0 {
		switch c {
		case ChainETH:
			u = "https://eth.public-rpc.com"
		case ChainBSC:
			u = "https://bsc-dataseed1.binance.org/"
		case ChainPOLYGON:
			u = "https://1rpc.io/matic"
		}
	}

	require.NotEmpty(t, u)
	return u
}

// TestProtocolRegistry_Validate test protocol registry validation.
func TestProtocolRegistry_Validate(t *testing.T) {
	registry, err := NewProtocolRegistry([]ChainConfig{
		{
			ChainID: big.NewInt(1),
			RPCURL:  getTestRPCURL(t, ChainETH),
		},
		{
			ChainID: big.NewInt(56),
			RPCURL:  getTestRPCURL(t, ChainBSC),
		},
		{
			ChainID: big.NewInt(137),
			RPCURL:  getTestRPCURL(t, ChainPOLYGON),
		},
	})

	require.NoError(t, err)

	t.Run("ValidateLido", func(t *testing.T) {
		operation, err := registry.GetProtocol(big.NewInt(1), LidoContractAddress)
		require.NoError(t, err)

		require.NoError(t, operation.Validate(context.Background(), big.NewInt(1), NativeStake, TransactionParams{
			Amount: big.NewInt(100),
			Asset:  common.HexToAddress(nativeDenomAddress),
		}))
	})

	t.Run("ValidateListaDaoContractAddress", func(t *testing.T) {
		operation, err := registry.GetProtocol(big.NewInt(56), ListaDaoContractAddress)
		require.NoError(t, err)

		require.NoError(t, operation.Validate(context.Background(), big.NewInt(56), NativeStake, TransactionParams{
			Amount: big.NewInt(100),
			Asset:  common.HexToAddress(nativeDenomAddress),
		}))
	})

	t.Run("ValidateAave_NativeAsset", func(t *testing.T) {
		operation, err := registry.GetProtocol(big.NewInt(1), AaveEthereumV3ContractAddress)
		require.NoError(t, err)

		// native token not supported
		require.Error(t, operation.Validate(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
			Amount: big.NewInt(1000),
			Asset:  common.HexToAddress(nativeDenomAddress),
		}))
	})
}

// TestProtocolRegistry test protocol registry.
func TestProtocolRegistry(t *testing.T) {
	registry, err := NewProtocolRegistry([]ChainConfig{
		{
			ChainID: big.NewInt(1),
			RPCURL:  getTestRPCURL(t, ChainETH),
		},
		{
			ChainID: big.NewInt(56),
			RPCURL:  getTestRPCURL(t, ChainBSC),
		},
		{
			ChainID: big.NewInt(137),
			RPCURL:  getTestRPCURL(t, ChainPOLYGON),
		},
	})

	require.NoError(t, err)

	t.Run("GetProtocolOperation_Exists", func(t *testing.T) {
		operation, err := registry.GetProtocol(big.NewInt(1), AaveEthereumV3ContractAddress)
		require.NoError(t, err)
		require.NotNil(t, operation)
	})

	t.Run("GetProtocolOperation_NotExists wrong chain", func(t *testing.T) {
		operation, err := registry.GetProtocol(big.NewInt(100), AaveEthereumV3ContractAddress)
		require.Error(t, err)
		require.Nil(t, operation)
	})

	t.Run("RegisterProtocolOperation_InvalidChainID", func(t *testing.T) {
		err := registry.RegisterProtocol(big.NewInt(11), AaveEthereumV3ContractAddress, nil)
		require.Error(t, err)
	})
}

// TestProtocolOperations test protocol operations.
func TestProtocolOperations(t *testing.T) {

	registry, err := NewProtocolRegistry([]ChainConfig{
		{
			ChainID: big.NewInt(1),
			RPCURL:  getTestRPCURL(t, ChainETH),
		},
		{
			ChainID: big.NewInt(56),
			RPCURL:  getTestRPCURL(t, ChainBSC),
		},
		{
			ChainID: big.NewInt(137),
			RPCURL:  getTestRPCURL(t, ChainPOLYGON),
		},
	})

	require.NoError(t, err)

	tests := []struct {
		name     string
		protocol ContractAddress
		action   ContractAction
		args     TransactionParams
		expected string
	}{
		{
			name:     "AaveV3 Supply",
			protocol: AaveEthereumV3ContractAddress,
			action:   LoanSupply,
			args: TransactionParams{
				Asset:  common.HexToAddress("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984"),
				Amount: big.NewInt(1000000000000000000),
				Sender: common.HexToAddress("0x0000000000000000000000000000000000000000"),
				ExtraData: map[string]interface{}{
					"referral_code": uint16(10),
				},
			},
			// cast calldata "supply(address,uint256,address,uint16)" 0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984 1000000000000000000 0x0000000000000000000000000000000000000000 10
			// 0x617ba0370000000000000000000000001f9840a85d5af5bf1d1762f925bdaddc4201f9840000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a
			expected: "0x617ba0370000000000000000000000001f9840a85d5af5bf1d1762f925bdaddc4201f9840000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a",
		},
		{
			name:     "SparkLend Withdraw",
			protocol: SparkLendContractAddress,
			action:   LoanWithdraw,
			args: TransactionParams{
				Asset:  common.HexToAddress("0xc0ffee254729296a45a3885639AC7E10F9d54979"),
				Amount: big.NewInt(500000000000000000),
				Sender: common.HexToAddress("0x0000000000000000000000000000000000000000"),
			},
			// cast calldata "withdraw(address,uint256,address)" 0xc0ffee254729296a45a3885639AC7E10F9d54979 500000000000000000 0x0000000000000000000000000000000000000000
			// 0x69328dec000000000000000000000000c0ffee254729296a45a3885639ac7e10f9d5497900000000000000000000000000000000000000000000000006f05b59d3b200000000000000000000000000000000000000000000000000000000000000000000
			expected: "0x69328dec000000000000000000000000c0ffee254729296a45a3885639ac7e10f9d5497900000000000000000000000000000000000000000000000006f05b59d3b200000000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:     "Lido Stake",
			protocol: LidoContractAddress,
			action:   NativeStake,
			args: TransactionParams{
				Sender: common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
			},
			// cast calldata "submit(address)" 0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6
			// 0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6
			expected: "0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6",
		},
		{
			name:     "Ankr staking ( deposit )",
			protocol: AnkrContractAddress,
			action:   NativeStake,
			args:     TransactionParams{},
			// cast calldata "stakeAndClaimAethC()"
			// 0x9fa65c56
			expected: "0x9fa65c56",
		},
		{
			name:     "Ankr staking ( withdrawal )",
			protocol: AnkrContractAddress,
			action:   NativeUnStake,
			args: TransactionParams{
				Amount: big.NewInt(3987509938965136896),
			},
			// cast calldata "unstakeAETH(uint256)" 3987509938965136896
			// 0xc957619d00000000000000000000000000000000000000000000000037567b29aa5b4600
			expected: "0xc957619d00000000000000000000000000000000000000000000000037567b29aa5b4600",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation, err := registry.GetProtocol(big.NewInt(1), tt.protocol)
			require.NoError(t, err)
			require.NotNil(t, operation)

			calldata, err := operation.GenerateCalldata(context.Background(), big.NewInt(1), tt.action, tt.args)
			require.NoError(t, err)
			require.Equal(t, tt.expected, calldata)
		})
	}
}

func TestProtocolOperationIntegration(t *testing.T) {

	registry, err := NewProtocolRegistry([]ChainConfig{
		{
			ChainID: big.NewInt(1),
			RPCURL:  getTestRPCURL(t, ChainETH),
		},
		{
			ChainID: big.NewInt(56),
			RPCURL:  getTestRPCURL(t, ChainBSC),
		},
		{
			ChainID: big.NewInt(137),
			RPCURL:  getTestRPCURL(t, ChainPOLYGON),
		},
	})

	require.NoError(t, err)

	protocol, err := registry.GetProtocol(big.NewInt(1), AaveEthereumV3ContractAddress)
	if err != nil {
		panic(err)
	}

	params := TransactionParams{
		Amount: big.NewInt(10 * 1e6),
		// top 10 holder of token onchain
		Sender: common.HexToAddress("0x70f1196Ce323B7B1A741C1E0f4FB84Ac3385FC02"),
		Asset:  common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"),
		ExtraData: map[string]interface{}{
			"referral_code": uint16(0),
		},
	}

	isSupported := protocol.IsSupportedAsset(context.Background(), big.NewInt(1),
		common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"))

	require.True(t, isSupported)

	t.Run("test loan withdraw", func(t *testing.T) {

		err = protocol.Validate(context.Background(), big.NewInt(1), LoanWithdraw, params)
		require.NoError(t, err)

		_, err = protocol.GenerateCalldata(context.Background(), big.NewInt(1), LoanWithdraw, params)
		require.NoError(t, err)
	})

	t.Run("test loan supply", func(t *testing.T) {

		err = protocol.Validate(context.Background(), big.NewInt(1), LoanSupply, params)
		require.NoError(t, err)

		_, err = protocol.GenerateCalldata(context.Background(), big.NewInt(1), LoanSupply, params)
		require.NoError(t, err)
	})
}
