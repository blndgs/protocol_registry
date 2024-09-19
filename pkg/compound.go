package pkg

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const cTokenABI = `
[
  {
    "constant": true,
    "inputs": [],
    "name": "underlying",
    "outputs": [
      {
        "name": "",
        "type": "address"
      }
    ],
    "payable": false,
    "stateMutability": "view",
    "type": "function"
  },
	{
    "constant": true,
    "inputs": [],
    "name": "symbol",
    "outputs": [
      {
        "name": "",
        "type": "string"
      }
    ],
    "payable": false,
    "stateMutability": "pure",
    "type": "function"
  }
]
	`

const compoundPoolABI = `
[
  {
    "inputs": [],
    "name": "numAssets",
    "outputs": [
      {
        "internalType": "uint8",
        "name": "",
        "type": "uint8"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "uint8",
        "name": "",
        "type": "uint8"
      }
    ],
    "name": "getAssetInfo",
    "outputs": [
      {
        "internalType": "uint8",
        "name": "offset",
        "type": "uint8"
      },
      {
        "internalType": "address",
        "name": "asset",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "priceFeed",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "scale",
        "type": "uint64"
      },
      {
        "internalType": "uint64",
        "name": "borrowCollateralFactor",
        "type": "uint64"
      },
      {
        "internalType": "uint64",
        "name": "liquidateCollateralFactor",
        "type": "uint64"
      },
      {
        "internalType": "uint64",
        "name": "liquidationFactor",
        "type": "uint64"
      },
      {
        "internalType": "uint128",
        "name": "supplyCap",
        "type": "uint128"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]
	`

