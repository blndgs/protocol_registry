package pkg

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

var hotWallet = common.HexToAddress("0xee5b5b923ffce93a870b3104b7ca09c3db80047a") // bybit hot wallet

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
			Sender: common.HexToAddress("0xee5b5b923ffce93a870b3104b7ca09c3db80047a"), //BYBIT hot wallet
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
