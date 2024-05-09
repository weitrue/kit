package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/weitrue/kit/evm/utils/types"
)

func StorageAt(ctx context.Context, c *rpc.Client, contract string, slot common.Hash) ([]byte, error) {
	var (
		res hexutil.Bytes
	)
	err := c.CallContext(ctx, &res, "eth_getStorageAt", common.HexToAddress(contract), slot, "latest")
	if err != nil {
		return nil, err
	}

	return res, nil
}

func ParseStorageLayout(ctx context.Context, c *rpc.Client, contract, storage string) (any, error) {
	storages := new(types.ContractStorage)
	err := json.Unmarshal([]byte(storage), &storages)
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
			slotP, _ := new(big.Int).SetString(v.Slot, 10)
			slotKey := common.BigToHash(slotP)
			if strings.HasPrefix(v.Type, "t_array") { // 数组类型
				if !strings.HasSuffix(t.Label, "[]") { //定长
					siz := t.Label[strings.Index(t.Label, "[")+1 : len(t.Label)-1]
					s, _ := strconv.Atoi(siz)
					datas := make([]any, s)
					for i := int64(0); i < int64(s); i++ {
						data, err := StorageAt(ctx, c, contract, slotKey)
						if err == nil {
							privateVariable, err := unpackPrivateVariable(data, storages.Types, t.Base, v.Offset)
							if err == nil {
								if privateVariable != nil {
									datas[i] = privateVariable.Value
								}
							}
						}
						dataSlotBig := new(big.Int).Add(slotKey.Big(), big.NewInt(1))
						slotKey = common.BigToHash(dataSlotBig)
					}
					variable.Value = datas
				} else {
					slot := crypto.Keccak256(slotP.Bytes())
					slot = append(slot, big.NewInt(1).Bytes()...)
					data, err := StorageAt(ctx, c, contract, slotKey)
					if err == nil {
						privateVariable, err := unpackPrivateVariable(data, storages.Types, t.Base, v.Offset)
						if err == nil {
							if privateVariable != nil {
								variable.Value = privateVariable.Value
							}
						}
					}
				}

			} else if strings.HasPrefix(v.Type, "t_mapping") {

			} else if strings.HasPrefix(v.Type, "t_string") {
				data, err := StorageAt(ctx, c, contract, slotKey)
				if err == nil {
					dataS := common.Bytes2Hex(data[31:])
					decimal, err := strconv.ParseInt(dataS, 16, 64)
					if err == nil && decimal%2 > 0 {
						size := (decimal - 1) / 2
						dataSlotKey := crypto.Keccak256Hash(slotKey.Bytes())
						val := ""
						for size/32 > 0 {
							rawData, err := StorageAt(ctx, c, contract, dataSlotKey)
							if err == nil {
								val += string(rawData)
							}
							dataSlotBig := new(big.Int).Add(dataSlotKey.Big(), big.NewInt(1))
							dataSlotKey = common.BigToHash(dataSlotBig)
							size -= 32
						}

						if size > 0 {
							rawData, err := StorageAt(ctx, c, contract, dataSlotKey)
							if err == nil {
								val += string(rawData[:size])
							}
						}
						if len(val) > 0 {
							variable.Value = val
						}
					} else {
						privateVariable, err := unpackPrivateVariable(data, storages.Types, v.Type, v.Offset)
						if err == nil {
							if privateVariable != nil {
								variable.Value = privateVariable.Value
							}
						}
					}
				}
			} else {
				data, err := StorageAt(ctx, c, contract, slotKey)
				if err == nil {
					privateVariable, err := unpackPrivateVariable(data, storages.Types, v.Type, v.Offset)
					if err == nil {
						if privateVariable != nil {
							variable.Value = privateVariable.Value
						}
					}
				}
			}
		}

		fmt.Println(variable.String())
	}

	return contractVariables, nil
}

func keccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

func ParseVyPerStorageLayout(ctx context.Context, c *rpc.Client, contract, storage, abiStr string) (any, error) {
	value, _, _, err := jsonparser.Get([]byte(storage), "storage_layout")
	if err == nil && len(value) > 0 {
		type VyPerStorage struct {
			StorageLayout map[string]types.VyPerStorage `json:"storage_layout"`
			CodeLayout    map[string]types.VyPerLayout  `json:"code_layout"`
		}

		data := new(VyPerStorage)
		err = json.Unmarshal([]byte(storage), &data)
		if err == nil {
			return ParseVyPerStorage(ctx, c, contract, abiStr, data.StorageLayout)
		}
	}

	if strings.Contains(storage, "nonreentrant.") {
		data := make(map[string]types.VyPerStorage)
		err = json.Unmarshal([]byte(storage), &data)
		if err == nil {
			return ParseVyPerStorage(ctx, c, contract, abiStr, data)
		}
	}

	return nil, types.ErrUnsupportedStorage
}