const configuratorABI = `
[
  {
    "inputs": [],
    "stateMutability": "nonpayable",
    "type": "constructor"
  },
  {
    "inputs": [],
    "name": "AlreadyInitialized",
    "type": "error"
  },
  {
    "inputs": [],
    "name": "AssetDoesNotExist",
    "type": "error"
  },
  {
    "inputs": [],
    "name": "ConfigurationAlreadyExists",
    "type": "error"
  },
  {
    "inputs": [],
    "name": "InvalidAddress",
    "type": "error"
  },
  {
    "inputs": [],
    "name": "Unauthorized",
    "type": "error"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "components": [
          {
            "internalType": "address",
            "name": "asset",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "priceFeed",
            "type": "address"
          },
          {
            "internalType": "uint8",
            "name": "decimals",
            "type": "uint8"
          },
          {
            "internalType": "uint64",
            "name": "borrowCollateralFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "liquidateCollateralFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "liquidationFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint128",
            "name": "supplyCap",
            "type": "uint128"
          }
        ],
        "indexed": false,
        "internalType": "struct CometConfiguration.AssetConfig",
        "name": "assetConfig",
        "type": "tuple"
      }
    ],
    "name": "AddAsset",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "newComet",
        "type": "address"
      }
    ],
    "name": "CometDeployed",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "oldGovernor",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "newGovernor",
        "type": "address"
      }
    ],
    "name": "GovernorTransferred",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint104",
        "name": "oldBaseBorrowMin",
        "type": "uint104"
      },
      {
        "indexed": false,
        "internalType": "uint104",
        "name": "newBaseBorrowMin",
        "type": "uint104"
      }
    ],
    "name": "SetBaseBorrowMin",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint104",
        "name": "oldBaseMinForRewards",
        "type": "uint104"
      },
      {
        "indexed": false,
        "internalType": "uint104",
        "name": "newBaseMinForRewards",
        "type": "uint104"
      }
    ],
    "name": "SetBaseMinForRewards",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "oldBaseTokenPriceFeed",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "newBaseTokenPriceFeed",
        "type": "address"
      }
    ],
    "name": "SetBaseTokenPriceFeed",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldBaseTrackingBorrowSpeed",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newBaseTrackingBorrowSpeed",
        "type": "uint64"
      }
    ],
    "name": "SetBaseTrackingBorrowSpeed",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldBaseTrackingSupplySpeed",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newBaseTrackingSupplySpeed",
        "type": "uint64"
      }
    ],
    "name": "SetBaseTrackingSupplySpeed",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldKink",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newKink",
        "type": "uint64"
      }
    ],
    "name": "SetBorrowKink",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldIRBase",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newIRBase",
        "type": "uint64"
      }
    ],
    "name": "SetBorrowPerYearInterestRateBase",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldIRSlopeHigh",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newIRSlopeHigh",
        "type": "uint64"
      }
    ],
    "name": "SetBorrowPerYearInterestRateSlopeHigh",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldIRSlopeLow",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newIRSlopeLow",
        "type": "uint64"
      }
    ],
    "name": "SetBorrowPerYearInterestRateSlopeLow",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "components": [
          {
            "internalType": "address",
            "name": "governor",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "pauseGuardian",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "baseToken",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "baseTokenPriceFeed",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "extensionDelegate",
            "type": "address"
          },
          {
            "internalType": "uint64",
            "name": "supplyKink",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "supplyPerYearInterestRateSlopeLow",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "supplyPerYearInterestRateSlopeHigh",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "supplyPerYearInterestRateBase",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowKink",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowPerYearInterestRateSlopeLow",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowPerYearInterestRateSlopeHigh",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowPerYearInterestRateBase",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "storeFrontPriceFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "trackingIndexScale",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "baseTrackingSupplySpeed",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "baseTrackingBorrowSpeed",
            "type": "uint64"
          },
          {
            "internalType": "uint104",
            "name": "baseMinForRewards",
            "type": "uint104"
          },
          {
            "internalType": "uint104",
            "name": "baseBorrowMin",
            "type": "uint104"
          },
          {
            "internalType": "uint104",
            "name": "targetReserves",
            "type": "uint104"
          },
          {
            "components": [
              {
                "internalType": "address",
                "name": "asset",
                "type": "address"
              },
              {
                "internalType": "address",
                "name": "priceFeed",
                "type": "address"
              },
              {
                "internalType": "uint8",
                "name": "decimals",
                "type": "uint8"
              },
              {
                "internalType": "uint64",
                "name": "borrowCollateralFactor",
                "type": "uint64"
              },
              {
                "internalType": "uint64",
                "name": "liquidateCollateralFactor",
                "type": "uint64"
              },
              {
                "internalType": "uint64",
                "name": "liquidationFactor",
                "type": "uint64"
              },
              {
                "internalType": "uint128",
                "name": "supplyCap",
                "type": "uint128"
              }
            ],
            "internalType": "struct CometConfiguration.AssetConfig[]",
            "name": "assetConfigs",
            "type": "tuple[]"
          }
        ],
        "indexed": false,
        "internalType": "struct CometConfiguration.Configuration",
        "name": "oldConfiguration",
        "type": "tuple"
      },
      {
        "components": [
          {
            "internalType": "address",
            "name": "governor",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "pauseGuardian",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "baseToken",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "baseTokenPriceFeed",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "extensionDelegate",
            "type": "address"
          },
          {
            "internalType": "uint64",
            "name": "supplyKink",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "supplyPerYearInterestRateSlopeLow",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "supplyPerYearInterestRateSlopeHigh",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "supplyPerYearInterestRateBase",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowKink",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowPerYearInterestRateSlopeLow",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowPerYearInterestRateSlopeHigh",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowPerYearInterestRateBase",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "storeFrontPriceFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "trackingIndexScale",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "baseTrackingSupplySpeed",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "baseTrackingBorrowSpeed",
            "type": "uint64"
          },
          {
            "internalType": "uint104",
            "name": "baseMinForRewards",
            "type": "uint104"
          },
          {
            "internalType": "uint104",
            "name": "baseBorrowMin",
            "type": "uint104"
          },
          {
            "internalType": "uint104",
            "name": "targetReserves",
            "type": "uint104"
          },
          {
            "components": [
              {
                "internalType": "address",
                "name": "asset",
                "type": "address"
              },
              {
                "internalType": "address",
                "name": "priceFeed",
                "type": "address"
              },
              {
                "internalType": "uint8",
                "name": "decimals",
                "type": "uint8"
              },
              {
                "internalType": "uint64",
                "name": "borrowCollateralFactor",
                "type": "uint64"
              },
              {
                "internalType": "uint64",
                "name": "liquidateCollateralFactor",
                "type": "uint64"
              },
              {
                "internalType": "uint64",
                "name": "liquidationFactor",
                "type": "uint64"
              },
              {
                "internalType": "uint128",
                "name": "supplyCap",
                "type": "uint128"
              }
            ],
            "internalType": "struct CometConfiguration.AssetConfig[]",
            "name": "assetConfigs",
            "type": "tuple[]"
          }
        ],
        "indexed": false,
        "internalType": "struct CometConfiguration.Configuration",
        "name": "newConfiguration",
        "type": "tuple"
      }
    ],
    "name": "SetConfiguration",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "oldExt",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "newExt",
        "type": "address"
      }
    ],
    "name": "SetExtensionDelegate",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "oldFactory",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "newFactory",
        "type": "address"
      }
    ],
    "name": "SetFactory",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "oldGovernor",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "newGovernor",
        "type": "address"
      }
    ],
    "name": "SetGovernor",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "oldPauseGuardian",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "newPauseGuardian",
        "type": "address"
      }
    ],
    "name": "SetPauseGuardian",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldStoreFrontPriceFactor",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newStoreFrontPriceFactor",
        "type": "uint64"
      }
    ],
    "name": "SetStoreFrontPriceFactor",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldKink",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newKink",
        "type": "uint64"
      }
    ],
    "name": "SetSupplyKink",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldIRBase",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newIRBase",
        "type": "uint64"
      }
    ],
    "name": "SetSupplyPerYearInterestRateBase",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldIRSlopeHigh",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newIRSlopeHigh",
        "type": "uint64"
      }
    ],
    "name": "SetSupplyPerYearInterestRateSlopeHigh",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldIRSlopeLow",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newIRSlopeLow",
        "type": "uint64"
      }
    ],
    "name": "SetSupplyPerYearInterestRateSlopeLow",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint104",
        "name": "oldTargetReserves",
        "type": "uint104"
      },
      {
        "indexed": false,
        "internalType": "uint104",
        "name": "newTargetReserves",
        "type": "uint104"
      }
    ],
    "name": "SetTargetReserves",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "components": [
          {
            "internalType": "address",
            "name": "asset",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "priceFeed",
            "type": "address"
          },
          {
            "internalType": "uint8",
            "name": "decimals",
            "type": "uint8"
          },
          {
            "internalType": "uint64",
            "name": "borrowCollateralFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "liquidateCollateralFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "liquidationFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint128",
            "name": "supplyCap",
            "type": "uint128"
          }
        ],
        "indexed": false,
        "internalType": "struct CometConfiguration.AssetConfig",
        "name": "oldAssetConfig",
        "type": "tuple"
      },
      {
        "components": [
          {
            "internalType": "address",
            "name": "asset",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "priceFeed",
            "type": "address"
          },
          {
            "internalType": "uint8",
            "name": "decimals",
            "type": "uint8"
          },
          {
            "internalType": "uint64",
            "name": "borrowCollateralFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "liquidateCollateralFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "liquidationFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint128",
            "name": "supplyCap",
            "type": "uint128"
          }
        ],
        "indexed": false,
        "internalType": "struct CometConfiguration.AssetConfig",
        "name": "newAssetConfig",
        "type": "tuple"
      }
    ],
    "name": "UpdateAsset",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "asset",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldBorrowCF",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newBorrowCF",
        "type": "uint64"
      }
    ],
    "name": "UpdateAssetBorrowCollateralFactor",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "asset",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldLiquidateCF",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newLiquidateCF",
        "type": "uint64"
      }
    ],
    "name": "UpdateAssetLiquidateCollateralFactor",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "asset",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "oldLiquidationFactor",
        "type": "uint64"
      },
      {
        "indexed": false,
        "internalType": "uint64",
        "name": "newLiquidationFactor",
        "type": "uint64"
      }
    ],
    "name": "UpdateAssetLiquidationFactor",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "asset",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "address",
        "name": "oldPriceFeed",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "address",
        "name": "newPriceFeed",
        "type": "address"
      }
    ],
    "name": "UpdateAssetPriceFeed",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "asset",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint128",
        "name": "oldSupplyCap",
        "type": "uint128"
      },
      {
        "indexed": false,
        "internalType": "uint128",
        "name": "newSupplyCap",
        "type": "uint128"
      }
    ],
    "name": "UpdateAssetSupplyCap",
    "type": "event"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "components": [
          {
            "internalType": "address",
            "name": "asset",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "priceFeed",
            "type": "address"
          },
          {
            "internalType": "uint8",
            "name": "decimals",
            "type": "uint8"
          },
          {
            "internalType": "uint64",
            "name": "borrowCollateralFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "liquidateCollateralFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "liquidationFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint128",
            "name": "supplyCap",
            "type": "uint128"
          }
        ],
        "internalType": "struct CometConfiguration.AssetConfig",
        "name": "assetConfig",
        "type": "tuple"
      }
    ],
    "name": "addAsset",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      }
    ],
    "name": "deploy",
    "outputs": [
      {
        "internalType": "address",
        "name": "",
        "type": "address"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "",
        "type": "address"
      }
    ],
    "name": "factory",
    "outputs": [
      {
        "internalType": "address",
        "name": "",
        "type": "address"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "asset",
        "type": "address"
      }
    ],
    "name": "getAssetIndex",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      }
    ],
    "name": "getConfiguration",
    "outputs": [
      {
        "components": [
          {
            "internalType": "address",
            "name": "governor",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "pauseGuardian",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "baseToken",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "baseTokenPriceFeed",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "extensionDelegate",
            "type": "address"
          },
          {
            "internalType": "uint64",
            "name": "supplyKink",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "supplyPerYearInterestRateSlopeLow",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "supplyPerYearInterestRateSlopeHigh",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "supplyPerYearInterestRateBase",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowKink",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowPerYearInterestRateSlopeLow",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowPerYearInterestRateSlopeHigh",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowPerYearInterestRateBase",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "storeFrontPriceFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "trackingIndexScale",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "baseTrackingSupplySpeed",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "baseTrackingBorrowSpeed",
            "type": "uint64"
          },
          {
            "internalType": "uint104",
            "name": "baseMinForRewards",
            "type": "uint104"
          },
          {
            "internalType": "uint104",
            "name": "baseBorrowMin",
            "type": "uint104"
          },
          {
            "internalType": "uint104",
            "name": "targetReserves",
            "type": "uint104"
          },
          {
            "components": [
              {
                "internalType": "address",
                "name": "asset",
                "type": "address"
              },
              {
                "internalType": "address",
                "name": "priceFeed",
                "type": "address"
              },
              {
                "internalType": "uint8",
                "name": "decimals",
                "type": "uint8"
              },
              {
                "internalType": "uint64",
                "name": "borrowCollateralFactor",
                "type": "uint64"
              },
              {
                "internalType": "uint64",
                "name": "liquidateCollateralFactor",
                "type": "uint64"
              },
              {
                "internalType": "uint64",
                "name": "liquidationFactor",
                "type": "uint64"
              },
              {
                "internalType": "uint128",
                "name": "supplyCap",
                "type": "uint128"
              }
            ],
            "internalType": "struct CometConfiguration.AssetConfig[]",
            "name": "assetConfigs",
            "type": "tuple[]"
          }
        ],
        "internalType": "struct CometConfiguration.Configuration",
        "name": "",
        "type": "tuple"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "governor",
    "outputs": [
      {
        "internalType": "address",
        "name": "",
        "type": "address"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "governor_",
        "type": "address"
      }
    ],
    "name": "initialize",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint104",
        "name": "newBaseBorrowMin",
        "type": "uint104"
      }
    ],
    "name": "setBaseBorrowMin",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint104",
        "name": "newBaseMinForRewards",
        "type": "uint104"
      }
    ],
    "name": "setBaseMinForRewards",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "newBaseTokenPriceFeed",
        "type": "address"
      }
    ],
    "name": "setBaseTokenPriceFeed",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newBaseTrackingBorrowSpeed",
        "type": "uint64"
      }
    ],
    "name": "setBaseTrackingBorrowSpeed",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newBaseTrackingSupplySpeed",
        "type": "uint64"
      }
    ],
    "name": "setBaseTrackingSupplySpeed",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newBorrowKink",
        "type": "uint64"
      }
    ],
    "name": "setBorrowKink",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newBase",
        "type": "uint64"
      }
    ],
    "name": "setBorrowPerYearInterestRateBase",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newSlope",
        "type": "uint64"
      }
    ],
    "name": "setBorrowPerYearInterestRateSlopeHigh",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newSlope",
        "type": "uint64"
      }
    ],
    "name": "setBorrowPerYearInterestRateSlopeLow",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "components": [
          {
            "internalType": "address",
            "name": "governor",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "pauseGuardian",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "baseToken",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "baseTokenPriceFeed",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "extensionDelegate",
            "type": "address"
          },
          {
            "internalType": "uint64",
            "name": "supplyKink",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "supplyPerYearInterestRateSlopeLow",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "supplyPerYearInterestRateSlopeHigh",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "supplyPerYearInterestRateBase",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowKink",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowPerYearInterestRateSlopeLow",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowPerYearInterestRateSlopeHigh",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "borrowPerYearInterestRateBase",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "storeFrontPriceFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "trackingIndexScale",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "baseTrackingSupplySpeed",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "baseTrackingBorrowSpeed",
            "type": "uint64"
          },
          {
            "internalType": "uint104",
            "name": "baseMinForRewards",
            "type": "uint104"
          },
          {
            "internalType": "uint104",
            "name": "baseBorrowMin",
            "type": "uint104"
          },
          {
            "internalType": "uint104",
            "name": "targetReserves",
            "type": "uint104"
          },
          {
            "components": [
              {
                "internalType": "address",
                "name": "asset",
                "type": "address"
              },
              {
                "internalType": "address",
                "name": "priceFeed",
                "type": "address"
              },
              {
                "internalType": "uint8",
                "name": "decimals",
                "type": "uint8"
              },
              {
                "internalType": "uint64",
                "name": "borrowCollateralFactor",
                "type": "uint64"
              },
              {
                "internalType": "uint64",
                "name": "liquidateCollateralFactor",
                "type": "uint64"
              },
              {
                "internalType": "uint64",
                "name": "liquidationFactor",
                "type": "uint64"
              },
              {
                "internalType": "uint128",
                "name": "supplyCap",
                "type": "uint128"
              }
            ],
            "internalType": "struct CometConfiguration.AssetConfig[]",
            "name": "assetConfigs",
            "type": "tuple[]"
          }
        ],
        "internalType": "struct CometConfiguration.Configuration",
        "name": "newConfiguration",
        "type": "tuple"
      }
    ],
    "name": "setConfiguration",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "newExtensionDelegate",
        "type": "address"
      }
    ],
    "name": "setExtensionDelegate",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "newFactory",
        "type": "address"
      }
    ],
    "name": "setFactory",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "newGovernor",
        "type": "address"
      }
    ],
    "name": "setGovernor",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "newPauseGuardian",
        "type": "address"
      }
    ],
    "name": "setPauseGuardian",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newStoreFrontPriceFactor",
        "type": "uint64"
      }
    ],
    "name": "setStoreFrontPriceFactor",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newSupplyKink",
        "type": "uint64"
      }
    ],
    "name": "setSupplyKink",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newBase",
        "type": "uint64"
      }
    ],
    "name": "setSupplyPerYearInterestRateBase",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newSlope",
        "type": "uint64"
      }
    ],
    "name": "setSupplyPerYearInterestRateSlopeHigh",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newSlope",
        "type": "uint64"
      }
    ],
    "name": "setSupplyPerYearInterestRateSlopeLow",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "uint104",
        "name": "newTargetReserves",
        "type": "uint104"
      }
    ],
    "name": "setTargetReserves",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "newGovernor",
        "type": "address"
      }
    ],
    "name": "transferGovernor",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "components": [
          {
            "internalType": "address",
            "name": "asset",
            "type": "address"
          },
          {
            "internalType": "address",
            "name": "priceFeed",
            "type": "address"
          },
          {
            "internalType": "uint8",
            "name": "decimals",
            "type": "uint8"
          },
          {
            "internalType": "uint64",
            "name": "borrowCollateralFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "liquidateCollateralFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint64",
            "name": "liquidationFactor",
            "type": "uint64"
          },
          {
            "internalType": "uint128",
            "name": "supplyCap",
            "type": "uint128"
          }
        ],
        "internalType": "struct CometConfiguration.AssetConfig",
        "name": "newAssetConfig",
        "type": "tuple"
      }
    ],
    "name": "updateAsset",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "asset",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newBorrowCF",
        "type": "uint64"
      }
    ],
    "name": "updateAssetBorrowCollateralFactor",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "asset",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newLiquidateCF",
        "type": "uint64"
      }
    ],
    "name": "updateAssetLiquidateCollateralFactor",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "asset",
        "type": "address"
      },
      {
        "internalType": "uint64",
        "name": "newLiquidationFactor",
        "type": "uint64"
      }
    ],
    "name": "updateAssetLiquidationFactor",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "asset",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "newPriceFeed",
        "type": "address"
      }
    ],
    "name": "updateAssetPriceFeed",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "cometProxy",
        "type": "address"
      },
      {
        "internalType": "address",
        "name": "asset",
        "type": "address"
      },
      {
        "internalType": "uint128",
        "name": "newSupplyCap",
        "type": "uint128"
      }
    ],
    "name": "updateAssetSupplyCap",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [],
    "name": "version",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "",
        "type": "uint256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]
	`

