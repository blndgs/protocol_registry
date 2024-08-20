package pkg

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

var emptyTestWallet = common.HexToAddress("0x6a22640F02F8c8b576a3193674c4aE97e0f8d007")

func TestAnkr_GenerateCalldata_Supply(t *testing.T) {

	// cast calldata "stakeAndClaimAethC()"
	// 0x9fa65c56
	expectedCalldata := "0x9fa65c56"

	ankr, err := NewAnkrOperation(getTestClient(t), big.NewInt(1))
	require.NoError(t, err)

	calldata, err := ankr.GenerateCalldata(context.Background(), big.NewInt(1), NativeStake, TransactionParams{})

	require.NoError(t, err)
	require.Equal(t, expectedCalldata, calldata)
}

func TestAnkr_GenerateCalldata_Withdraw(t *testing.T) {
	// cast calldata "unstakeAETH(uint256)" 3987509938965136896
	// 0xc957619d00000000000000000000000000000000000000000000000037567b29aa5b4600

	expectedCalldata := "0xc957619d00000000000000000000000000000000000000000000000037567b29aa5b4600"

	ankr, err := NewAnkrOperation(getTestClient(t), big.NewInt(1))
	require.NoError(t, err)

	calldata, err := ankr.GenerateCalldata(context.Background(), big.NewInt(1), NativeUnStake, TransactionParams{
		Amount: big.NewInt(3987509938965136896),
		Sender: emptyTestWallet,
	})

	require.NoError(t, err)
	require.Equal(t, expectedCalldata, calldata)
}

func TestAnkr_Validate(t *testing.T) {

	ankr, err := NewAnkrOperation(getTestClient(t), big.NewInt(1))
	require.NoError(t, err)

	t.Run("zero value staked", func(t *testing.T) {

		err = ankr.Validate(context.Background(), big.NewInt(1), NativeStake, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  common.HexToAddress(nativeDenomAddress),
			Sender: emptyTestWallet,
		})

		require.Error(t, err)
	})

	t.Run("unsupported chain", func(t *testing.T) {

		err = ankr.Validate(context.Background(), big.NewInt(100), NativeStake, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("unsupported action", func(t *testing.T) {

		err = ankr.Validate(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
			Amount: big.NewInt(1),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("staking more than available balance", func(t *testing.T) {

		err = ankr.Validate(context.Background(), big.NewInt(1), NativeStake, TransactionParams{
			Amount: big.NewInt(1 * 1e18),
			Asset:  common.HexToAddress(nativeDenomAddress),
			Sender: emptyTestWallet,
		})

		require.Error(t, err)
	})

	t.Run("user with eth balance can stake", func(t *testing.T) {

		err = ankr.Validate(context.Background(), big.NewInt(1), NativeStake, TransactionParams{
			Amount: big.NewInt(1 * 1e18),
			Sender: hotWallet,
			Asset:  common.HexToAddress(nativeDenomAddress),
		})

		require.NoError(t, err)
	})
}

func TestAnkr_IsSupportedAsset(t *testing.T) {

	ankr, err := NewAnkrOperation(getTestClient(t), big.NewInt(1))
	require.NoError(t, err)

	t.Run("unsupported chain", func(t *testing.T) {
		supported := ankr.IsSupportedAsset(context.TODO(), big.NewInt(100), common.HexToAddress(nativeDenomAddress))
		require.False(t, supported)
	})

	t.Run("native asset supported", func(t *testing.T) {
		supported := ankr.IsSupportedAsset(context.TODO(), big.NewInt(1), common.HexToAddress(nativeDenomAddress))
		require.True(t, supported)
	})

	t.Run("ankrEth asset supported", func(t *testing.T) {
		supported := ankr.IsSupportedAsset(context.TODO(), big.NewInt(1), ankrEthER20Account)
		require.True(t, supported)
	})
}

func TestAnkr_GetBalance(t *testing.T) {

	ankr, err := NewAnkrOperation(getTestClient(t), big.NewInt(1))
	require.NoError(t, err)

	bal, err := ankr.GetBalance(context.Background(), big.NewInt(1), emptyTestWallet,
		common.HexToAddress(nativeDenomAddress))

	require.NoError(t, err)
	require.NotNil(t, bal)

	bal, err = ankr.GetBalance(context.Background(), big.NewInt(1), emptyTestWallet, ankrEthER20Account)

	require.NoError(t, err)
	require.NotNil(t, bal)
}
