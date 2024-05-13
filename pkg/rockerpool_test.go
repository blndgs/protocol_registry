package pkg

import (
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func getTestRPCURL(t *testing.T) string {
	t.Helper()

	u := os.Getenv("TEST_ETH_RPC_URL")

	require.NotEmpty(t, u)
	return u
}

func TestRocketPoolOperation_GenerateCallData_UnsupportedAction(t *testing.T) {

	rp, err := NewRocketPool(getTestRPCURL(t), SupplyAction)
	require.Error(t, err)

	require.Nil(t, rp)
}

func TestRocketPoolOperation_GenerateCallData(t *testing.T) {

	tt := []struct {
		name     string
		action   ContractAction
		expected string
		args     []interface{}
	}{
		{
			name:   "Supply action",
			action: SubmitAction,
			// cast calldata "deposit()"
			// 0xd0e30db0
			expected: "0xd0e30db0",
			args: []interface{}{
				big.NewInt(1 * 1e6),
			},
		},
		{
			name:   "Withdraw action",
			action: WithdrawAction,
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

			rp, err := NewRocketPool(getTestRPCURL(t), v.action)
			require.NoError(t, err)

			rp.Register(registry)

			calldata, err := rp.GenerateCalldata(StakeKind, v.args)

			require.NoError(t, err)

			require.Equal(t, v.expected, calldata)
		})
	}
}
