package pkg

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// TestRocketPoolOperation_GenerateCallData_UnsupportedAction test rocket pool unsupported action.
func TestRocketPoolOperation_GenerateCallData_UnsupportedAction(t *testing.T) {

	rp, err := NewRocketPool(getTestRPCURL(t), RocketPoolStorageAddress, LoanSupply, aaveSupply)
	require.Error(t, err)

	require.Nil(t, rp)
}

// TestRocketPoolOperation_GenerateCallData_SupportedAction test rocket pool supported action.
func TestRocketPoolOperation_GenerateCallData_SupportedAction(t *testing.T) {

	rp, err := NewRocketPool(getTestRPCURL(t), RocketPoolStorageAddress, NativeStake, rocketPoolStake)
	require.NoError(t, err)

	require.NotNil(t, rp)
}

// TestRocketPoolOperation_GenerateCallData test rocket pool Stake UnStake calldata.
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
		args     []interface{}
		hasError bool
	}{
		{
			name:   "Supply action ( failure, staked amount too low)",
			action: NativeStake,
			method: rocketPoolStake,
			// cast calldata "deposit()"
			// 0xd0e30db0
			expected: "0xd0e30db0",
			args: []interface{}{
				big.NewInt(1 * 1e6),
			},
			hasError: true,
		},
		{
			name:   "Supply action ( failure, staked amount too high)",
			action: NativeStake,
			method: rocketPoolStake,
			// cast calldata "deposit()"
			// 0xd0e30db0
			expected: "0xd0e30db0",
			args:     []interface{}{amountInWei},
			hasError: true,
		},
		{
			name:   "Supply action",
			action: NativeStake,
			method: rocketPoolStake,
			// cast calldata "deposit()"
			// 0xd0e30db0
			expected: "0xd0e30db0",
			args: []interface{}{
				big.NewInt(1 * 1e18), // 1 ETH
			},
		},
		{
			name:   "Withdraw action",
			action: NativeUnStake,
			method: rocketPoolUnStake,
			// cast calldata "transfer(address,uint256)" 0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6 1000000000000000000
			// 0xa9059cbb000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d60000000000000000000000000000000000000000000000000de0b6b3a7640000
			expected: "0xa9059cbb000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d60000000000000000000000000000000000000000000000000de0b6b3a7640000",
			args: []interface{}{
				common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
				big.NewInt(1 * 1e18),
			},
		},
	}

	registry := NewProtocolRegistry()

	for _, v := range tt {

		t.Run(v.name, func(t *testing.T) {

			rp, err := NewRocketPool(getTestRPCURL(t), RocketPoolStorageAddress, v.action, v.method)
			require.NoError(t, err)

			rp.Register(registry)

			calldata, err := rp.GenerateCalldata(v.args)

			if v.hasError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, v.expected, calldata)
		})
	}
}
