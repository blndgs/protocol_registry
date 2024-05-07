package pkg

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// MatchType checks if the provided Go type matches the expected ABI type.
func MatchType(abiType abi.Type, value interface{}) bool {
	switch abiType.T {
	case abi.IntTy, abi.UintTy:
		switch abiType.Size {
		case 8:
			_, ok := value.(uint8)
			return ok
		case 16:
			_, ok := value.(uint16)
			return ok
		case 32:
			_, ok := value.(uint32)
			return ok
		case 64:
			_, ok := value.(uint64)
			return ok
		default:
			_, ok := value.(*big.Int)
			return ok
		}
	case abi.BoolTy:
		_, ok := value.(bool)
		return ok
	case abi.StringTy:
		_, ok := value.(string)
		return ok
	case abi.AddressTy:
		_, ok := value.(common.Address)
		return ok
	case abi.BytesTy, abi.FixedBytesTy:
		_, ok := value.([]byte)
		return ok
	default:
		return false
	}
}
