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

func TestBinanceWrappedEthOperation_New(t *testing.T) {
	t.Run("unsupported chain", func(t *testing.T) {
		binance, err := NewBinanceWrappedEthOperation(getTestClient(t, ChainETH), big.NewInt(1))
		require.Error(t, err)
		require.Nil(t, binance)
	})

	t.Run("binance correctly setup", func(t *testing.T) {
		binance, err := NewBinanceWrappedEthOperation(getTestClient(t, ChainBSC), big.NewInt(56))
		require.NoError(t, err)
		require.NotNil(t, binance)
	})
}

func TestBinanceWrappedEthOperation_Validate(t *testing.T) {
	t.Run("unsupported chain", func(t *testing.T) {
		binance, err := NewBinanceWrappedEthOperation(getTestClient(t, ChainBSC), big.NewInt(56))
		require.NoError(t, err)

		err = binance.Validate(context.Background(), big.NewInt(1), NativeStake, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})
		require.Error(t, err)
	})

	t.Run("unsupported action", func(t *testing.T) {
		binance, err := NewBinanceWrappedEthOperation(getTestClient(t, ChainBSC), big.NewInt(56))
		require.NoError(t, err)

		err = binance.Validate(context.Background(), big.NewInt(56), LoanSupply, TransactionParams{
			Amount: big.NewInt(1),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})
		require.Error(t, err)
	})

	t.Run("unsupported asset", func(t *testing.T) {
		binance, err := NewBinanceWrappedEthOperation(getTestClient(t, ChainBSC), big.NewInt(56))
		require.NoError(t, err)

		err = binance.Validate(context.Background(), big.NewInt(56), NativeStake, TransactionParams{
			Amount: big.NewInt(1),
			Asset:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
		})
		require.Error(t, err)
	})
}

func TestBinanceWrappedEthOperation_GetBalance(t *testing.T) {
	client := getTestClient(t, ChainBSC)

	binance, err := NewBinanceWrappedEthOperation(client, big.NewInt(56))
	require.NoError(t, err)

	t.Run("unsupported chain", func(t *testing.T) {
		_, _, err := binance.GetBalance(context.Background(), big.NewInt(1), emptyTestWallet, common.HexToAddress(""))
		require.Error(t, err)
	})

	t.Run("get balance for address", func(t *testing.T) {
		token, balance, err := binance.GetBalance(context.Background(), big.NewInt(56), emptyTestWallet, common.HexToAddress(""))
		require.NoError(t, err)
		require.NotNil(t, balance)
		require.Equal(t, BinanceStakedETHBNBContractAddress, token)
	})
}

func TestBinanceWrappedEthOperation_GenerateCalldata(t *testing.T) {
	binance, err := NewBinanceWrappedEthOperation(getTestClient(t, ChainBSC), big.NewInt(56))
	require.NoError(t, err)

	t.Run("unsupported chain", func(t *testing.T) {
		_, err := binance.GenerateCalldata(context.Background(), big.NewInt(1), NativeStake, TransactionParams{})
		require.Error(t, err)
	})

	t.Run("unsupported action", func(t *testing.T) {
		_, err := binance.GenerateCalldata(context.Background(), big.NewInt(56), LoanSupply, TransactionParams{})
		require.Error(t, err)
	})

	t.Run("generate stake calldata", func(t *testing.T) {
		params := TransactionParams{
			Amount: big.NewInt(100000000000000000), // 0.1 ETH
			ExtraData: map[string]interface{}{
				"referral_address": common.HexToAddress(nativeDenomAddress),
			},
		}
		calldata, err := binance.GenerateCalldata(context.Background(), big.NewInt(56), NativeStake, params)
		require.NoError(t, err)
		require.NotEmpty(t, calldata)
		t.Log(calldata)
	})
}

func TestBinanceWrappedEthOperation_IsSupportedAsset(t *testing.T) {
	binance, err := NewBinanceWrappedEthOperation(getTestClient(t, ChainBSC), big.NewInt(56))
	require.NoError(t, err)

	t.Run("unsupported chain", func(t *testing.T) {
		isSupported := binance.IsSupportedAsset(context.Background(), big.NewInt(1), common.HexToAddress(nativeDenomAddress))
		require.False(t, isSupported)
	})

	t.Run("native token", func(t *testing.T) {
		isSupported := binance.IsSupportedAsset(context.Background(), big.NewInt(56), common.HexToAddress(nativeDenomAddress))
		require.False(t, isSupported)
	})

	t.Run("unsupported token", func(t *testing.T) {
		isSupported := binance.IsSupportedAsset(context.Background(), big.NewInt(56), common.HexToAddress("0x1234567890123456789012345678901234567890"))
		require.False(t, isSupported)
	})

	t.Run("supported token", func(t *testing.T) {
		isSupported := binance.IsSupportedAsset(
			context.Background(),
			big.NewInt(56),
			bnbBinanceWrappedETHOperationContract,
		)
		require.True(t, isSupported)
	})
}
