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
		t.Log(err)
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

	t.Run("zero value supplied", func(t *testing.T) {

		err = listaStaking.Validate(context.Background(), big.NewInt(56), NativeStake, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  hotWallet,
			Sender: common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("unsupported action", func(t *testing.T) {

		err = listaStaking.Validate(context.Background(), big.NewInt(56), NativeUnStake, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  common.HexToAddress("0xae78736cd615f374d3085123a210448e74fc6393"),
			Sender: common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("user without balance balance cannot stake", func(t *testing.T) {

		err = listaStaking.Validate(context.Background(), big.NewInt(56), NativeStake, TransactionParams{
			Amount: big.NewInt(1),
			Sender: common.HexToAddress("0xFc21d6d146E6086B8359705C8b28512a983db0cb"),
		})

		require.Error(t, err)
	})

}

func TestListaStaking_GetBalance(t *testing.T) {

	listaStaking, err := NewListaStakingOperation(getTestClient(t, ChainBSC), big.NewInt(56))
	require.NoError(t, err)

	bal, err := listaStaking.GetBalance(context.Background(), big.NewInt(56), hotWallet,
		common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"))

	require.NoError(t, err)
	require.NotNil(t, bal)
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
