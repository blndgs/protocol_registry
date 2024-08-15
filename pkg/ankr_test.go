package pkg

// func TestAnkr_GenerateCalldata_Supply(t *testing.T) {
//
// 	// cast calldata "stakeAndClaimAethC()"
// 	// 0x9fa65c56
// 	expectedCalldata := "0x9fa65c56"
//
// 	ankr := &AnkrOperation{}
//
// 	err := ankr.Initialize(context.Background(), ProtocolConfig{
// 		RPCURL: getTestRPCURL(t),
// 	})
// 	require.NoError(t, err)
//
// 	calldata, err := ankr.GenerateCalldata(context.Background(), big.NewInt(1), NativeStake, TransactionParams{})
//
// 	require.NoError(t, err)
// 	require.Equal(t, expectedCalldata, calldata)
// }
//
// func TestAnkr_GenerateCalldata_Withdraw(t *testing.T) {
// 	// cast calldata "unstakeAETH(uint256)" 3987509938965136896
// 	// 0xc957619d00000000000000000000000000000000000000000000000037567b29aa5b4600
//
// 	expectedCalldata := "0xc957619d00000000000000000000000000000000000000000000000037567b29aa5b4600"
//
// 	ankr := &AnkrOperation{}
//
// 	err := ankr.Initialize(context.Background(), ProtocolConfig{
// 		RPCURL: getTestRPCURL(t),
// 	})
// 	require.NoError(t, err)
//
// 	calldata, err := ankr.GenerateCalldata(context.Background(), big.NewInt(1), NativeUnStake, TransactionParams{
// 		Amount: big.NewInt(3987509938965136896),
// 	})
//
// 	require.Equal(t, expectedCalldata, calldata)
// }
