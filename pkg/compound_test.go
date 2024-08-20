package pkg

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestCompoundV3_New(t *testing.T) {

	t.Run("unsupported chain", func(t *testing.T) {
		compoundImpl, err := NewCompoundOperation(getTestClient(t), big.NewInt(100),
			common.HexToAddress("0xa17581a9e3356d9a858b789d68b4d866e593ae94"))

		require.Error(t, err)
		require.Nil(t, compoundImpl)
	})

	t.Run("unsupported pool market", func(t *testing.T) {

		compoundImpl, err := NewCompoundOperation(getTestClient(t), big.NewInt(1),
			common.HexToAddress(nativeDenomAddress))

		require.Error(t, err)
		require.Nil(t, compoundImpl)
	})
}

func TestCompound_GenerateCalldata_Supply(t *testing.T) {

	// cast calldata "supply(address,uint256)" 0x514910771AF9Ca656af840dff83E8264EcF986CA 1000000000000000000
	// 0xf2b9fdb8000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000
	expectedCalldata := "0xf2b9fdb8000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000"

	compoundClient, err := NewCompoundOperation(getTestClient(t), big.NewInt(1),
		common.HexToAddress("0xc3d688b66703497daa19211eedff47f25384cdc3"))
	require.NoError(t, err)

	calldata, err := compoundClient.GenerateCalldata(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
		Asset:  common.HexToAddress("0x514910771AF9Ca656af840dff83E8264EcF986CA"),
		Amount: big.NewInt(1e18),
	})

	require.NoError(t, err)
	require.Equal(t, expectedCalldata, calldata)
}

func TestCompound_GenerateCalldata_Withdraw(t *testing.T) {

	// cast calldata "withdraw(address,uint256)" 0x514910771AF9Ca656af840dff83E8264EcF986CA 1000000000000000000
	// 0xf3fef3a3000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a764000
	expectedCalldata := "0xf3fef3a3000000000000000000000000514910771af9ca656af840dff83e8264ecf986ca0000000000000000000000000000000000000000000000000de0b6b3a7640000"

	compoundClient, err := NewCompoundOperation(getTestClient(t), big.NewInt(1),
		common.HexToAddress("0xc3d688b66703497daa19211eedff47f25384cdc3"))
	require.NoError(t, err)

	calldata, err := compoundClient.GenerateCalldata(context.Background(), big.NewInt(1),
		LoanWithdraw, TransactionParams{
			Asset:  common.HexToAddress("0x514910771AF9Ca656af840dff83E8264EcF986CA"),
			Amount: big.NewInt(1 * 1e18),
		})
	require.NoError(t, err)
	require.Equal(t, expectedCalldata, calldata)
}

func TestCompound_IsSupportedAsset(t *testing.T) {

	compoundImpl, err := NewCompoundOperation(getTestClient(t), big.NewInt(1),
		common.HexToAddress("0xa17581a9e3356d9a858b789d68b4d866e593ae94"))

	require.NoError(t, err)

	t.Run("unsupported chain", func(t *testing.T) {
		supported := compoundImpl.IsSupportedAsset(context.TODO(), big.NewInt(100), common.HexToAddress(nativeDenomAddress))
		require.False(t, supported)
	})

	t.Run("native asset supported", func(t *testing.T) {
		supported := compoundImpl.IsSupportedAsset(context.TODO(), big.NewInt(1), common.HexToAddress(nativeDenomAddress))
		require.False(t, supported)
	})

	t.Run("rEth  asset supported", func(t *testing.T) {
		supported := compoundImpl.IsSupportedAsset(context.TODO(), big.NewInt(1), common.HexToAddress("0xae78736Cd615f374D3085123A210448E74Fc6393"))
		require.True(t, supported)
	})
}

func TestCompound_GetBalance(t *testing.T) {

	compoundImpl, err := NewCompoundOperation(getTestClient(t), big.NewInt(1),
		common.HexToAddress("0xa17581a9e3356d9a858b789d68b4d866e593ae94"))

	require.NoError(t, err)

	bal, err := compoundImpl.GetBalance(context.Background(), big.NewInt(1), emptyTestWallet,
		common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")) // empty address since the erc20 interface is used

	require.NoError(t, err)
	require.NotNil(t, bal)
}

func TestCompoundV3_Validate_ETH_Market(t *testing.T) {

	compoundImpl, err := NewCompoundOperation(getTestClient(t), big.NewInt(1),
		common.HexToAddress("0xa17581a9e3356d9a858b789d68b4d866e593ae94"))

	require.NoError(t, err)
	require.NotNil(t, compoundImpl)

	tt := []struct {
		name     string
		address  string
		hasError bool
	}{
		{
			name:     "wbtc cannot be supplied",
			address:  "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
			hasError: true,
		},
	}

	for _, v := range tt {
		err = compoundImpl.Validate(context.Background(), big.NewInt(1), LoanSupply, TransactionParams{
			Amount: big.NewInt(1 * 1e8),
			Asset:  common.HexToAddress(v.address),
			Sender: hotWallet,
		})

		if v.hasError {
			require.Error(t, err)
			continue
		}

		require.NoError(t, err)
	}
}