func ParseVyPerStorage(ctx context.Context, c *rpc.Client, contract, abiStr string, storageLayout map[string]types.VyPerStorage) ([]types.ContractVariable, error) {
	contractVariables := make([]types.ContractVariable, 0)
	abiO, err := DecodeABI(abiStr)
	if err != nil {
		return contractVariables, err
	}

	if len(storageLayout) == 0 {
		return contractVariables, nil
	}

	storages := make(types.VyPerStorageLayouts, 0)
	for k, v := range storageLayout {
		if strings.Contains(k, "nonreentrant.") {
			continue
		}

		storages = append(storages, &types.VyPerStorageLayout{
			Label:        k,
			VyPerStorage: v,
		})
	}

	sort.Sort(storages)
	for _, v := range storages {
		if ok, size := isSupported(v.Type); ok {
			variable := types.ContractVariable{
				Name: v.Label,
				Type: v.Type,
			}
			if method, exist := abiO.Methods[variable.Name]; exist { //public variable
				if len(method.Inputs) == 0 {
					value, err := Call(ctx, c, contract, method)
					if err == nil {
						if strings.HasPrefix(v.Type, "int") || strings.HasPrefix(v.Type, "uint") {
							if vv, can := value.(*big.Int); can {
								value = vv.String()
							}
						}

						//if strings.HasPrefix(v.Type, "bytes") {
						//	if len(v.Type) > 5 {
						//		le := v.Type[5:]
						//		siz, _ := strconv.Atoi(le)
						//		switch siz {
						//		case 1:
						//			if vv, can := value.([1]uint8); can {
						//				value = strings.TrimRight(string(vv[:]), "\x00")
						//			}
						//		case 2:
						//			if vv, can := value.([2]uint8); can {
						//				value = strings.TrimRight(string(vv[:]), "\x00")
						//			}
						//		case 4:
						//			if vv, can := value.([4]uint8); can {
						//				value = strings.TrimRight(string(vv[:]), "\x00")
						//			}
						//		case 8:
						//			if vv, can := value.([8]uint8); can {
						//				value = strings.TrimRight(string(vv[:]), "\x00")
						//			}
						//		case 16:
						//			if vv, can := value.([16]uint8); can {
						//				value = strings.TrimRight(string(vv[:]), "\x00")
						//			}
						//		case 32:
						//			if vv, can := value.([32]uint8); can {
						//				value = strings.TrimRight(string(vv[:]), "\x00")
						//			}
						//		case 64:
						//			if vv, can := value.([64]uint8); can {
						//				value = strings.TrimRight(string(vv[:]), "\x00")
						//			}
						//		case 128:
						//			if vv, can := value.([128]uint8); can {
						//				value = strings.TrimRight(string(vv[:]), "\x00")
						//			}
						//		case 256:
						//			if vv, can := value.([256]uint8); can {
						//				value = strings.TrimRight(string(vv[:]), "\x00")
						//			}
						//		default:
						//
						//		}
						//	}
						//
						//}

						variable.Value = value
					}
					fmt.Println(variable)
					contractVariables = append(contractVariables, variable)
				}
			} else {
				slot := big.NewInt(v.Slot)
				data, err := StorageAt(ctx, c, contract, common.BigToHash(slot))
				if err == nil {
					keyData, err := extractData(data, 0, size)
					if err != nil {
						return nil, err
					}
					if err == nil {
						variable.Value = parseVyPerValue(v.Type, keyData)
					}
				}
				fmt.Println(variable)
				contractVariables = append(contractVariables, variable)
			}
		}
	}

	return contractVariables, nil
}

func unpackPublicVariable(ctx context.Context, c *rpc.Client, contract string, method abi.Method, args ...any) (any, error) {
	if len(method.Inputs) == 0 {
		switch method.Outputs[0].Type.T {
		case abi.IntTy, abi.UintTy, abi.BoolTy, abi.StringTy, abi.AddressTy:
		case abi.FixedBytesTy, abi.HashTy, abi.FixedPointTy, abi.FunctionTy:
		case abi.BytesTy, abi.SliceTy:
		case abi.ArrayTy:
		case abi.TupleTy:
		}

		return Call(ctx, c, contract, method)
	}

	return nil, nil
}