// chainID -> Contract address -> ERC20s that can be used as collateral
// Compound has different markets and each market only supports a
// few assets as collateral
var compoundSupportedAssets = map[int64]map[string][]string{
	// Ethereum
	1: {
		// USDC pool
		"0xc3d688b66703497daa19211eedff47f25384cdc3": []string{
			nativeDenomAddress,                           // ETH
			"0x514910771AF9Ca656af840dff83E8264EcF986CA", // LINK
			"0xc00e94Cb662C3520282E6f5717214004A7f26888", // COMP
			"0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984", // UNI
			"0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", // WBTC
		},
		// ETH pool
		"0xa17581a9e3356d9a858b789d68b4d866e593ae94": []string{
			"0xBe9895146f7AF43049ca1c1AE358B0541Ea49704", // cbETH
			"0x7f39C581F595B53c5cb19bD0b3f8dA6c935E2Ca0", // wsETH (Lido)
			"0xae78736Cd615f374D3085123A210448E74Fc6393", // RocketPool ETH
			"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", // WETH
		},
	},
}

// dynamically registers all supported pools
func registerCompoundRegistry(registry ProtocolRegistry, client *ethclient.Client) error {
	for chainID, v := range compoundSupportedAssets {
		for poolAddr := range v {
			c, err := NewCompoundOperation(client, big.NewInt(chainID), common.HexToAddress(poolAddr))
			if err != nil {
				return err
			}

			if err := registry.RegisterProtocol(big.NewInt(chainID), common.HexToAddress(poolAddr), c); err != nil {
				return err
			}
		}
	}

	return nil
}

