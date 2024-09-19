//go:build integration
// +build integration

package pkg

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestCompoundV3_New(t *testing.T) {

	t.Run("unsupported chain", func(t *testing.T) {
		compoundImpl, err := NewCompoundOperation(getTestClient(t, ChainETH), big.NewInt(100),
			common.HexToAddress("0xa17581a9e3356d9a858b789d68b4d866e593ae94"))

		require.Error(t, err)
		require.Nil(t, compoundImpl)
	})

	t.Run("unsupported pool market", func(t *testing.T) {

		compoundImpl, err := NewCompoundOperation(getTestClient(t, ChainETH), big.NewInt(1),
			common.HexToAddress(nativeDenomAddress))

		require.Error(t, err)
		require.Nil(t, compoundImpl)
	})

	t.Run("compund correctly setup", func(t *testing.T) {
		_, err := NewCompoundOperation(getTestClient(t, ChainETH), big.NewInt(1),
			common.HexToAddress("0xc3d688b66703497daa19211eedff47f25384cdc3"))
		require.NoError(t, err)
	})
}

func TestCompound_GenerateCalldata_Supply(t *testing.T) {

	// cast calldata "supply(address,uint256)" 0x514910771AF9Ca656af840dff83E8264EcF986CA 1000000000000000000
	// 0xf2b9fdb8000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000
	expectedCalldata := "0xf2b9fdb8000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000"

	compoundClient, err := NewCompoundOperation(getTestClient(t, ChainETH), big.NewInt(1),
		common.HexToAddress("0xc3d688b66703497daa19211eedff47f25384cdc3"))
	require.NoError(t, err)

	calldata, err := compoundClient.GenerateCalldata(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
		Asset:  common.HexToAddress("0x514910771AF9Ca656af840dff83E8264EcF986CA"),
		Amount: big.NewInt(1e18),
	})

	require.NoError(t, err)
	require.Equal(t, expectedCalldata, calldata)
}

func TestCompound_GenerateCalldata_Withdraw(t *testing.T) {

	// cast calldata "withdraw(address,uint256)" 0x514910771AF9Ca656af840dff83E8264EcF986CA 1000000000000000000
	// 0xf3fef3a3000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a764000
	expectedCalldata := "0xf3fef3a3000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000"

	compoundClient, err := NewCompoundOperation(getTestClient(t, ChainETH), big.NewInt(1),
		common.HexToAddress("0xc3d688b66703497daa19211eedff47f25384cdc3"))
	require.NoError(t, err)

	calldata, err := compoundClient.GenerateCalldata(context.Background(), big.NewInt(1),
		LoanWithdraw, TransactionParams{
			Asset:  common.HexToAddress("0x514910771AF9Ca656af840dff83E8264EcF986CA"),
			Amount: big.NewInt(1 * 1e18),
		})
	require.NoError(t, err)
	require.Equal(t, expectedCalldata, calldata)
}

func TestCompound_IsSupportedAsset(t *testing.T) {

	compoundImpl, err := NewCompoundOperation(getTestClient(t, ChainETH), big.NewInt(1),
		common.HexToAddress("0xa17581a9e3356d9a858b789d68b4d866e593ae94"))

	require.NoError(t, err)

	t.Run("unsupported chain", func(t *testing.T) {
		supported := compoundImpl.IsSupportedAsset(context.TODO(), big.NewInt(100), common.HexToAddress(nativeDenomAddress))
		require.False(t, supported)
	})

	t.Run("native asset supported", func(t *testing.T) {
		supported := compoundImpl.IsSupportedAsset(context.TODO(), big.NewInt(1), common.HexToAddress(nativeDenomAddress))
		require.False(t, supported)
	})

	t.Run("rEth  asset supported", func(t *testing.T) {
		supported := compoundImpl.IsSupportedAsset(context.TODO(), big.NewInt(1), common.HexToAddress("0xae78736Cd615f374D3085123A210448E74Fc6393"))
		require.True(t, supported)
	})
}

func TestGetCTokens(t *testing.T) {

	client := getTestClient(t, ChainETH)

	tokens, err := getCTokens(client, common.HexToAddress(""))
	require.NoError(t, err)

	tt := []struct {
		name            string
		expectedCToken  string
		underlyingToken string
		hasError        bool
	}{
		{
			name:            "Uniswap",
			expectedCToken:  "0x35A18000230DA775CAc24873d00Ff85BccdeD550",
			underlyingToken: "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984",
			hasError:        false,
		},
		{
			name:            "BAT",
			expectedCToken:  "0x6C8c6b02E7b2BE14d4fA6022Dfd6d75921D90E4E",
			underlyingToken: "0x0D8775F648430679A709E98d2b0Cb6250d2887EF",
			hasError:        false,
		},
		{
			name:            "ETH",
			expectedCToken:  "0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5",
			underlyingToken: common.HexToAddress(nativeDenomAddress).Hex(),
			hasError:        false,
		},
		{
			name:            "Shiba INU",
			expectedCToken:  "",
			underlyingToken: "0x95aD61b0a150d79219dCF64E1E6Cc01f0B64C4cE",
			hasError:        true,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			cToken, ok := tokens[v.underlyingToken]
			require.Equal(t, !v.hasError, ok)

			if !v.hasError {
				return
			}

			require.Equal(t, v.expectedCToken, cToken)
		})
	}
}

func TestCompound_GetBalance(t *testing.T) {

	client := getTestClient(t, ChainETH)

	compoundImpl, err := NewCompoundOperation(client, big.NewInt(1),
		common.HexToAddress("0xa17581a9e3356d9a858b789d68b4d866e593ae94"))

	require.NoError(t, err)

	t.Run("fetch uni cToken balance", func(t *testing.T) {

		uniAsset := common.HexToAddress("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984")

		token, bal, err := compoundImpl.GetBalance(context.Background(), big.NewInt(1),
			common.HexToAddress("0x94fa8efDD58e1721ad8Bf5D4001060e0E1C4d58e"), uniAsset)

		require.NoError(t, err)
		require.NotNil(t, bal)

		validateSymbolFromToken(t, client, token, "cUNI")
	})

	t.Run("fetch unsupported token balance", func(t *testing.T) {

		rethToken := common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393")

		_, _, err := compoundImpl.GetBalance(context.Background(), big.NewInt(1),
			common.HexToAddress("0x94fa8efDD58e1721ad8Bf5D4001060e0E1C4d58e"), rethToken)

		require.Error(t, err)
	})
}

func TestCompoundV3_Validate_ETH_Market(t *testing.T) {

	compoundImpl, err := NewCompoundOperation(getTestClient(t, ChainETH), big.NewInt(1),
		common.HexToAddress("0xa17581a9e3356d9a858b789d68b4d866e593ae94"))

	require.NoError(t, err)
	require.NotNil(t, compoundImpl)

	tt := []struct {
		name     string
		address  string
		hasError bool
	}{
		{
			name:     "wbtc cannot be supplied",
			address:  "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
			hasError: true,
		},
	}

	for _, v := range tt {
		err = compoundImpl.Validate(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
			Amount: big.NewInt(1 * 1e8),
			Asset:  common.HexToAddress(v.address),
			Sender: hotWallet,
		})

		if v.hasError {
			require.Error(t, err)
			continue
		}

		require.NoError(t, err)
	}
}
