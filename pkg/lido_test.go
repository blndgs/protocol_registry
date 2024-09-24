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

const abiString = `
  [{
    "constant": true,
    "inputs": [],
    "name": "symbol",
    "outputs": [
      {
        "name": "",
        "type": "string"
      }
    ],
    "payable": false,
    "stateMutability": "pure",
    "type": "function"
  }]
		`

// ENUM(ETH,BSC,POLYGON)
type Chain string

func getTestClient(t *testing.T, c Chain) *ethclient.Client {
	client, err := ethclient.Dial(getTestRPCURL(t, c))
	require.NoError(t, err)

	return client
}

func TestLido_Validate(t *testing.T) {

	t.Run("unsupported chain", func(t *testing.T) {

		lido, err := NewLidoOperation(getTestClient(t, ChainETH), big.NewInt(1))
		require.NoError(t, err)

		err = lido.Validate(context.Background(), big.NewInt(100), NativeStake, TransactionParams{
			Amount: big.NewInt(0),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("unsupported action", func(t *testing.T) {

		lido, err := NewLidoOperation(getTestClient(t, ChainETH), big.NewInt(1))
		require.NoError(t, err)

		err = lido.Validate(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
			Amount: big.NewInt(1),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("user without eth balance cannot stake", func(t *testing.T) {

		lido, err := NewLidoOperation(getTestClient(t, ChainETH), big.NewInt(1))
		require.NoError(t, err)

		err = lido.Validate(context.Background(), big.NewInt(1), NativeStake, TransactionParams{
			Amount: big.NewInt(1),
			Sender: common.HexToAddress("0x4Ce67641D9D25aC651ef3DaC64736f621A4d14D"),
			Asset:  common.HexToAddress(nativeDenomAddress),
		})

		require.Error(t, err)
	})

	t.Run("user with eth balance can stake", func(t *testing.T) {

		lido, err := NewLidoOperation(getTestClient(t, ChainETH), big.NewInt(1))
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

	client := getTestClient(t, ChainETH)

	lido, err := NewLidoOperation(client, big.NewInt(1))
	require.NoError(t, err)

	account := common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6")

	token, bal, err := lido.GetBalance(context.Background(), big.NewInt(1),
		account, common.HexToAddress(""))

	require.NoError(t, err)
	require.NotNil(t, bal)

	validateSymbolFromToken(t, client, token, "stETH")
}

func TestLido_GenerateCalldata(t *testing.T) {
	// cast calldata "submit(address)" 0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6
	// 0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6

	expectedCalldata := "0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6"

	lido, err := NewLidoOperation(getTestClient(t, ChainETH), big.NewInt(1))
	require.NoError(t, err)

	calldata, err := lido.GenerateCalldata(context.Background(), big.NewInt(1), NativeStake, TransactionParams{
		Recipient: common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
	})

	require.NoError(t, err)
	require.Equal(t, expectedCalldata, calldata)
}
