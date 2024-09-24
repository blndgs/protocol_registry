package pkg

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

const zeroAddress = "0x0000000000000000000000000000000000000000"

// nativeDenomAddress native denom token address.
const nativeDenomAddress = "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"

var tokenSupportedMap = map[int64]map[ProtocolName][]string{
	EthChainID.Int64(): {
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
	BscChainID.Int64(): {
		AaveV3: {
			"0x2170ed0880ac9a755fd29b2688956bd959f933f8", // ETH ( Binance pegged ETH )
			"0x7130d2a12b9bcbfae4f2634d864a1ee1ce3ead9c", // BTC ( Binance pegged BTC )
			"0x55d398326f99059fF775485246999027B3197955", // USDT ( Binance pegged USDT )
			"0xc5f0f7b66764F6ec8C8Dff7BA683102295E16409", // FDSD
			"0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d", // USDC ( Binance pegged USDC )
			"0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c", // WBNB
		},
		AvalonFinance: {
			"0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d", // USDC
			"0x55d398326f99059ff775485246999027b3197955", // USDT
			"0x2170ed0880ac9a755fd29b2688956bd959f933f8", // ETH
			"0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c", // WBNB
			"0x7130d2A12B9BCbFAe4f2634d864A1Ee1Ce3Ead9c", // BTCB
			"0x53E63a31fD1077f949204b94F431bCaB98F72BCE", // SolvBTC.ena
			"0x4aae823a6a0b376De6A78e74eCC5b079d38cBCf7", // SolvBTC
			"0x1346b618dC92810EC74163e4c27004c921D446a5", // SolvBTC.BBN
		},
	},
	PolygonChainID.Int64(): {
		AaveV3: {
			"0x8f3Cf7ad23Cd3CaDbD9735AFf958023239c6A063",
			"0x53E0bca35eC356BD5ddDFebbD1Fc0fD03FaBad39",
			"0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174",
			"0x1BFD67037B42Cf73acF2047067bd4F2C47D9BfD6",
			"0x7ceB23fD6bC0adD59E62ac25578270cFf1b9f619",
			"0xc2132D05D31c914a87C6611C10748AEb04B58e8F",
			"0xD6DF932A45C0f255f85145f286eA0b292B21C90B",
			"0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270",
			"0x172370d5Cd63279eFa6d502DAB29171933a610AF",
			"0x0b3F868E0BE5597D5DB7fEB59E1CADBb0fdDa50a",
			"0x385Eeac5cB85A38A9a07A70c73e0a3271CfB54A7",
			"0x03b54A6e9a984069379fae1a4fC4dBAE93B3bCCD",
			"0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359",
		},
	},
}

// IsNativeToken checks if the token is ETH
func IsNativeToken(asset common.Address) bool {
	return strings.ToLower(asset.Hex()) == nativeDenomAddress
}
