package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/blndgs/protocol_registry/pkg"
	"github.com/blndgs/protocol_registry/tokens"
)

func main() {
	chains := []struct {
		Name    string
		ChainID *big.Int
	}{
		{
			Name:    "Ethereum",
			ChainID: big.NewInt(1),
		},
		{
			Name:    "BNB Chain",
			ChainID: big.NewInt(56),
		},
		{
			Name:    "Polygon",
			ChainID: big.NewInt(137),
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

		// Read existing token data
		fileName := fmt.Sprintf("tokens/%d.json", chain.ChainID)
		existingData, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Printf("error reading existing data for %s: %v\n", chain.Name, err)
			continue
		}

		var protocolData tokens.Data
		err = json.Unmarshal(existingData, &protocolData)
		if err != nil {
			fmt.Printf("error unmarshaling existing data for %s: %v\n", chain.Name, err)
			continue
		}

		// Clear existing protocols but keep tokens
		existingTokens := protocolData.Tokens
		protocolData.Protocols = nil

		protocols := registry.ListProtocols(chain.ChainID)

		// Convert protocols to the format expected by tokens.Data
		for _, p := range protocols {
			supportedAssets, err := p.GetSupportedAssets(ctx, chain.ChainID)
			if err != nil {
				fmt.Printf("error getting supported assets for protocol %s on chain %s: %v\n", p.GetName(), chain.Name, err)
				continue
			}

			// Convert addresses to strings
			tokenAddrs := make([]string, len(supportedAssets))
			for i, addr := range supportedAssets {
				tokenAddrs[i] = addr.Hex()
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

		// Restore existing tokens
		protocolData.Tokens = existingTokens

		// Write the data to the corresponding chain file
		data, err := json.MarshalIndent(protocolData, "", "  ")
		if err != nil {
			fmt.Printf("error marshaling data for %s: %v\n", chain.Name, err)
			continue
		}

		err = os.WriteFile(fileName, data, 0644)
		if err != nil {
			fmt.Printf("error writing file for %s: %v\n", chain.Name, err)
			continue
		}
	}
}
