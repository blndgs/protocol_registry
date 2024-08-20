# Whitelisted token

This document outlines the standard for maintaining whitelisted token lists across different blockchain networks.
The goal is to create a unified, easily accessible list of supported assets for all clients and services within our ecosystem.

This repository serves as the single source of truth for all whitelisted tokens across supported chains.

## File Format

Each supported blockchain network has its own JSON file named after its chain ID:

```sh
[chain_id].json
```

For example:

* `1.json` for Ethereum Mainnet
* `56.json` for Binance Smart Chain
* `137.json` for Polygon

## JSON Structure

The JSON file follows this structure:

```json
{
  "tokens": [
    {
      "token_address": "0xdcee70654261af21c44c093c300ed3bb97b78192"
    },
    {
      "token_address": "0xd2af830e8cbdfed6cc11bab697bb25496ed6fa62"
    }
  ]
}
```

## Token Selection Criteria

The whitelisted tokens in each list are chosen based on the following criteria:

1. Common assets supported by multiple providers (e.g., 0x, 1inch, native, enso).
2. Tokens supported by the frontend application.
3. Assets with reliable price data available through our price API.

## Importance of Standardization

Maintaining a standardized, common list of whitelisted tokens is crucial for several reasons:

1. `Consistency`: Ensures all clients and services work with the same set of supported assets.
2. `Security`: Helps prevent interactions with unauthorized or potentially malicious tokens.
3. `Efficiency`: Simplifies asset management and reduces the need for redundant whitelists.
4. `Scalability`: Makes it easier to add support for new chains or tokens across the entire ecosystem.

## Usage Guidelines

1. All clients, services, and applications within our ecosystem should reference the token lists from the official repository.
2. Regular updates to the lists should be made through pull requests to the repository.
3. Clients should implement a mechanism to fetch and cache the latest token lists periodically.
4. When interacting with tokens, always check against the whitelisted token list for the respective chain.