// CompoundOperation implements the Protocol interface for Ankr
type CompoundOperation struct {
	parsedABI abi.ABI
	contract  common.Address
	chainID   *big.Int
	version   string
	erc20ABI  abi.ABI
	// assets that are supported in this pool
	supportedAssets []string

	client *ethclient.Client

	// no mutex since there are no writes ever
	cTokenMap map[string]string
}

func NewCompoundOperation(client *ethclient.Client, chainID *big.Int,
	marketPool common.Address) (*CompoundOperation, error) {

	parsedABI, err := abi.JSON(strings.NewReader(compoundv3ABI))
	if err != nil {
		return nil, err
	}

	erc20ABI, err := abi.JSON(strings.NewReader(erc20BalanceOfABI))
	if err != nil {
		return nil, err
	}

	supportedChain, ok := compoundSupportedAssets[chainID.Int64()]
	if !ok {
		return nil, errors.New("unsupported chain for Compound in Protocol registry")
	}

	supportedAssets, ok := supportedChain[strings.ToLower(marketPool.Hex())]
	if !ok {
		return nil, errors.New("unsupported Compound pool address")
	}

	cachedCTokens, err := getCTokens(client, marketPool)
	if err != nil {
		return nil, err
	}

	return &CompoundOperation{
		supportedAssets: supportedAssets,
		parsedABI:       parsedABI,
		contract:        marketPool,
		chainID:         chainID,
		version:         "3",
		client:          client,
		erc20ABI:        erc20ABI,
		cTokenMap:       cachedCTokens,
	}, nil
}

