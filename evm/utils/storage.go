package utils

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/weitrue/kit/evm/utils/types"
	"strings"
)

func StorageValue(ctx context.Context, c *rpc.Client, contractAddress, slot string) ([]byte, error) {
	client := ethclient.NewClient(c)
	storageAt, err := client.StorageAt(ctx, common.HexToAddress(contractAddress), common.HexToHash(slot), nil)
	if err != nil {
		return nil, err
	}

	return storageAt, nil
}

func ParseStorageLayout(ctx context.Context, c *rpc.Client, contract, storage, abiStr string) (any, error) {
	abiO, err := DecodeABI(abiStr)
	if err != nil {
		return nil, err
	}

	storages := new(types.ContractStorage)
	err = json.Unmarshal([]byte(storage), &storages)
	if err != nil {
		return nil, errors.New("invalid contract")
	}

	contractVariables := make([]types.ContractVariable, 0)
	for _, v := range storages.Storage {
		variable := types.ContractVariable{
			Name: v.Label,
			Type: v.Type,
		}
		if t, ok := storages.Types[v.Type]; ok {
			variable.Type = t.Label
		} else {
			variable.Type = strings.ReplaceAll(v.Label, "t_", "")
		}

		if method, ok := abiO.Methods[v.Label]; ok {
			if len(method.Outputs) == 1 {
				var ret any
				switch method.Outputs[0].Type.T {
				case abi.IntTy, abi.UintTy, abi.BoolTy, abi.StringTy, abi.AddressTy:
					ret, err = Call(ctx, c, contract, method)
				case abi.FixedBytesTy, abi.HashTy, abi.FixedPointTy, abi.FunctionTy:
				case abi.BytesTy, abi.SliceTy:
				case abi.ArrayTy:
				case abi.TupleTy:

				}
				if err == nil {
					variable.Value = ret
				}

				contractVariables = append(contractVariables, variable)
			}
		}

	}

	return contractVariables, nil
}