func unpackPrivateVariable(data []byte, allType map[string]types.StorageKeyType, keyType string, offSet int) (*types.ContractVariable, error) {
	if t, ok := allType[keyType]; ok {
		if types.IsDynamicType(keyType) {
			size, _ := strconv.Atoi(t.NumberOfBytes)
			keyData, err := extractData(data, offSet, size)
			if err != nil {
				return nil, err
			}

			return &types.ContractVariable{
				Type:   t.Label,
				IsBase: true,
				Value:  parseValue(keyType, keyData),
			}, nil
		}

		if len(t.Members) > 0 { // struct
			values := make([]*types.ContractVariable, 0)
			for _, m := range t.Members {
				mData, err := unpackPrivateVariable(data, allType, m.Type, m.Offset)
				if err == nil {
					mData.Name = m.Label
					values = append(values, mData)
				}
			}
			return &types.ContractVariable{
				Type:  t.Label,
				Value: values,
			}, nil
		}

		if strings.HasPrefix(keyType, "t_array") {
			if !strings.HasSuffix(t.Label, "[]") {
				fmt.Println(t.Label)
			} else {

			}
		}

	}

	return nil, nil
}

func extractData(data []byte, offset, size int) ([]byte, error) {
	end := common.HashLength - offset
	if end < size {
		return nil, types.ErrSize
	}
	start := end - size
	return data[start:end], nil
}

// 返回变量的字节数
func getTypeSize(typeName string) uint64 {
	if strings.HasPrefix(typeName, "t_bool") {
		return 1
	} else if strings.HasPrefix(typeName, "t_enum") {
		return 32
	} else if strings.HasPrefix(typeName, "t_uint") {
		sizeStr := typeName[6:]
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			panic("err in parse type size")
		}
		return uint64(size / 8)
	} else if strings.HasPrefix(typeName, "t_int") {
		sizeStr := typeName[5:]
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			panic("err in parse type size")
		}
		return uint64(size / 8)
	} else if strings.HasPrefix(typeName, "t_address") || strings.HasPrefix(typeName, "t_contract") {
		return 20
	} else if strings.HasPrefix(typeName, "t_string") {
	} else if strings.HasPrefix(typeName, "t_bytes") {
		// TODO: byte32 类型
		return 32
	} else {
		// panic("not impl!")
	}
	return 0
}

func parseValue(typeName string, data []byte) any {
	if strings.HasPrefix(typeName, "t_bool") {
		return new(big.Int).SetBytes(data).Cmp(common.Big0) > 0
	}
	if strings.HasPrefix(typeName, "t_uint") || strings.HasPrefix(typeName, "t_enum") {
		return new(big.Int).SetBytes(data)
	}

	if strings.HasPrefix(typeName, "t_int") {
		return new(big.Int).SetBytes(data)
	}

	if strings.HasPrefix(typeName, "t_address") || strings.HasPrefix(typeName, "t_contract") {
		return common.BytesToAddress(data).Hex()
	}

	if strings.HasPrefix(typeName, "t_string") {
		i := strings.Index(string(data), "\u0000")
		return string(data)[:i]
	}

	if strings.HasPrefix(typeName, "t_bytes") {
		return common.Bytes2Hex(data)
	}

	return string(data)
}

func isSupported(typeName string) (bool, int) {
	if strings.HasSuffix(typeName, "]") {
		return false, 0
	}

	if strings.HasPrefix(typeName, "bool") {
		return true, 1
	}

	if strings.HasPrefix(typeName, "uint") {
		sizeStr := typeName[4:]
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return false, 0
		}
		return true, size / 8
	}
	if strings.HasPrefix(typeName, "int") {
		sizeStr := typeName[3:]
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return false, 0
		}
		return true, size / 8
	}
	if strings.HasPrefix(typeName, "address") || strings.HasPrefix(typeName, "contract") {
		return true, 20
	}
	if strings.HasPrefix(typeName, "bytes") {
		return true, 32
	}

	return false, 0
}

func parseVyPerValue(typeName string, data []byte) any {
	if strings.HasPrefix(typeName, "bool") {
		return new(big.Int).SetBytes(data).Cmp(common.Big0) > 0
	}
	if strings.HasPrefix(typeName, "uint") || strings.HasPrefix(typeName, "enum") {
		return new(big.Int).SetBytes(data).String()
	}

	if strings.HasPrefix(typeName, "int") {
		return new(big.Int).SetBytes(data).String()
	}

	if strings.HasPrefix(typeName, "address") || strings.HasPrefix(typeName, "contract") {
		return strings.ToLower(common.BytesToAddress(data).Hex())
	}

	if strings.HasPrefix(typeName, "string") {
		i := strings.Index(string(data), "\u0000")
		return string(data)[:i]
	}

	if strings.HasPrefix(typeName, "bytes") {
		return common.Bytes2Hex(data)
	}

	return string(data)
}
