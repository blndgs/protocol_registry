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

func TestVenusOperation_New(t *testing.T) {
	t.Run("unsupported chain", func(t *testing.T) {
		venus, err := NewVenusOperation(getTestClient(t, ChainETH),
			big.NewInt(1),
			common.HexToAddress("0xd933909A4a2b7A4638903028f44D1d38ce27c352"))
		require.Error(t, err)
		require.Nil(t, venus)
	})

	t.Run("venus correctly setup", func(t *testing.T) {
		venus, err := NewVenusOperation(getTestClient(t, ChainBSC),
			big.NewInt(56),
			common.HexToAddress("0xd933909A4a2b7A4638903028f44D1d38ce27c352"))
		require.NoError(t, err)
		require.NotNil(t, venus)
	})
}

func TestVenusOperation_Validate(t *testing.T) {
	t.Run("unsupported chain", func(t *testing.T) {
		venus, err := NewVenusOperation(getTestClient(t, ChainBSC),
			big.NewInt(56),
			common.HexToAddress("0xd933909A4a2b7A4638903028f44D1d38ce27c352"))
		require.NoError(t, err)

		err = venus.Validate(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})
		require.Error(t, err)
	})

	t.Run("unsupported action", func(t *testing.T) {
		venus, err := NewVenusOperation(getTestClient(t, ChainBSC),
			big.NewInt(56),
			common.HexToAddress("0xd933909A4a2b7A4638903028f44D1d38ce27c352"))
		require.NoError(t, err)

		err = venus.Validate(context.Background(), big.NewInt(56), NativeStake, TransactionParams{
			Amount: big.NewInt(1),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})
		require.Error(t, err)
	})

	t.Run("unsupported asset", func(t *testing.T) {
		venus, err := NewVenusOperation(getTestClient(t, ChainBSC),
			big.NewInt(56),
			common.HexToAddress("0xd933909A4a2b7A4638903028f44D1d38ce27c352"))
		require.NoError(t, err)

		err = venus.Validate(context.Background(), big.NewInt(56), LoanSupply, TransactionParams{
			Amount: big.NewInt(1),
			Asset:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
		})
		require.Error(t, err)
	})
}