func getCTokens(client *ethclient.Client, marketPool common.Address) (map[string]string, error) {

	parsedCTokenABI, err := abi.JSON(strings.NewReader(cTokenABI))
	if err != nil {
		return nil, err
	}

	parsedPoolABI, err := abi.JSON(strings.NewReader(compoundPoolABI))
	if err != nil {
		return nil, err
	}

	numAssetsCallData, err := parsedPoolABI.Pack("numAssets")
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		To:   &marketPool,
		Data: numAssetsCallData,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, err
	}

	var numAssets uint8

	err = parsedPoolABI.UnpackIntoInterface(&numAssets, "numAssets", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack output: %v", err)
	}

	// Fetch info for each collateral asset
	// for i := uint8(0); i < numAssets; i++ {
	// 	var assetInfo struct {
	// 		Offset                    uint8
	// 		Asset                     common.Address
	// 		PriceFeed                 common.Address
	// 		Scale                     uint64
	// 		BorrowCollateralFactor    uint64
	// 		LiquidateCollateralFactor uint64
	// 		LiquidationFactor         uint64
	// 		SupplyCap                 *big.Int
	// 	}
	//
	// 	assetInfoCalldata, err := parsedPoolABI.Pack("getAssetInfo", i)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	msg := ethereum.CallMsg{
	// 		To:   &marketPool,
	// 		Data: assetInfoCalldata,
	// 	}
	//
	// 	result, err := client.CallContract(context.Background(), msg, nil)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	err = parsedPoolABI.UnpackIntoInterface(&assetInfo, "getAssetInfo", result)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to unpack output: %v", err)
	// 	}
	// }

	_ = parsedCTokenABI

	configuratorAddress := common.HexToAddress("0x316f9708bB98af7dA9c68C1C3b5e79039cD336E3")
	configuratorABI, err := abi.JSON(strings.NewReader(configuratorABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse configurator ABI: %v", err)
	}

	input, err := configuratorABI.Pack("getConfiguration", marketPool)
	if err != nil {
		return nil, fmt.Errorf("failed to pack input: %v", err)
	}

	msg = ethereum.CallMsg{
		To:   &configuratorAddress,
		Data: input,
	}

	output, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call configurator contract: %v", err)
	}

	type configs struct {
		Governor                           common.Address `json:"governor"`
		PauseGuardian                      common.Address `json:"pauseGuardian"`
		BaseToken                          common.Address `json:"baseToken"`
		BaseTokenPriceFeed                 common.Address `json:"baseTokenPriceFeed"`
		ExtensionDelegate                  common.Address `json:"extensionDelegate"`
		SupplyKink                         uint64         `json:"supplyKink"`
		SupplyPerYearInterestRateSlopeLow  uint64         `json:"supplyPerYearInterestRateSlopeLow"`
		SupplyPerYearInterestRateSlopeHigh uint64         `json:"supplyPerYearInterestRateSlopeHigh"`
		SupplyPerYearInterestRateBase      uint64         `json:"supplyPerYearInterestRateBase"`
		BorrowKink                         uint64         `json:"borrowKink"`
		BorrowPerYearInterestRateSlopeLow  uint64         `json:"borrowPerYearInterestRateSlopeLow"`
		BorrowPerYearInterestRateSlopeHigh uint64         `json:"borrowPerYearInterestRateSlopeHigh"`
		BorrowPerYearInterestRateBase      uint64         `json:"borrowPerYearInterestRateBase"`
		StoreFrontPriceFactor              uint64         `json:"storeFrontPriceFactor"`
		TrackingIndexScale                 uint64         `json:"trackingIndexScale"`
		BaseTrackingSupplySpeed            uint64         `json:"baseTrackingSupplySpeed"`
		BaseTrackingBorrowSpeed            uint64         `json:"baseTrackingBorrowSpeed"`
		BaseMinForRewards                  *big.Int       `json:"baseMinForRewards"`
		BaseBorrowMin                      *big.Int       `json:"baseBorrowMin"`
		TargetReserves                     *big.Int       `json:"targetReserves"`
		AssetConfigs                       []struct {
			Asset                     common.Address `json:"asset"`
			PriceFeed                 common.Address `json:"priceFeed"`
			Decimals                  uint8          `json:"decimals"`
			BorrowCollateralFactor    uint64         `json:"borrowCollateralFactor"`
			LiquidateCollateralFactor uint64         `json:"liquidateCollateralFactor"`
			LiquidationFactor         uint64         `json:"liquidationFactor"`
			SupplyCap                 *big.Int       `json:"supplyCap"`
		} `json:"assetConfigs"`
	}

	var c configs
	res, err := configuratorABI.Unpack("getConfiguration", output)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack output: %v", err)
	}

	var b = new(bytes.Buffer)

	if err := json.NewEncoder(b).Encode(res[0]); err != nil {
		return nil, err
	}

	if err := json.NewDecoder(b).Decode(&c); err != nil {
		return nil, err
	}

	wrappedTokens := make(map[common.Address]common.Address)

	for _, assetConfig := range c.AssetConfigs {
		// wrappedTokens[assetConfig.Asset] = ass
	}

	fmt.Println(wrappedTokens)

	//
	// var markets []common.Address
	// err = parsedComptrollerABI.UnpackIntoInterface(&markets, "getAllMarkets", result)
	// if err != nil {
	// 	return nil, err
	// }
	//
	var cachedCTokens = make(map[string]string)
	//
	// underlyingCalldata, err := parsedComptrollerABI.Pack("underlying")
	// if err != nil {
	// 	return nil, err
	// }

	// for _, marketAddress := range markets {
	//
	// 	// Get the underlying token for this market
	// 	// All tokens have an underlying token except cETH
	// 	//
	// 	// There is an edge case here where we check for an invalid opcode.
	// 	// This is because of Tenderly (and maybe other custom RPCs?).
	// 	// We have to skip this error because we still need to verify if the market we
	// 	// are in is cETH or not
	// 	msg := ethereum.CallMsg{
	// 		To:   &marketAddress,
	// 		Data: underlyingCalldata,
	// 	}
	// 	result, err := client.CallContract(context.Background(), msg, nil)
	// 	if err != nil && !strings.Contains(err.Error(), "invalid opcode") {
	// 		return nil, err
	// 	}
	//
	// 	var underlying common.Address
	// 	err = parsedComptrollerABI.UnpackIntoInterface(&underlying, "underlying", result)
	// 	if err != nil {
	//
	// 		// cETH does not have an underlying token
	// 		// so check if it is a token with the cETH symbol.
	// 		// If it is one, we have to add the nativeDenomAddress here too
	// 		symbolCalldata, err := parsedComptrollerABI.Pack("symbol")
	// 		if err != nil {
	// 			return nil, fmt.Errorf("could not pack symbol to check if cETH")
	// 		}
	//
	// 		msg := ethereum.CallMsg{
	// 			To:   &marketAddress,
	// 			Data: symbolCalldata,
	// 		}
	//
	// 		result, err := client.CallContract(context.Background(), msg, nil)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	//
	// 		var symbol string
	// 		err = parsedComptrollerABI.UnpackIntoInterface(&symbol, "symbol", result)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	//
	// 		if symbol == "cETH" {
	// 			cachedCTokens[common.HexToAddress(nativeDenomAddress).Hex()] = marketAddress.Hex()
	// 		}
	//
	// 		continue
	// 	}
	//
	// 	cachedCTokens[underlying.Hex()] = marketAddress.Hex()
	// }

	return cachedCTokens, err
}

