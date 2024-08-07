package pkg

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestCompoundV3_Validate_ETH_Market(t *testing.T) {

	compoundImpl, err := NewCompoundV3(big.NewInt(1), common.HexToAddress("0xa17581a9e3356d9a858b789d68b4d866e593ae94"))

	require.NoError(t, err)
	require.NotNil(t, compoundImpl)

	tt := []struct {
		name     string
		address  string
		hasError bool
	}{
		{
			name:    "CBeth can be supplied",
			address: "0xBe9895146f7AF43049ca1c1AE358B0541Ea49704",
		},
		{
			name:    "lido wrapped eth can be supplied",
			address: "0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0",
		},
		{
			name:    "rEth can be supplied",
			address: "0xae78736Cd615f374D3085123A210448E74Fc6393",
		},
		{
			name:     "wbtc cannot be supplied",
			address:  "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
			hasError: true,
		},
	}

	for _, v := range tt {

		err := compoundImpl.Validate(common.HexToAddress(v.address))

		if v.hasError {
			require.Error(t, err)
			continue
		}

		require.NoError(t, err)
	}
}

func TestCompoundV3_New(t *testing.T) {

	t.Run("unsupported chain", func(t *testing.T) {
		// wrong chain
		compoundImpl, err := NewCompoundV3(big.NewInt(800),
			common.HexToAddress("0xc3d688b66703497daa19211eedff47f25384cdc3"))
		require.Error(t, err)

		require.Nil(t, compoundImpl)
	})

	t.Run("unsupported pool market", func(t *testing.T) {

		compoundImpl, err := NewCompoundV3(big.NewInt(1),
			common.HexToAddress(nativeDenomAddress))
		require.Error(t, err)

		require.Nil(t, compoundImpl)
	})
}

func TestCompoundV3_GenerateCallData(t *testing.T) {

	tt := []struct {
		name     string
		chainID  *big.Int
		expected string
		opts     GenerateCalldataOptions
		hasError bool
		action   ContractAction
	}{
		{
			name:    "Supply action",
			chainID: big.NewInt(1),
			// cast calldata "supply(address,uint256)" 0x514910771AF9Ca656af840dff83E8264EcF986CA 1000000000000000000
			// 0xf2b9fdb8000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000
			expected: "0xf2b9fdb8000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000",
			opts: GenerateCalldataOptions{
				Asset:  common.HexToAddress("0x514910771AF9Ca656af840dff83E8264EcF986CA"),
				Amount: big.NewInt(1e18),
			},
			action: LoanSupply,
		},
		{
			name:    "Withdraw action",
			chainID: big.NewInt(1),
			// cast calldata "withdraw(address,uint256)" 0x514910771AF9Ca656af840dff83E8264EcF986CA 1000000000000000000
			// 0xf3fef3a3000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a764000
			expected: "0xf3fef3a3000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000",
			opts: GenerateCalldataOptions{
				Asset:  common.HexToAddress("0x514910771AF9Ca656af840dff83E8264EcF986CA"),
				Amount: big.NewInt(1 * 1e18),
			},
			action: LoanWithdraw,
		},
		{
			name:    "Unsupported action",
			chainID: big.NewInt(1),
			// cast calldata "withdraw(address,uint256)" 0x514910771AF9Ca656af840dff83E8264EcF986CA 1000000000000000000
			// 0xf3fef3a3000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a764000
			expected: "0xf3fef3a3000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000",
			opts: GenerateCalldataOptions{
				Asset:  common.HexToAddress("0x514910771AF9Ca656af840dff83E8264EcF986CA"),
				Amount: big.NewInt(1 * 1e18),
			},
			action: LoanWithdraw,
		},
	}

	for _, v := range tt {

		t.Run(v.name, func(t *testing.T) {

			compoundImpl, err := NewCompoundV3(v.chainID,
				common.HexToAddress("0xc3d688b66703497daa19211eedff47f25384cdc3"))
			require.NoError(t, err)

			calldata, err := compoundImpl.GenerateCalldata(v.action, v.opts)

			if v.hasError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, v.expected, calldata)
		})
	}
}
