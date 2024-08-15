package pkg

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRocketPoolOperation_GenerateCallData_UnsupportedAction(t *testing.T) {

	rp, err := NewRocketpoolOperation(getTestClient(t), big.NewInt(1))
	require.NoError(t, err)

	_, err = rp.GenerateCalldata(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{})
	require.Error(t, err)
}

func TestRocketPoolOperation_GenerateCallData_SupportedAction(t *testing.T) {

	rp, err := NewRocketpoolOperation(getTestClient(t), big.NewInt(1))
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
			method: rocketPoolStake,
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
			method: rocketPoolStake,
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

			rp, err := NewRocketpoolOperation(getTestClient(t), big.NewInt(1))
			require.NoError(t, err)

			err = rp.Validate(context.Background(), big.NewInt(1), v.action, v.args)

			if v.hasError {
				require.Error(t, err)
			} else {

				require.NoError(t, err)
			}

			calldata, err := rp.GenerateCalldata(context.Background(), big.NewInt(1), v.action, v.args)
			require.Equal(t, v.expected, calldata)
		})
	}
}
