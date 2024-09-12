//go:build integration
// +build integration

package pkg

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

var hotWallet = common.HexToAddress("0xee5b5b923ffce93a870b3104b7ca09c3db80047a") // bybit hot wallet

func TestAave_New(t *testing.T) {

	t.Run("unsupported chain", func(t *testing.T) {
		_, err := NewAaveOperation(getTestClient(t, ChainETH), big.NewInt(100), AaveProtocolForkAave)
		require.Error(t, err)
		require.Contains(t, err.Error(), "only eth and bnb chains are supported")
	})

	t.Run("spark finance is not supported on bnb chain", func(t *testing.T) {
		_, err := NewAaveOperation(getTestClient(t, ChainETH), big.NewInt(56), AaveProtocolForkSpark)
		require.Error(t, err)
		require.Contains(t, err.Error(), "spark finance is not supported on Bnb chain")
	})

	t.Run("network id check fails", func(t *testing.T) {
		_, err := NewAaveOperation(getTestClient(t, ChainBSC), big.NewInt(1), AaveProtocolForkAave)
		require.Error(t, err)
		require.Contains(t, err.Error(), "network id of client")
	})

	t.Run("network id of bsc network client does not match eth chain", func(t *testing.T) {
		_, err := NewAaveOperation(getTestClient(t, ChainBSC), big.NewInt(1), AaveProtocolForkAave)
		require.Error(t, err)
		require.Contains(t, err.Error(), "network id of client")
	})
}

func TestAave_GetSupportedAsset(t *testing.T) {

	aave, err := NewAaveOperation(getTestClient(t, ChainETH), big.NewInt(1), AaveProtocolForkAave)
	require.NoError(t, err)

	sparklend, err := NewAaveOperation(getTestClient(t, ChainETH), big.NewInt(1), AaveProtocolForkSpark)
	require.NoError(t, err)

	t.Run("aave on eth", func(t *testing.T) {
		assets, err := aave.GetSupportedAssets(context.Background(), big.NewInt(1))
		require.NoError(t, err)
		require.NotEmpty(t, assets)

		// WETH
		require.Contains(t, assets, common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"))
	})

	t.Run("sparklend on eth", func(t *testing.T) {
		assets, err := sparklend.GetSupportedAssets(context.Background(), big.NewInt(1))
		require.NoError(t, err)
		require.NotEmpty(t, assets)
	})

	t.Run("aave on bsc", func(t *testing.T) {

		aave, err := NewAaveOperation(getTestClient(t, ChainBSC), big.NewInt(56), AaveProtocolForkAave)
		require.NoError(t, err)

		assets, err := aave.GetSupportedAssets(context.Background(), big.NewInt(56))
		require.NoError(t, err)
		require.NotEmpty(t, assets)
		// WBNB
		require.Contains(t, assets, common.HexToAddress("0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"))
	})

	t.Run("avalon finance on bsc", func(t *testing.T) {

		avalonFinance, err := NewAaveOperation(getTestClient(t, ChainBSC), big.NewInt(56), AaveProtocolForkAvalonFinance)
		require.NoError(t, err)

		assets, err := avalonFinance.GetSupportedAssets(context.Background(), big.NewInt(56))
		require.NoError(t, err)
		require.NotEmpty(t, assets)
		// SolvBTC
		require.Contains(t, assets, common.HexToAddress("0x4aae823a6a0b376De6A78e74eCC5b079d38cBCf7"))
	})
}

func TestAave_IsSupportedAsset(t *testing.T) {

	aave, err := NewAaveOperation(getTestClient(t, ChainETH), big.NewInt(1), AaveProtocolForkAave)
	require.NoError(t, err)

	sparklend, err := NewAaveOperation(getTestClient(t, ChainETH), big.NewInt(1), AaveProtocolForkSpark)
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
		client                *ethclient.Client
	}{
		{
			asset:          common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393"),
			name:           "Rocketpool ETH on Aave",
			expectedAToken: common.HexToAddress("0xCc9EE9483f662091a1de4795249E24aC0aC2630f"),
			fork:           AaveProtocolForkAave,
			client:         getTestClient(t, ChainETH),
		},
		{
			asset:          common.HexToAddress("0x55d398326f99059fF775485246999027B3197955"),
			name:           "USDT on Aave (BSC)",
			expectedAToken: common.HexToAddress("0xa9251ca9DE909CB71783723713B21E4233fbf1B1"),
			fork:           AaveProtocolForkAave,
			client:         getTestClient(t, ChainBSC),
		},
		{
			name:           "USDC on Aave",
			asset:          common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
			expectedAToken: common.HexToAddress("0x98C23E9d8f34FEFb1B7BD6a91B7FF122F4e16F5c"),
			fork:           AaveProtocolForkAave,
			client:         getTestClient(t, ChainETH),
		},
		{
			name:           "USDC on sparklend",
			asset:          common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
			expectedAToken: common.HexToAddress("0x377C3bd93f2a2984E1E7bE6A5C22c525eD4A4815"),
			fork:           AaveProtocolForkSpark,
			client:         getTestClient(t, ChainETH),
		},
		{
			name:           "USDC on AvalonFinance",
			asset:          common.HexToAddress("0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d"),
			expectedAToken: common.HexToAddress("0xfcefbD84BA5d64cd530Afb2e8DDEa7b399A9fC53"),
			fork:           AaveProtocolForkAvalonFinance,
			client:         getTestClient(t, ChainBSC),
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {

			id, err := v.client.NetworkID(context.Background())
			require.NoError(t, err)

			protocol, err := NewAaveOperation(v.client, id, v.fork)
			require.NoError(t, err)

			aToken, err := protocol.getAToken(context.Background(), v.asset)
			require.NoError(t, err)

			require.Equal(t, v.expectedAToken, aToken)
		})
	}
}

