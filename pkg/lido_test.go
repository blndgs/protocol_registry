package pkg

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestLido_GenerateCalldata(t *testing.T) {
	// cast calldata "submit(address)" 0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6
	// 0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6

	expectedCalldata := "0xa1903eab000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6"

	lido, err := NewLidoOperation()
	require.NoError(t, err)

	calldata, err := lido.GenerateCalldata(NativeStake, GenerateCalldataOptions{
		Sender: common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
	})
	require.NoError(t, err)

	require.Equal(t, expectedCalldata, calldata)
}

func TestLido_GenerateCalldataUnspportedAction(t *testing.T) {

	lido, err := NewLidoOperation()
	require.NoError(t, err)

	_, err = lido.GenerateCalldata(LoanSupply, GenerateCalldataOptions{
		Sender: common.HexToAddress("0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"),
	})
	require.Error(t, err)
}
