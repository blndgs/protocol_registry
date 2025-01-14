package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/blndgs/protocol_registry/pkg"
	"github.com/blndgs/protocol_registry/tokens"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	erc20ABI = `[
	{
		"constant": true,
		"inputs": [],
		"name": "name",
		"outputs": [{"name": "", "type": "string"}],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "symbol",
		"outputs": [{"name": "", "type": "string"}],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "decimals",
		"outputs": [{"name": "", "type": "uint8"}],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	}
]`
	nativeTokenAddress = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"
	mkrTokenAddress    = "0x9f8F72aA9304c8B593d555F12eF6589cC3A579A2"
)

func getTokenMetadata(ctx context.Context, client *ethclient.Client, tokenAddr common.Address, chainID *big.Int) (string, string, int, error) {
	// Special handling for MKR token on Ethereum
	if chainID.Cmp(big.NewInt(1)) == 0 && strings.EqualFold(tokenAddr.Hex(), mkrTokenAddress) {
		return "Maker", "MKR", 18, nil
	}

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// Get name
	nameData, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddr,
		Data: parsedABI.Methods["name"].ID,
	}, nil)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get name: %v", err)
	}
	var name string
	err = parsedABI.UnpackIntoInterface(&name, "name", nameData)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to unpack name: %v", err)
	}

	// Get symbol
	symbolData, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddr,
		Data: parsedABI.Methods["symbol"].ID,
	}, nil)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get symbol: %v", err)
	}
	var symbol string
	err = parsedABI.UnpackIntoInterface(&symbol, "symbol", symbolData)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to unpack symbol: %v", err)
	}

	// Get decimals
	decimalsData, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddr,
		Data: parsedABI.Methods["decimals"].ID,
	}, nil)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get decimals: %v", err)
	}
	var decimals uint8
	err = parsedABI.UnpackIntoInterface(&decimals, "decimals", decimalsData)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to unpack decimals: %v", err)
	}

	return name, symbol, int(decimals), nil
}

func main() {
	chains := []struct {
		Name         string
		ChainID      *big.Int
		RPCURL       string
		NativeSymbol string
		NativeName   string
	}{
		{
			Name:         "Ethereum",
			ChainID:      big.NewInt(1),
			RPCURL:       "https://rpc.ankr.com/eth",
			NativeSymbol: "ETH",
			NativeName:   "Ethereum",
		},
		{
			Name:         "BNB Chain",
			ChainID:      big.NewInt(56),
			RPCURL:       "https://rpc.ankr.com/bsc",
			NativeSymbol: "BNB",
			NativeName:   "Binance Coin",
		},
		{
			Name:         "Polygon",
			ChainID:      big.NewInt(137),
			RPCURL:       "https://rpc.ankr.com/polygon",
			NativeSymbol: "MATIC",
			NativeName:   "Polygon",
		},
	}

	registry, err := pkg.NewProtocolRegistry([]pkg.ChainConfig{
		{
			ChainID: big.NewInt(1),
			RPCURL:  chains[0].RPCURL,
		},
		{
			ChainID: big.NewInt(56),
			RPCURL:  chains[1].RPCURL,
		},
		{
			ChainID: big.NewInt(137),
			RPCURL:  chains[2].RPCURL,
		},
	})

	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()

	for _, chain := range chains {
		fmt.Println("updating file for", chain.Name)

		client, err := ethclient.Dial(chain.RPCURL)
		if err != nil {
			panic(fmt.Sprintf("error connecting to %s: %v", chain.Name, err))
		}
		defer client.Close()

		var protocolData tokens.Data
		uniqueTokens := make(map[string]bool)

		protocols := registry.ListProtocols(chain.ChainID)

		// Convert protocols to the format expected by tokens.Data
		for _, p := range protocols {
			supportedAssets, err := p.GetSupportedAssets(ctx, chain.ChainID)
			if err != nil {
				panic(fmt.Sprintf("error getting supported assets for protocol %s on chain %s: %v", p.GetName(), chain.Name, err))
			}

			// Convert addresses to strings and collect unique tokens
			tokenAddrs := make([]string, len(supportedAssets))
			for i, addr := range supportedAssets {
				addrStr := addr.Hex()
				tokenAddrs[i] = addrStr
				uniqueTokens[addrStr] = true
			}

			protocolData.Protocols = append(protocolData.Protocols, tokens.Protocol{
				Address:     p.GetContractAddress(chain.ChainID).Hex(),
				Name:        p.GetName(),
				Type:        string(p.GetType()),
				Source:      true,
				Destination: true,
				Tokens:      tokenAddrs,
			})
		}

		// Convert unique tokens to array and fetch metadata
		for addr := range uniqueTokens {
			var token tokens.Token

			if strings.EqualFold(addr, nativeTokenAddress) {
				// Handle native token
				token = tokens.Token{
					TokenAddress: addr,
					Name:         chain.NativeName,
					Symbol:       chain.NativeSymbol,
					Decimals:     18, // Native tokens always have 18 decimals
				}
			} else {
				// Handle ERC20 token
				tokenAddr := common.HexToAddress(addr)
				name, symbol, decimals, err := getTokenMetadata(ctx, client, tokenAddr, chain.ChainID)
				if err != nil {
					panic(fmt.Sprintf("error getting metadata for token %s on chain %s: %v", addr, chain.Name, err))
				}

				token = tokens.Token{
					TokenAddress: addr,
					Name:         name,
					Symbol:       symbol,
					Decimals:     decimals,
				}
			}

			protocolData.Tokens = append(protocolData.Tokens, token)
		}

		// Write the data to the corresponding chain file
		fileName := fmt.Sprintf("tokens/%d.json", chain.ChainID)
		data, err := json.MarshalIndent(protocolData, "", "  ")
		if err != nil {
			panic(fmt.Sprintf("error marshaling data for %s: %v", chain.Name, err))
		}

		err = os.WriteFile(fileName, data, 0644)
		if err != nil {
			panic(fmt.Sprintf("error writing file for %s: %v", chain.Name, err))
		}
	}
}
