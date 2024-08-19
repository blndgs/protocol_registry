package pkg

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

var hotWallet = common.HexToAddress("0xee5b5b923ffce93a870b3104b7ca09c3db80047a") // bybit hot wallet

func TestAave_IsSupportedAsset(t *testing.T) {

	aave, err := NewAaveOperation(getTestClient(t), big.NewInt(1), AaveProtocolForkAave)
	require.NoError(t, err)

	sparklend, err := NewAaveOperation(getTestClient(t), big.NewInt(1), AaveProtocolForkSpark)
	require.NoError(t, err)

	t.Run("(aave) Lido stETH not supported", func(t *testing.T) {
		require.False(t, aave.IsSupportedAsset(context.Background(),
			big.NewInt(1), common.HexToAddress("0xE95A203B1a91a908F9B9CE46459d101078c2c3cb")))
	})

	t.Run("(aave) wrapped btc supported ", func(t *testing.T) {
		require.True(t, aave.IsSupportedAsset(context.Background(),
			big.NewInt(1), common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599")))
	})

	t.Run("(sparklend) wrapped btc supported ", func(t *testing.T) {
		require.True(t, sparklend.IsSupportedAsset(context.Background(),
			big.NewInt(1), common.HexToAddress("0x2260fac5e5542a773aa44fbcfedf7c193bc2c599")))
	})

	t.Run("(sparklend) aave token not supported", func(t *testing.T) {
		require.False(t, sparklend.IsSupportedAsset(context.Background(),
			big.NewInt(1), common.HexToAddress("0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9")))
	})
}

func TestAave_GetAToken(t *testing.T) {

	tt := []struct {
		name                  string
		asset, expectedAToken common.Address
		fork                  AaveProtocolFork
	}{
		{
			asset:          common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393"),
			name:           "Rocketpool ETH on Aave",
			expectedAToken: common.HexToAddress("0xCc9EE9483f662091a1de4795249E24aC0aC2630f"),
			fork:           AaveProtocolForkAave,
		},
		{
			name:           "USDC on Aave",
			asset:          common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
			expectedAToken: common.HexToAddress("0x98C23E9d8f34FEFb1B7BD6a91B7FF122F4e16F5c"),
			fork:           AaveProtocolForkAave,
		},
		{
			name:           "USDC on sparklend",
			asset:          common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
			expectedAToken: common.HexToAddress("0x377C3bd93f2a2984E1E7bE6A5C22c525eD4A4815"),
			fork:           AaveProtocolForkSpark,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {

			protocol, err := NewAaveOperation(getTestClient(t), big.NewInt(1), v.fork)
			require.NoError(t, err)

			aToken, err := protocol.getAToken(context.Background(), v.asset)
			require.NoError(t, err)

			require.Equal(t, v.expectedAToken, aToken)
		})
	}
}

func TestAave_Validate(t *testing.T) {

	aave, err := NewAaveOperation(getTestClient(t), big.NewInt(1), AaveProtocolForkAave)
	require.NoError(t, err)

	t.Run("zero value supplied", func(t *testing.T) {

		err = aave.Validate(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393"),
			Sender: common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("zero value supplied when withdrawing. atoken balance not enough", func(t *testing.T) {

		err = aave.Validate(context.Background(), big.NewInt(1), LoanWithdraw, TransactionParams{
			Amount: big.NewInt(1),
			Asset:  common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393"),
			Sender: common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("unsupported chain", func(t *testing.T) {

		err = aave.Validate(context.Background(), big.NewInt(100), LoanSupply, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393"),
			Sender: common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("unsupported action", func(t *testing.T) {

		err = aave.Validate(context.Background(), big.NewInt(100), NativeStake, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393"),
			Sender: common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("user without usdt balance cannot supply", func(t *testing.T) {

		err = aave.Validate(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
			Amount: big.NewInt(100000),
			Asset:  common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393"),
			Sender: common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("user with usdt balance can supply", func(t *testing.T) {

		err = aave.Validate(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
			Amount: big.NewInt(100000000),
			Sender: hotWallet,
			Asset:  common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"),
		})

		require.NoError(t, err)
	})

	t.Run("(sparklend) user with usdt balance can supply", func(t *testing.T) {

		aave, err := NewAaveOperation(getTestClient(t), big.NewInt(1), AaveProtocolForkSpark)
		require.NoError(t, err)

		err = aave.Validate(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
			Amount: big.NewInt(100000000),
			Sender: hotWallet,
			Asset:  common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"),
		})

		require.NoError(t, err)
	})
}

func TestAave_GetBalance(t *testing.T) {

	aave, err := NewAaveOperation(getTestClient(t), big.NewInt(1), AaveProtocolForkAave)
	require.NoError(t, err)

	bal, err := aave.GetBalance(context.Background(), big.NewInt(1), hotWallet,
		common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"))

	require.NoError(t, err)
	require.NotNil(t, bal)
}

func TestAave_GenerateCalldata_Withdraw(t *testing.T) {
	// cast calldata "withdraw(address,uint256,address)" 0xc0ffee254729296a45a3885639AC7E10F9d54979 500000000000000000 0x0000000000000000000000000000000000000000
	// 0x69328dec000000000000000000000000c0ffee254729296a45a3885639ac7e10f9d5497900000000000000000000000000000000000000000000000006f05b59d3b200000000000000000000000000000000000000000000000000000000000000000000

	expectedCalldata := "0x69328dec000000000000000000000000c0ffee254729296a45a3885639ac7e10f9d5497900000000000000000000000000000000000000000000000006f05b59d3b200000000000000000000000000000000000000000000000000000000000000000000"

	aave, err := NewAaveOperation(getTestClient(t), big.NewInt(1), AaveProtocolForkAave)
	require.NoError(t, err)

	calldata, err := aave.GenerateCalldata(context.Background(), big.NewInt(1), LoanWithdraw, TransactionParams{
		Amount: big.NewInt(500000000000000000),
		Sender: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		Asset:  common.HexToAddress("0xc0ffee254729296a45a3885639AC7E10F9d54979"),
		ExtraData: map[string]interface{}{
			"referral_code": 0,
		},
	})

	require.NoError(t, err)
	require.Equal(t, expectedCalldata, calldata)
}

func TestAave_GenerateCalldata_Supply(t *testing.T) {
	// cast calldata "supply(address,uint256,address,uint16)" 0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984 1000000000000000000 0x0000000000000000000000000000000000000000 10
	// 0x617ba0370000000000000000000000001f9840a85d5af5bf1d1762f925bdaddc4201f9840000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a

	expectedCalldata := "0x617ba0370000000000000000000000001f9840a85d5af5bf1d1762f925bdaddc4201f9840000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a"

	aave, err := NewAaveOperation(getTestClient(t), big.NewInt(1), AaveProtocolForkAave)
	require.NoError(t, err)

	calldata, err := aave.GenerateCalldata(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
		Asset:  common.HexToAddress("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984"),
		Amount: big.NewInt(1000000000000000000),
		Sender: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		ExtraData: map[string]interface{}{
			"referral_code": uint16(10),
		},
	})
	require.NoError(t, err)

	require.Equal(t, expectedCalldata, calldata)
}
