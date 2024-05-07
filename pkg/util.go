package pkg

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// ConvertToABIType converts a Go value to the corresponding ABI type.
func ConvertToABIType(abiType abi.Type, value interface{}) interface{} {
	switch abiType.T {
	case abi.IntTy, abi.UintTy:
		switch v := value.(type) {
		case *big.Int:
			return v
		case string:
			bi, ok := new(big.Int).SetString(v, 10)
			if !ok {
				panic(fmt.Sprintf("invalid big.Int string: %s", v))
			}
			return bi
		case uint16:
			if abiType.Size < 256 {
				return v
			}
			return new(big.Int).SetUint64(uint64(v))
		default:
			panic(fmt.Sprintf("unsupported value type for %s: %T", abiType.String(), value))
		}
	case abi.BoolTy:
		return value.(bool)
	case abi.StringTy:
		return value.(string)
	case abi.AddressTy:
		switch v := value.(type) {
		case common.Address:
			return v
		case string:
			return common.HexToAddress(v)
		default:
			panic(fmt.Sprintf("unsupported value type for %s: %T", abiType.String(), value))
		}
	case abi.BytesTy, abi.FixedBytesTy:
		switch v := value.(type) {
		case []byte:
			return v
		case string:
			return common.FromHex(v)
		default:
			panic(fmt.Sprintf("unsupported value type for %s: %T", abiType.String(), value))
		}
	default:
		panic(fmt.Sprintf("unsupported ABI type: %s", abiType.String()))
	}
}

// MatchType checks if the provided Go type matches the expected ABI type.
func MatchType(abiType abi.Type, value interface{}) bool {
	switch abiType.T {
	case abi.IntTy:
		if _, ok := value.(*big.Int); ok {
			return true
		}
		if v, ok := value.(string); ok {
			_, ok := new(big.Int).SetString(v, 10)
			return ok
		}
		return false
	case abi.UintTy:
		if _, ok := value.(*big.Int); ok {
			return true
		}
		if v, ok := value.(string); ok {
			_, ok := new(big.Int).SetString(v, 10)
			return ok
		}
		if _, ok := value.(uint16); ok {
			return abiType.Size <= 16
		}
		return false
	case abi.BoolTy:
		_, ok := value.(bool)
		return ok
	case abi.StringTy:
		_, ok := value.(string)
		return ok
	case abi.AddressTy:
		if _, ok := value.(common.Address); ok {
			return true
		}
		if v, ok := value.(string); ok {
			return common.IsHexAddress(v)
		}
		return false
	case abi.BytesTy, abi.FixedBytesTy:
		if _, ok := value.([]byte); ok {
			return true
		}
		if v, ok := value.(string); ok {
			return IsHex(v)
		}
		return false
	default:
		return false
	}
}

// IsHex checks if a string represents a valid hex value.
func IsHex(s string) bool {
	s = strings.TrimPrefix(s, "0x")
	_, err := hex.DecodeString(s)
	return err == nil
}
