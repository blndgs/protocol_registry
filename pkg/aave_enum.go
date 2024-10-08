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
	// AaveProtocolDeploymentEthereum is a AaveProtocolDeployment of type Ethereum.
	AaveProtocolDeploymentEthereum AaveProtocolDeployment = iota
	// AaveProtocolDeploymentSpark is a AaveProtocolDeployment of type Spark.
	AaveProtocolDeploymentSpark
	// AaveProtocolDeploymentAvalonFinance is a AaveProtocolDeployment of type Avalon_finance.
	AaveProtocolDeploymentAvalonFinance
	// AaveProtocolDeploymentPolygon is a AaveProtocolDeployment of type Polygon.
	AaveProtocolDeploymentPolygon
)

var ErrInvalidAaveProtocolDeployment = errors.New("not a valid AaveProtocolDeployment")

const _AaveProtocolDeploymentName = "ethereumsparkavalon_financepolygon"

var _AaveProtocolDeploymentMap = map[AaveProtocolDeployment]string{
	AaveProtocolDeploymentEthereum:      _AaveProtocolDeploymentName[0:8],
	AaveProtocolDeploymentSpark:         _AaveProtocolDeploymentName[8:13],
	AaveProtocolDeploymentAvalonFinance: _AaveProtocolDeploymentName[13:27],
	AaveProtocolDeploymentPolygon:       _AaveProtocolDeploymentName[27:34],
}

// String implements the Stringer interface.
func (x AaveProtocolDeployment) String() string {
	if str, ok := _AaveProtocolDeploymentMap[x]; ok {
		return str
	}
	return fmt.Sprintf("AaveProtocolDeployment(%d)", x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x AaveProtocolDeployment) IsValid() bool {
	_, ok := _AaveProtocolDeploymentMap[x]
	return ok
}

var _AaveProtocolDeploymentValue = map[string]AaveProtocolDeployment{
	_AaveProtocolDeploymentName[0:8]:   AaveProtocolDeploymentEthereum,
	_AaveProtocolDeploymentName[8:13]:  AaveProtocolDeploymentSpark,
	_AaveProtocolDeploymentName[13:27]: AaveProtocolDeploymentAvalonFinance,
	_AaveProtocolDeploymentName[27:34]: AaveProtocolDeploymentPolygon,
}

// ParseAaveProtocolDeployment attempts to convert a string to a AaveProtocolDeployment.
func ParseAaveProtocolDeployment(name string) (AaveProtocolDeployment, error) {
	if x, ok := _AaveProtocolDeploymentValue[name]; ok {
		return x, nil
	}
	return AaveProtocolDeployment(0), fmt.Errorf("%s is %w", name, ErrInvalidAaveProtocolDeployment)
}