// GenerateCalldata creates the necessary blockchain transaction data
func (a *CompoundOperation) GenerateCalldata(ctx context.Context, chainID *big.Int,
	action ContractAction, params TransactionParams) (string, error) {
	if chainID.Int64() != 1 {
		return "", ErrChainUnsupported
	}

	switch action {
	case LoanSupply:
		return a.supply(params)
	case LoanWithdraw:
		return a.withdraw(params)
	default:
		return "", errors.New("unsupported operation")
	}
}

func (c *CompoundOperation) withdraw(opts TransactionParams) (string, error) {
	calldata, err := c.parsedABI.Pack("withdraw", opts.Asset, opts.Amount)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", "withdraw", err)
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

func (c *CompoundOperation) supply(opts TransactionParams) (string, error) {

	calldata, err := c.parsedABI.Pack("supply", opts.Asset, opts.Amount)
	if err != nil {
		return "", fmt.Errorf("failed to generate calldata for %s: %w", "supply", err)
	}

	return HexPrefix + hex.EncodeToString(calldata), nil
}

// Validate checks if the provided parameters are valid for the specified action
func (l *CompoundOperation) Validate(ctx context.Context,
	chainID *big.Int, action ContractAction, params TransactionParams) error {

	if chainID.Int64() != 1 {
		return ErrChainUnsupported
	}

	if !l.IsSupportedAsset(ctx, l.chainID, params.Asset) {
		return fmt.Errorf("asset not supported %s", params.Asset)
	}

	if action != LoanSupply && action != LoanWithdraw {
		return errors.New("action not supported")
	}

	if action == LoanSupply {
		return nil
	}

	if params.Amount.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("amount must be greater than zero")
	}

	_, balance, err := l.GetBalance(ctx, l.chainID, params.Sender, params.Asset)
	if err != nil {
		return err
	}

	if balance.Cmp(params.Amount) == -1 {
		return errors.New("your balance not enough")
	}

	return nil
}

