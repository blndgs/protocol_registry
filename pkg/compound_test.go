package pkg

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestCompoundV3_New(t *testing.T) {

	t.Run("unsupported chain", func(t *testing.T) {
		// wrong chain
		compoundImpl, err := NewCompoundV3(big.NewInt(800), common.HexToAddress("0xc3d688b66703497daa19211eedff47f25384cdc3"),
			NativeStake)
		require.Error(t, err)

		require.Nil(t, compoundImpl)
	})

	t.Run("unsupported pool market", func(t *testing.T) {

		compoundImpl, err := NewCompoundV3(big.NewInt(1), common.HexToAddress(nativeDenomAddress), LoanSupply)
		require.Error(t, err)

		require.Nil(t, compoundImpl)
	})

	t.Run("unsupported action", func(t *testing.T) {
		compoundImpl, err := NewCompoundV3(big.NewInt(1), common.HexToAddress("0xc3d688b66703497daa19211eedff47f25384cdc3"),
			NativeStake)
		require.Error(t, err)

		require.Nil(t, compoundImpl)
	})

}

func TestCompoundV3_GenerateCallData(t *testing.T) {

	tt := []struct {
		name     string
		chainID  *big.Int
		action   ContractAction
		expected string
		args     []interface{}
		hasError bool
	}{
		{
			name:    "Supply action",
			action:  LoanSupply,
			chainID: big.NewInt(1),
			// cast calldata "supply(address,uint256)" 0x514910771AF9Ca656af840dff83E8264EcF986CA 1000000000000000000
			// 0xf2b9fdb8000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000
			expected: "0xf2b9fdb8000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000",
			args: []interface{}{
				common.HexToAddress("0x514910771AF9Ca656af840dff83E8264EcF986CA"),
				big.NewInt(1 * 1e18),
			},
		},
		{
			name:    "Withdraw action",
			action:  LoanWithdraw,
			chainID: big.NewInt(1),
			// cast calldata "withdraw(address,uint256)" 0x514910771AF9Ca656af840dff83E8264EcF986CA 1000000000000000000
			// 0xf3fef3a3000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a764000
			expected: "0xf3fef3a3000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000",
			args: []interface{}{
				common.HexToAddress("0x514910771AF9Ca656af840dff83E8264EcF986CA"),
				big.NewInt(1 * 1e18),
			},
		},
	}

	for _, v := range tt {

		t.Run(v.name, func(t *testing.T) {

			compoundImpl, err := NewCompoundV3(v.chainID, common.HexToAddress("0xc3d688b66703497daa19211eedff47f25384cdc3"),
				v.action)
			require.NoError(t, err)

			calldata, err := compoundImpl.GenerateCalldata(v.args)

			if v.hasError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, v.expected, calldata)
		})
	}
}
