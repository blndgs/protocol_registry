# Token Registry Specifications

The `Token Registry` is provides functionality for managing token and protocol data across different blockchain networks.
It offers a flexible and thread-safe way to retrieve information about tokens and protocols based on their chain ID and address.

```go
// Token represents a token with its properties
type Token struct {
 TokenAddress string `json:"token_address"`
 Name         string `json:"name"`
 Symbol       string `json:"symbol"`
 Decimals     int    `json:"decimals"`
}

// Protocol represents a protocol with its properties
type Protocol struct {
 Address     string   `json:"address"`
 Name        string   `json:"name"`
 Type        string   `json:"type`"
 Source      bool     `json:"source"`
 Destination bool     `json:"destination"`
 Tokens      []string `json:"tokens"`
}

// Data represents the entire structure of the JSON file
type Data struct {
 Tokens    []Token    `json:"tokens"`
 Protocols []Protocol `json:"protocols"`
}

// TokenRegistry is an interface for retrieving token and protocol data
type TokenRegistry interface {
 // GetTokens retrieves all tokens for a given chain ID.
 GetTokens(chainID *big.Int) ([]Token, error)

 // GetProtocols retrieves all protocols for a given chain ID.
 GetProtocols(chainID *big.Int) ([]Protocol, error)

 // GetTokenByAddress retrieves a specific token by its address for a given chain ID.
 GetTokenByAddress(chainID *big.Int, address string) (*Token, error)

 // GetProtocolByAddress retrieves a specific protocol by its address for a given chain ID.
 GetProtocolByAddress(chainID *big.Int, address string) (*Protocol, error)
}

// JSONTokenRegistry implements TokenRegistry for JSON files
type JSONTokenRegistry struct {
 data     map[string]*Data
 dataLock sync.RWMutex
}
```

## Implementation Considerations

1. Flexibility: By providing an `Initialize` method, each protocol can perform necessary setup actions, which might include establishing network connections, fetching contract addresses, or loading ABIs.

2. Dynamic Interaction: The `GetABI` method allows client code to interact dynamically with the protocol’s contract without hardcoding function calls, making the interface adaptable to future changes in the contract.

3. Action and Validation: `GenerateCalldata` and `Validate` take an action and parameters, offering the flexibility to support a wide range of operations without changing the interface. This approach accommodates custom actions that may be specific to certain.

4. Supported Assets: By including a method `GetSupportedAssets` to fetch supported assets per chain, and balance function `GetBalance` to provide the balance of an address per protocol can provide necessary information for front-end applications or other services that need to display or utilize asset data dynamically.

## Usage

To use the Token Registry:

1. Initialize a `JSONTokenRegistry` instance.
2. Use the methods defined in the `TokenRegistry` interface to retrieve token and protocol data.
3. Always provide the appropriate chain ID when calling methods to ensure you're working with the correct network data.

## Thread Safety

The `JSONTokenRegistry` implementation is thread-safe. It uses a sync.RWMutex to protect concurrent access to the data map. This allows for safe use in multi-goroutine environments.

## Extensibility

The `TokenRegistry interface` allows for easy extension of the package. New implementations can be created for different data sources while maintaining the same method signatures, ensuring compatibility with existing code.
