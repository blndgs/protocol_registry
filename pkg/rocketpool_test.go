package pkg

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// TestRocketPoolOperation_GenerateCallData test rocket pool Stake UnStake calldata.
func TestRocketPoolOperation_GenerateCallData(t *testing.T) {

	amountInWei := new(big.Int)

	// 10,0000 ETH in wei
	amountInWei, ok := amountInWei.SetString("10000000000000000000000", 10)
	require.True(t, ok)

	tt := []struct {
		name     string
		action   ContractAction
		expected string
		opts     GenerateCalldataOptions
		hasError bool
	}{
		{
			name:   "Supply action ( failure, staked amount too low)",
			action: NativeStake,
			// cast calldata "deposit()"
			// 0xd0e30db0
			expected: "0xd0e30db0",
			opts: GenerateCalldataOptions{
				Amount: big.NewInt(1 * 1e6),
			},
			hasError: true,
		},
		{
			name:   "Supply action ( failure, staked amount too high)",
			action: NativeStake,
			// cast calldata "deposit()"
			// 0xd0e30db0
			expected: "0xd0e30db0",
			opts: GenerateCalldataOptions{
				Amount: amountInWei,
			},
			hasError: true,
		},
		// Rocketpool not taking eth right now
		// {
		// 	name:   "Supply action",
		// 	action: NativeStake,
		// 	// cast calldata "deposit()"
		// 	// 0xd0e30db0
		// 	expected: "0xd0e30db0",
		// 	opts: GenerateCalldataOptions{
		// 		Amount: big.NewInt(1e5), // 1 ETH
		// 	},
		// },
		{
			name:   "Withdraw action using sender",
			action: NativeUnStake,
			// cast calldata "transfer(address,uint256)" 0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6 1000000000000000000
			// 0xa9059cbb000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d60000000000000000000000000000000000000000000000000de0b6b3a7640000
			expected: "0xa9059cbb000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d60000000000000000000000000000000000000000000000000de0b6b3a7640000",
			opts: GenerateCalldataOptions{
				Sender: common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
				Amount: big.NewInt(1 * 1e18),
			},
		},
		{
			name:   "Withdraw action using recipient",
			action: NativeUnStake,
			// cast calldata "transfer(address,uint256)" 0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6 1000000000000000000
			// 0xa9059cbb000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d60000000000000000000000000000000000000000000000000de0b6b3a7640000
			expected: "0xa9059cbb000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d60000000000000000000000000000000000000000000000000de0b6b3a7640000",
			opts: GenerateCalldataOptions{
				Recipient: common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
				Amount:    big.NewInt(1 * 1e18),
			},
		},
	}

	for _, v := range tt {

		t.Run(v.name, func(t *testing.T) {

			rp, err := NewRocketPool(getTestRPCURL(t))
			require.NoError(t, err)

			calldata, err := rp.GenerateCalldata(v.action, v.opts)

			if v.hasError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, v.expected, calldata)
		})
	}
}
