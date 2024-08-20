package pkg

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func getTestClient(t *testing.T) *ethclient.Client {
	client, err := ethclient.Dial(getTestRPCURL(t))
	require.NoError(t, err)

	return client
}

func TestLido_Validate(t *testing.T) {

	t.Run("zero value staked", func(t *testing.T) {

		lido, err := NewLidoOperation(getTestClient(t), big.NewInt(1))
		require.NoError(t, err)

		err = lido.Validate(context.Background(), big.NewInt(1), NativeStake, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("unsupported chain", func(t *testing.T) {

		lido, err := NewLidoOperation(getTestClient(t), big.NewInt(1))
		require.NoError(t, err)

		err = lido.Validate(context.Background(), big.NewInt(100), NativeStake, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("unsupported action", func(t *testing.T) {

		lido, err := NewLidoOperation(getTestClient(t), big.NewInt(1))
		require.NoError(t, err)

		err = lido.Validate(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
			Amount: big.NewInt(1),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("user without eth balance cannot stake", func(t *testing.T) {

		lido, err := NewLidoOperation(getTestClient(t), big.NewInt(1))
		require.NoError(t, err)

		err = lido.Validate(context.Background(), big.NewInt(1), NativeStake, TransactionParams{
			Amount: big.NewInt(1),
			Sender: common.HexToAddress("0x4Ce67641D9D25aC651ef3DaC64736f621A4d14D"),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("user with eth balance can stake", func(t *testing.T) {

		lido, err := NewLidoOperation(getTestClient(t), big.NewInt(1))
		require.NoError(t, err)

		err = lido.Validate(context.Background(), big.NewInt(1), NativeStake, TransactionParams{
			Amount: big.NewInt(1),
			Sender: common.HexToAddress("0xee5b5b923ffce93a870b3104b7ca09c3db80047a"), //BYBIT hot wallet
			Asset:  common.HexToAddress(nativeDenomAddress),
		})

		require.NoError(t, err)
	})
}

func TestLido_GetBalance(t *testing.T) {

	lido, err := NewLidoOperation(getTestClient(t), big.NewInt(1))
	require.NoError(t, err)

	account := common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6")

	bal, err := lido.GetBalance(context.Background(), big.NewInt(1), account, common.Address{})

	require.NoError(t, err)
	require.NotNil(t, bal)
}

func TestLido_GenerateCalldata(t *testing.T) {
	// cast calldata "submit(address)" 0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6
	// 0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6

	expectedCalldata := "0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6"

	lido, err := NewLidoOperation(getTestClient(t), big.NewInt(1))
	require.NoError(t, err)

	calldata, err := lido.GenerateCalldata(context.Background(), big.NewInt(1), NativeStake, TransactionParams{
		Recipient: common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
	})

	require.NoError(t, err)
	require.Equal(t, expectedCalldata, calldata)
}
