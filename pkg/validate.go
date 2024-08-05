package pkg

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// IsNativeToken checks if the token is ETH
func IsNativeToken(asset common.Address) bool {
	return strings.ToLower(asset.Hex()) == nativeDenomAddress
}

// nativeDenomAddress native denom token address.
const nativeDenomAddress = "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"

var tokenSupportedMap = map[int64]map[ProtocolName][]string{
	1: {
		AaveV3: {
			"0x7f39c581f595b53c5cb19bd0b3f8da6c935e2ca0", // Wrapped Liquid Staked Ether
			"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599", // Wrapped BTC
			"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", // USDC
			"0xdac17f958d2ee523a2206206994597c13d831ec7", // Tether
			"0xCd5fE23C85820F7B72D0926FC9b05b43E359b7ee", // Wrapped eETH
			"0x514910771af9ca656af840dff83e8264ecf986ca", // ChainLink
			"0xae78736cd615f374d3085123a210448e74fc6393", // RocketPool ETH
			"0x6b175474e89094c44da98b954eedeac495271d0f", // DAI
			"0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9", // Aave
			"0x83f20f44975d03b1b09e64809b757c47f942beea", // savingsDAI
			"0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2", // Maker
			"0xBe9895146f7AF43049ca1c1AE358B0541Ea49704", // Coinbase Ether
			"0xf1C9acDc66974dFB6dEcB12aA385b9cD01190E38", // osETH
			"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", // WETH
			"0xae78736cd615f374d3085123a210448e74fc6393", // RocketPool ETH
		},
		SparkLend: {
			"0x83f20f44975d03b1b09e64809b757c47f942beea", // savingsDAI
			"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599", // Wrapped BTC
			"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", // USDC
			"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", // WETH
			"0x7f39c581f595b53c5cb19bd0b3f8da6c935e2ca0", // wsETH ( Lido )
			"0xae78736cd615f374d3085123a210448e74fc6393", // RocketPool ETH
			"0xdac17f958d2ee523a2206206994597c13d831ec7", // Tether
		},
		Lido: {},
	},
}

// chainID -> Contract address -> ERC20s that can be used as collateral
// Compound has different markets and each market only supports a
// few assets as collateral
var compoundSupportedAssets = map[int64]map[string][]string{
	// Ethereum
	1: {
		// USDC pool
		"0xc3d688b66703497daa19211eedff47f25384cdc3": []string{
			nativeDenomAddress,                           // ETH
			"0x514910771AF9Ca656af840dff83E8264EcF986CA", // LINK
			"0xc00e94Cb662C3520282E6f5717214004A7f26888", // COMP
			"0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984", // UNI
			"0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", // WBTC
		},
		// ETH pool
		"0xa17581a9e3356d9a858b789d68b4d866e593ae94": []string{
			"0xBe9895146f7AF43049ca1c1AE358B0541Ea49704", // cbETH
			"0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0", // wsETH (Lido)
			"0xae78736Cd615f374D3085123A210448E74Fc6393", // RocketPool ETH
			"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", // WETH
		},
	},
}