// GetBalance retrieves the balance for a specified account and asset
func (l *CompoundOperation) GetBalance(ctx context.Context, chainID *big.Int,
	account, asset common.Address) (common.Address, *big.Int, error) {

	var address common.Address

	if chainID.Int64() != 1 {
		return address, nil, ErrChainUnsupported
	}

	cToken, ok := l.cTokenMap[asset.Hex()]
	if !ok {
		return address, nil, errors.New("token does not have an equivalent cToken")
	}

	callData, err := l.erc20ABI.Pack("balanceOf", account)
	if err != nil {
		return address, nil, err
	}

	var assetHex = common.HexToAddress(cToken)
	result, err := l.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &assetHex,
		Data: callData,
	}, nil)
	if err != nil {
		return address, nil, err
	}

	balance := new(big.Int)
	err = l.erc20ABI.UnpackIntoInterface(&balance, "balanceOf", result)
	return assetHex, balance, err
}

// GetSupportedAssets returns a list of assets supported by the protocol on the specified chain
func (c *CompoundOperation) GetSupportedAssets(ctx context.Context, chainID *big.Int) ([]common.Address, error) {
	var addrs = make([]common.Address, 0, len(c.supportedAssets))

	for _, v := range c.supportedAssets {
		addrs = append(addrs, common.HexToAddress(v))
	}

	return addrs, nil
}

