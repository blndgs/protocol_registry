//go:build integration
// +build integration
//
// Code generated by go-enum DO NOT EDIT.
// Version:
// Revision:
// Build Date:
// Built By:

package pkg

import (
	"errors"
	"fmt"
)

const (
	// ChainETH is a Chain of type ETH.
	ChainETH Chain = "ETH"
	// ChainBSC is a Chain of type BSC.
	ChainBSC Chain = "BSC"
	// ChainPOLYGON is a Chain of type POLYGON.
	ChainPOLYGON Chain = "POLYGON"
)

var ErrInvalidChain = errors.New("not a valid Chain")

// String implements the Stringer interface.
func (x Chain) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x Chain) IsValid() bool {
	_, err := ParseChain(string(x))
	return err == nil
}

var _ChainValue = map[string]Chain{
	"ETH":     ChainETH,
	"BSC":     ChainBSC,
	"POLYGON": ChainPOLYGON,
}

// ParseChain attempts to convert a string to a Chain.
func ParseChain(name string) (Chain, error) {
	if x, ok := _ChainValue[name]; ok {
		return x, nil
	}
	return Chain(""), fmt.Errorf("%s is %w", name, ErrInvalidChain)
}
