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

func TestListaStaking_New(t *testing.T) {

	t.Run("unsupported chain", func(t *testing.T) {
		_, err := NewListaStakingOperation(getTestClient(t, ChainETH), big.NewInt(100))
		require.Error(t, err)
		require.Equal(t, err, ErrChainUnsupported)
	})

	t.Run("only bnb supported", func(t *testing.T) {
		_, err := NewListaStakingOperation(getTestClient(t, ChainBSC), big.NewInt(56))
		require.NoError(t, err)
	})

	t.Run("network id of bsc network client does not match eth chain", func(t *testing.T) {
		_, err := NewListaStakingOperation(getTestClient(t, ChainETH), big.NewInt(56))
		require.Error(t, err)
		require.Contains(t, err.Error(), "network id does not match")
	})
}

func TestListaStaking_Validate(t *testing.T) {

	listaStaking, err := NewListaStakingOperation(getTestClient(t, ChainBSC), big.NewInt(56))
	require.NoError(t, err)

	t.Run("unsupported action", func(t *testing.T) {

		err = listaStaking.Validate(context.Background(), big.NewInt(56), NativeUnStake, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393"),
			Sender: common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	// This is because the solver can have multiple dynamic steps
	// User can have USDT but wants to stake in Lista.
	// Solver goes from USDT-> BNB -> Lista
	// Being strict here fails the validation and halts the solver
	t.Run("user without balance can stake", func(t *testing.T) {

		err = listaStaking.Validate(context.Background(), big.NewInt(56), NativeStake, TransactionParams{
			Amount: big.NewInt(1),
			Sender: common.HexToAddress("0xFc21d6d146E6086B8359705C8b28512a983db0cb"),
		})

		require.NoError(t, err)
	})

	t.Run("user with balance can stake", func(t *testing.T) {

		err = listaStaking.Validate(context.Background(), big.NewInt(56), NativeStake, TransactionParams{
			Amount: big.NewInt(1),
			// Binance hot wallet would always have BNB
			Sender: common.HexToAddress("0x8894E0a0c962CB723c1976a4421c95949bE2D4E3"),
		})

		require.NoError(t, err)
	})
}

func TestListaStaking_GetBalance(t *testing.T) {

	client := getTestClient(t, ChainBSC)

	listaStaking, err := NewListaStakingOperation(client, big.NewInt(56))
	require.NoError(t, err)

	wallet := common.HexToAddress("0x6F28FeC449dbd2056b76ac666350Af8773E03873")

	token, bal, err := listaStaking.GetBalance(context.Background(),
		big.NewInt(56), wallet, common.HexToAddress(""))

	require.NoError(t, err)
	require.NotNil(t, bal)

	validateSymbolFromToken(t, client, token, "slisBNB")
}

func TestListaStaking_GenerateCalldata_Supply(t *testing.T) {
	// cast calldata "deposit()"
	// 0xd0e30db0

	expectedCalldata := "0xd0e30db0"

	staking, err := NewListaStakingOperation(getTestClient(t, ChainBSC), big.NewInt(56))
	require.NoError(t, err)

	calldata, err := staking.GenerateCalldata(context.Background(), big.NewInt(56), NativeStake, TransactionParams{
		Asset:  common.HexToAddress("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984"),
		Amount: big.NewInt(1000000000000000000),
		Sender: common.HexToAddress("0x0000000000000000000000000000000000000000"),
	})
	require.NoError(t, err)

	require.Equal(t, expectedCalldata, calldata)
}