// IsSupportedAsset checks if the specified asset is supported on the given chain
func (c *CompoundOperation) IsSupportedAsset(ctx context.Context, chainID *big.Int, asset common.Address) bool {
	if chainID.Int64() != 1 {
		return false
	}

	for _, addr := range c.supportedAssets {
		if strings.EqualFold(strings.ToLower(asset.Hex()), strings.ToLower(addr)) {
			return true
		}
	}

	return false
}

// GetProtocolConfig returns the protocol config for a specific chain
func (l *CompoundOperation) GetProtocolConfig(chainID *big.Int) ProtocolConfig {
	return ProtocolConfig{
		ChainID:  l.chainID,
		Contract: l.contract,
		ABI:      l.parsedABI,
		Type:     TypeStake,
	}
}

// GetABI returns the ABI of the protocol's contract
func (l *CompoundOperation) GetABI(chainID *big.Int) abi.ABI { return l.parsedABI }

// GetType returns the protocol type
func (l *CompoundOperation) GetType() ProtocolType { return TypeLoan }

// GetContractAddress returns the contract address for a specific chain
func (l *CompoundOperation) GetContractAddress(chainID *big.Int) common.Address { return l.contract }

// Name returns the human readable name for the protocol
func (l *CompoundOperation) GetName() string { return Compound }

// GetVersion returns the version of the protocol
func (l *CompoundOperation) GetVersion() string { return l.version }