func TestVenusOperation_GetBalance(t *testing.T) {
	client := getTestClient(t, ChainBSC)

	venus, err := NewVenusOperation(client,
		big.NewInt(56),
		common.HexToAddress("0xd933909A4a2b7A4638903028f44D1d38ce27c352"))
	require.NoError(t, err)

	t.Run("unsupported chain", func(t *testing.T) {
		_, _, err := venus.GetBalance(context.Background(), big.NewInt(1), emptyTestWallet, common.HexToAddress(""))
		require.Error(t, err)
	})

	t.Run("get balance for address", func(t *testing.T) {
		token, balance, err := venus.GetBalance(context.Background(), big.NewInt(56), emptyTestWallet, common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"))
		require.NoError(t, err)
		require.NotNil(t, balance)
		require.NotEmpty(t, token)
	})
}

func TestVenusOperation_GenerateCalldata(t *testing.T) {
	venus, err := NewVenusOperation(getTestClient(t, ChainBSC),
		big.NewInt(56),
		common.HexToAddress("0xd933909A4a2b7A4638903028f44D1d38ce27c352"))
	require.NoError(t, err)

	t.Run("unsupported chain", func(t *testing.T) {
		_, err := venus.GenerateCalldata(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{})
		require.Error(t, err)
	})

	t.Run("unsupported action", func(t *testing.T) {
		_, err := venus.GenerateCalldata(context.Background(), big.NewInt(56), NativeStake, TransactionParams{})
		require.Error(t, err)
	})

	t.Run("generate supply calldata with WBNB", func(t *testing.T) {
		params := TransactionParams{
			Amount: big.NewInt(100000000000000000), // 0.1 BNB
			Asset:  common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"),
		}
		calldata, err := venus.GenerateCalldata(context.Background(), big.NewInt(56), LoanSupply, params)
		require.NoError(t, err)
		require.NotEmpty(t, calldata)
	})

	t.Run("generate supply calldata with native BNB", func(t *testing.T) {
		params := TransactionParams{
			Amount: big.NewInt(100000000000000000), // 0.1 BNB
			Asset:  common.HexToAddress(nativeDenomAddress),
		}
		calldata, err := venus.GenerateCalldata(context.Background(), big.NewInt(56), LoanSupply, params)
		require.NoError(t, err)
		require.NotEmpty(t, calldata)
	})

	t.Run("generate supply calldata with recipient", func(t *testing.T) {
		recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")
		params := TransactionParams{
			Amount:    big.NewInt(100000000000000000), // 0.1 BNB
			Asset:     common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"),
			Recipient: recipient,
		}
		calldata, err := venus.GenerateCalldata(context.Background(), big.NewInt(56), LoanSupply, params)
		require.NoError(t, err)
		require.NotEmpty(t, calldata)
	})

	t.Run("generate withdraw calldata with WBNB", func(t *testing.T) {
		params := TransactionParams{
			Amount: big.NewInt(100000000000000000), // 0.1 BNB
			Asset:  common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"),
		}
		calldata, err := venus.GenerateCalldata(context.Background(), big.NewInt(56), LoanWithdraw, params)
		require.NoError(t, err)
		require.NotEmpty(t, calldata)
	})

	t.Run("generate withdraw calldata with native BNB", func(t *testing.T) {
		params := TransactionParams{
			Amount: big.NewInt(100000000000000000), // 0.1 BNB
			Asset:  common.HexToAddress(nativeDenomAddress),
		}
		calldata, err := venus.GenerateCalldata(context.Background(), big.NewInt(56), LoanWithdraw, params)
		require.NoError(t, err)
		require.NotEmpty(t, calldata)
	})

	t.Run("generate withdraw calldata with recipient", func(t *testing.T) {
		recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")
		params := TransactionParams{
			Amount:    big.NewInt(100000000000000000), // 0.1 BNB
			Asset:     common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"),
			Recipient: recipient,
		}
		calldata, err := venus.GenerateCalldata(context.Background(), big.NewInt(56), LoanWithdraw, params)
		require.NoError(t, err)
		require.NotEmpty(t, calldata)
	})
}

func TestVenus(t *testing.T) {

	t.Run("Liquid staking pool", func(t *testing.T) {

		venus, err := NewVenusOperation(getTestClient(t, ChainBSC),
			big.NewInt(56),
			common.HexToAddress("0xd933909A4a2b7A4638903028f44D1d38ce27c352"))
		require.NoError(t, err)

		tt := []struct {
			asset       string
			name        string
			unsupported bool
		}{
			{asset: "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c", name: "WBNB is supported"},
			{asset: nativeDenomAddress, name: "BNB is supported"},
			{asset: "0x55d398326f99059fF775485246999027B3197955", name: "USDT is not supported", unsupported: true},
		}

		_, err = venus.GetSupportedAssets(context.Background(), big.NewInt(56))
		require.NoError(t, err)

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				if tc.unsupported {
					require.False(t, venus.IsSupportedAsset(context.Background(), big.NewInt(56), common.HexToAddress(tc.asset)))
					return
				}

				require.True(t, venus.IsSupportedAsset(context.Background(), big.NewInt(56), common.HexToAddress(tc.asset)))
			})
		}
	})

	t.Run("Stablecoins pool", func(t *testing.T) {

		venus, err := NewVenusOperation(getTestClient(t, ChainBSC), big.NewInt(56), common.HexToAddress("0x94c1495cD4c557f1560Cbd68EAB0d197e6291571"))
		require.NoError(t, err)

		tt := []struct {
			asset       string
			name        string
			unsupported bool
		}{
			{asset: "0x55d398326f99059fF775485246999027B3197955", name: "USDT is supported"},
			{asset: "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c", name: "WBNB is not supported", unsupported: true},
			{asset: nativeDenomAddress, name: "BNB is not supported", unsupported: true},
		}

		_, err = venus.GetSupportedAssets(context.Background(), big.NewInt(56))
		require.NoError(t, err)

		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				if tc.unsupported {
					require.False(t, venus.IsSupportedAsset(context.Background(), big.NewInt(56), common.HexToAddress(tc.asset)))
					return
				}

				require.True(t, venus.IsSupportedAsset(context.Background(), big.NewInt(56), common.HexToAddress(tc.asset)))
			})
		}
	})
}