func TestAave_Validate(t *testing.T) {

	aave, err := NewAaveOperation(getTestClient(t, ChainETH), big.NewInt(1), AaveProtocolForkAave)
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

	t.Run("user without balance balance cannot withdraw", func(t *testing.T) {

		err = aave.Validate(context.Background(), big.NewInt(1), LoanWithdraw, TransactionParams{
			Amount: big.NewInt(1),
			Sender: common.HexToAddress("0xFc21d6d146E6086B8359705C8b28512a983db0cb"),
			Asset:  common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"),
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

		aave, err := NewAaveOperation(getTestClient(t, ChainETH), big.NewInt(1), AaveProtocolForkSpark)
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

	client := getTestClient(t, ChainETH)

	aave, err := NewAaveOperation(client, big.NewInt(1), AaveProtocolForkAave)
	require.NoError(t, err)

	token, bal, err := aave.GetBalance(context.Background(), big.NewInt(1), hotWallet,
		common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"))

	require.NoError(t, err)
	require.NotNil(t, bal)

	validateSymbolFromToken(t, client, token, "aEthUSDT")
}

func TestAave_GenerateCalldata_Withdraw(t *testing.T) {
	// cast calldata "withdraw(address,uint256,address)" 0xc0ffee254729296a45a3885639AC7E10F9d54979 500000000000000000 0x0000000000000000000000000000000000000000
	// 0x69328dec000000000000000000000000c0ffee254729296a45a3885639ac7e10f9d5497900000000000000000000000000000000000000000000000006f05b59d3b200000000000000000000000000000000000000000000000000000000000000000000

	expectedCalldata := "0x69328dec000000000000000000000000c0ffee254729296a45a3885639ac7e10f9d5497900000000000000000000000000000000000000000000000006f05b59d3b200000000000000000000000000000000000000000000000000000000000000000000"

	t.Run("bsc chain for aave", func(t *testing.T) {

		aave, err := NewAaveOperation(getTestClient(t, ChainBSC), big.NewInt(56), AaveProtocolForkAave)
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
	})

	t.Run("ethereum chain for aave", func(t *testing.T) {

		aave, err := NewAaveOperation(getTestClient(t, ChainETH), big.NewInt(1), AaveProtocolForkAave)
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
	})

	t.Run("bsc chain for avalon finance", func(t *testing.T) {

		avalonFinance, err := NewAaveOperation(getTestClient(t, ChainBSC), big.NewInt(56), AaveProtocolForkAvalonFinance)
		require.NoError(t, err)

		calldata, err := avalonFinance.GenerateCalldata(context.Background(), big.NewInt(56), LoanWithdraw, TransactionParams{
			Amount: big.NewInt(500000000000000000),
			Sender: common.HexToAddress("0x0000000000000000000000000000000000000000"),
			Asset:  common.HexToAddress("0xc0ffee254729296a45a3885639AC7E10F9d54979"),
			ExtraData: map[string]interface{}{
				"referral_code": 0,
			},
		})

		require.NoError(t, err)
		require.Equal(t, expectedCalldata, calldata)
	})
}

func TestAave_GenerateCalldata_Supply(t *testing.T) {
	// cast calldata "supply(address,uint256,address,uint16)" 0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984 1000000000000000000 0x0000000000000000000000000000000000000000 10
	// 0x617ba0370000000000000000000000001f9840a85d5af5bf1d1762f925bdaddc4201f9840000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a

	expectedCalldata := "0x617ba0370000000000000000000000001f9840a85d5af5bf1d1762f925bdaddc4201f9840000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a"

	aave, err := NewAaveOperation(getTestClient(t, ChainETH), big.NewInt(1), AaveProtocolForkAave)
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
