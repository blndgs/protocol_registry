package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/blndgs/protocol_registry/pkg"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	// Define command-line flags
	protocolFlag := flag.String("protocol", "", "Protocol name (e.g., AaveV3, SparkLend)")
	actionFlag := flag.String("action", "", "Action name (e.g., supply, withdraw)")
	chainIDFlag := flag.Int64("chainID", 1, "Chain ID (default: 1)")
	helpFlag := flag.Bool("help", false, "Display help information")

	// Parse command-line flags
	flag.Parse()

	// Display help information if requested
	if *helpFlag {
		displayHelp()
		os.Exit(0)
	}

	// Validate required flags
	if *protocolFlag == "" {
		fmt.Println("Error: Protocol name is required")
		os.Exit(1)
	}
	if *actionFlag == "" {
		fmt.Println("Error: Action name is required")
		os.Exit(1)
	}

	// Create a new protocol registry and set up the operations
	registry := pkg.NewProtocolRegistry()
	pkg.SetupProtocolOperations(registry)

	// Convert action flag to ContractAction type
	action := pkg.ContractAction(*actionFlag)

	// Retrieve the operation from the registry
	operation, err := registry.GetProtocolOperation(*protocolFlag, action, big.NewInt(*chainIDFlag))
	if err != nil {
		fmt.Printf("Error retrieving operation: %v\n", err)
		os.Exit(1)
	}

	// Prompt the user to enter the args for the operation
	fmt.Println("Enter the args for the operation (comma-separated):")
	var argsInput string
	fmt.Scanln(&argsInput)

	// Parse and validate the args input based on the protocol and action
	args, err := parseAndValidateArgs(*protocolFlag, *actionFlag, argsInput)
	if err != nil {
		fmt.Printf("Error parsing args: %v\n", err)
		os.Exit(1)
	}

	// Set the parsed args in the operation
	operation.(*pkg.GenericProtocolOperation).DynamicOperation.Args = args

	// Generate the calldata
	calldata, err := operation.GenerateCalldata()
	if err != nil {
		fmt.Printf("Error generating calldata: %v\n", err)
		os.Exit(1)
	}

	// Print the generated calldata
	fmt.Printf("Generated calldata: %s\n", calldata)
}

func parseAndValidateArgs(protocol, action, input string) ([]interface{}, error) {
	var args []interface{}

	// Split the input by comma
	argValues := strings.Split(input, ",")

	// Parse and validate args based on the protocol and action
	switch protocol {
	case "AaveV3", "SparkLend":
		if action == "supply" {
			if len(argValues) != 4 {
				return nil, fmt.Errorf("supply action requires 4 args")
			}
			asset := common.HexToAddress(strings.TrimSpace(argValues[0]))
			amount, ok := new(big.Int).SetString(strings.TrimSpace(argValues[1]), 10)
			if !ok {
				return nil, fmt.Errorf("invalid amount")
			}
			onBehalfOf := common.HexToAddress(strings.TrimSpace(argValues[2]))
			referralCode, err := strconv.ParseUint(strings.TrimSpace(argValues[3]), 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid referral code")
			}
			args = []interface{}{asset, amount, onBehalfOf, uint16(referralCode)}
		} else if action == "withdraw" {
			if len(argValues) != 3 {
				return nil, fmt.Errorf("withdraw action requires 3 args")
			}
			asset := common.HexToAddress(strings.TrimSpace(argValues[0]))
			amount, ok := new(big.Int).SetString(strings.TrimSpace(argValues[1]), 10)
			if !ok {
				return nil, fmt.Errorf("invalid amount")
			}
			to := common.HexToAddress(strings.TrimSpace(argValues[2]))
			args = []interface{}{asset, amount, to}
		} else {
			return nil, fmt.Errorf("unsupported action: %s", action)
		}
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	return args, nil
}

func displayHelp() {
	fmt.Println("Usage:")
	fmt.Println("  protocall [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Supported Protocols:")
	fmt.Println("  - AaveV3")
	fmt.Println("  - SparkLend")
	fmt.Println()
	fmt.Println("Supported Actions:")
	fmt.Println("  - supply")
	fmt.Println("    Args: asset (address), amount (uint256), onBehalfOf (address), referralCode (uint16)")
	fmt.Println("  - withdraw")
	fmt.Println("    Args: asset (address), amount (uint256), to (address)")
}
