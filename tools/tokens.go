package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/blndgs/protocol_registry/pkg"
	"github.com/blndgs/protocol_registry/tokens"
	"github.com/joho/godotenv"
)

const (
	nativeTokenAddress = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"
)

type MoralisResponse struct {
	TokenName     string `json:"tokenName"`
	TokenSymbol   string `json:"tokenSymbol"`
	TokenDecimals string `json:"tokenDecimals"`
}

func getTokenMetadata(ctx context.Context, tokenAddr string, chainID *big.Int) (string, string, int, error) {

	var moralisAPIKey = os.Getenv("MORALIS_API_KEY")
	if moralisAPIKey == "" {
		panic("provide moralis api key in .env")
	}

	var chainStr string
	switch chainID.Int64() {
	case 1:
		chainStr = "eth"
	case 56:
		chainStr = "bsc"
	case 137:
		chainStr = "polygon"
	default:
		return "", "", 0, fmt.Errorf("unsupported chain ID: %d", chainID)
	}

	url := fmt.Sprintf("https://deep-index.moralis.io/api/v2.2/erc20/%s/price?chain=%s&include=percent_change", tokenAddr, chainStr)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-API-Key", moralisAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check if response status is greater than 201
	if resp.StatusCode > 201 {
		return "", "", 0, fmt.Errorf("invalid token or API error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to read response body: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to parse response: %v", err)
	}

	name, ok := result["tokenName"].(string)
	if !ok {
		return "", "", 0, fmt.Errorf("token name not found")
	}

	symbol, ok := result["tokenSymbol"].(string)
	if !ok {
		return "", "", 0, fmt.Errorf("token symbol not found")
	}

	decimalsStr, ok := result["tokenDecimals"].(string)
	if !ok {
		return "", "", 0, fmt.Errorf("token decimals not found")
	}

	var decimals int
	_, err = fmt.Sscanf(decimalsStr, "%d", &decimals)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to parse decimals: %v", err)
	}

	return name, symbol, decimals, nil
}

// sortProtocolData sorts both tokens and protocols arrays in the protocol data
func sortProtocolData(data *tokens.Data) {
	// Sort tokens by symbol
	sort.Slice(data.Tokens, func(i, j int) bool {
		return strings.ToLower(data.Tokens[i].Symbol) < strings.ToLower(data.Tokens[j].Symbol)
	})

	// Sort protocols by name
	sort.Slice(data.Protocols, func(i, j int) bool {
		return strings.ToLower(data.Protocols[i].Name) < strings.ToLower(data.Protocols[j].Name)
	})

	// Deduplicate and sort token arrays within each protocol
	for i := range data.Protocols {
		// Use a map to deduplicate tokens
		tokenMap := make(map[string]bool)
		var uniqueTokens []string

		for _, token := range data.Protocols[i].Tokens {
			if !tokenMap[token] {
				tokenMap[token] = true
				uniqueTokens = append(uniqueTokens, token)
			}
		}

		// Sort the unique tokens
		sort.Strings(uniqueTokens)

		// Replace the original tokens array with the sorted, unique tokens
		data.Protocols[i].Tokens = uniqueTokens
	}
}

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Println("add an .env file with MORALIS_API_KEY=xyz")
		panic(err)
	}

	chains := []struct {
		Name         string
		ChainID      *big.Int
		MoralisChain string
		NativeSymbol string
		NativeName   string
	}{
		{
			Name:         "Ethereum",
			ChainID:      big.NewInt(1),
			MoralisChain: "eth",
			NativeSymbol: "ETH",
			NativeName:   "Ethereum",
		},
		{
			Name:         "BNB Chain",
			ChainID:      big.NewInt(56),
			MoralisChain: "bsc",
			NativeSymbol: "BNB",
			NativeName:   "Binance Coin",
		},
		{
			Name:         "Polygon",
			ChainID:      big.NewInt(137),
			MoralisChain: "polygon",
			NativeSymbol: "MATIC",
			NativeName:   "Polygon",
		},
	}

	registry, err := pkg.NewProtocolRegistry([]pkg.ChainConfig{
		{
			ChainID: big.NewInt(1),
			RPCURL:  "https://rpc.ankr.com/eth",
		},
		{
			ChainID: big.NewInt(56),
			RPCURL:  "https://rpc.ankr.com/bsc",
		},
		{
			ChainID: big.NewInt(137),
			RPCURL:  "https://rpc.ankr.com/polygon",
		},
	})

	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()

	for _, chain := range chains {
		fmt.Println("updating file for", chain.Name)

		var protocolData tokens.Data
		uniqueTokens := make(map[string]bool)
		validTokens := make(map[string]bool)

		protocols := registry.ListProtocols(chain.ChainID)

		// First, collect all unique tokens and validate them
		for _, p := range protocols {
			supportedAssets, err := p.GetSupportedAssets(ctx, chain.ChainID)
			if err != nil {
				panic(fmt.Sprintf("error getting supported assets for protocol %s on chain %s: %v", p.GetName(), chain.Name, err))
			}

			for _, addr := range supportedAssets {
				addrStr := addr.Hex()
				if !uniqueTokens[addrStr] {
					uniqueTokens[addrStr] = true

					// Skip validation for native token
					if strings.EqualFold(addrStr, nativeTokenAddress) {
						validTokens[addrStr] = true
						continue
					}

					// Validate token using Moralis
					_, _, _, err := getTokenMetadata(ctx, addrStr, chain.ChainID)
					if err == nil {
						validTokens[addrStr] = true
					} else {
						fmt.Printf("Skipping invalid token %s on chain %s: %v\n", addrStr, chain.Name, err)
					}
				}
			}
		}

		// Now add protocols, but only include valid tokens
		for _, p := range protocols {
			supportedAssets, _ := p.GetSupportedAssets(ctx, chain.ChainID)
			var validTokenAddrs []string

			for _, addr := range supportedAssets {
				addrStr := addr.Hex()
				if validTokens[addrStr] {
					validTokenAddrs = append(validTokenAddrs, addrStr)
				}
			}

			// Only add protocol if it has valid tokens
			if len(validTokenAddrs) > 0 {
				protocolData.Protocols = append(protocolData.Protocols, tokens.Protocol{
					Address:     p.GetContractAddress(chain.ChainID).Hex(),
					Name:        p.GetName(),
					Type:        string(p.GetType()),
					Source:      true,
					Destination: true,
					Tokens:      validTokenAddrs,
				})
			}
		}

		// Add valid tokens to the tokens array
		for addr := range validTokens {
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
				name, symbol, decimals, err := getTokenMetadata(ctx, addr, chain.ChainID)
				if err != nil {
					// This shouldn't happen as we already validated the token
					continue
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

		// Sort the protocol data
		sortProtocolData(&protocolData)

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
