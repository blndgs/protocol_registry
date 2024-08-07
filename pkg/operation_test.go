package pkg

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestGenerateCalldataOptions_UBO(t *testing.T) {

	tt := []struct {
		opts     GenerateCalldataOptions
		expected common.Address
		name     string
	}{
		{
			name: "sender",
			opts: GenerateCalldataOptions{
				Sender: common.HexToAddress(nativeDenomAddress),
			},
			expected: common.HexToAddress(nativeDenomAddress),
		},
		{
			name: "recipient",
			opts: GenerateCalldataOptions{
				Recipient: common.HexToAddress(nativeDenomAddress),
			},
			expected: common.HexToAddress(nativeDenomAddress),
		},
		{
			name: "both recipient and sender are set - recipient takes precedence",
			opts: GenerateCalldataOptions{
				Recipient: common.HexToAddress(nativeDenomAddress),
				Sender:    common.HexToAddress("0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2"),
			},
			expected: common.HexToAddress(nativeDenomAddress),
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			require.Equal(t, v.expected, v.opts.UBO())
		})
	}

}
