//go:build integration
// +build integration

package pkg

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func validateSymbolFromToken(t *testing.T,
	client *ethclient.Client,
	tokenContract common.Address,
	expectedName string) {
	t.Helper()

	parsedABI, err := abi.JSON(strings.NewReader(abiString))
	require.NoError(t, err)

	callData, err := parsedABI.Pack("symbol")
	require.NoError(t, err)

	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &tokenContract,
		Data: callData,
	}, nil)
	require.NoError(t, err)

	name := ""
	err = parsedABI.UnpackIntoInterface(&name, "symbol", result)

	require.Equal(t, expectedName, name)
}

func TestRocketPoolOperation_GenerateCallData_UnsupportedAction(t *testing.T) {

	rp, err := NewRocketpoolOperation(getTestClient(t, ChainETH), big.NewInt(1))
	require.NoError(t, err)

	_, err = rp.GenerateCalldata(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{})
	require.Error(t, err)
}

func TestRocketPoolOperation_GetBalance(t *testing.T) {

	client := getTestClient(t, ChainETH)

	rp, err := NewRocketpoolOperation(client, big.NewInt(1))
	require.NoError(t, err)

	token, got, err := rp.GetBalance(context.Background(), big.NewInt(1), emptyTestWallet, common.HexToAddress(""))
	require.NoError(t, err)
	require.Empty(t, got.Int64())

	validateSymbolFromToken(t, client, token, "rETH")
}

func TestRocketPoolOperation_Validate(t *testing.T) {

	rp, err := NewRocketpoolOperation(getTestClient(t, ChainETH), big.NewInt(1))
	require.NoError(t, err)

	t.Run("unsupported chain", func(t *testing.T) {
		err := rp.Validate(context.Background(), big.NewInt(100), LoanSupply, TransactionParams{
			Amount: big.NewInt(1),
		})
		require.Error(t, err)
	})

	t.Run("unsupported action", func(t *testing.T) {
		err := rp.Validate(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
			Amount: big.NewInt(1),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})
		require.Error(t, err)
	})

	t.Run("error", func(t *testing.T) {
		err := rp.Validate(context.Background(), big.NewInt(1), NativeStake, TransactionParams{
			Amount: big.NewInt(1),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})
		require.Error(t, err)
	})
}

func TestRocketPoolOperation_IsSupportedAsset(t *testing.T) {

	rp, err := NewRocketpoolOperation(getTestClient(t, ChainETH), big.NewInt(1))
	require.NoError(t, err)

	t.Run("native token", func(t *testing.T) {
		got := rp.IsSupportedAsset(context.Background(), big.NewInt(1), common.HexToAddress(nativeDenomAddress))
		require.True(t, got)
	})

	t.Run("rEth", func(t *testing.T) {
		got := rp.IsSupportedAsset(context.Background(), big.NewInt(1),
			common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393"))
		require.True(t, got)
	})
}

func TestRocketPoolOperation_GenerateCallData_SupportedAction(t *testing.T) {

	rp, err := NewRocketpoolOperation(getTestClient(t, ChainETH), big.NewInt(1))
	require.NoError(t, err)

	_, err = rp.GenerateCalldata(context.Background(), big.NewInt(1), NativeStake, TransactionParams{})
	require.NoError(t, err)
}

func TestRocketPoolOperation_GenerateCallData(t *testing.T) {

	amountInWei := new(big.Int)

	// 10,0000 ETH in wei
	amountInWei, ok := amountInWei.SetString("10000000000000000000000", 10)
	require.True(t, ok)

	tt := []struct {
		name     string
		action   ContractAction
		method   ProtocolMethod
		expected string
		args     TransactionParams
		hasError bool
	}{
		{
			name:   "Supply action ( failure, staked amount too low)",
			action: NativeStake,
			method: "deposit",
			// cast calldata "deposit()"
			// 0xd0e30db0
			expected: "0xd0e30db0",
			args: TransactionParams{
				Amount: big.NewInt(1 * 1e6),
			},
			hasError: true,
		},
		{
			name:   "Supply action ( failure, staked amount too high)",
			action: NativeStake,
			method: "deposit",
			// cast calldata "deposit()"
			// 0xd0e30db0
			expected: "0xd0e30db0",
			args: TransactionParams{
				Amount: amountInWei,
			},
			hasError: true,
		},
		// disabling the test as currently rocketpool not accepting eth deposit at this time
		// {
		// 	name:   "Supply action",
		// 	action: NativeStake,
		// 	method: rocketPoolStake,
		// 	// cast calldata "deposit()"
		// 	// 0xd0e30db0
		// 	expected: "0xd0e30db0",
		// 	args: []interface{}{
		// 		big.NewInt(1 * 1e18), // 1 ETH
		// 	},
		// },
	}

	for _, v := range tt {

		t.Run(v.name, func(t *testing.T) {

			rp, err := NewRocketpoolOperation(getTestClient(t, ChainETH), big.NewInt(1))
			require.NoError(t, err)

			err = rp.Validate(context.Background(), big.NewInt(1), v.action, v.args)

			if v.hasError {
				require.Error(t, err)
			} else {

				require.NoError(t, err)
			}

			calldata, err := rp.GenerateCalldata(context.Background(), big.NewInt(1), v.action, v.args)
			require.NoError(t, err)
			require.Equal(t, v.expected, calldata)
		})
	}
}
