package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/crypto/sha3"
	"strings"
)

const (
	eip165SupportsInterfaceABI = ""
	nft721InterfaceId          = ""
)

var (
	ErrNoData = errors.New("no data")
)

// IsERC721 supportsInterface(0x80ac58cd)是否返回true
func IsERC721() {

}

func CalculateInterfaceId(contractABI string) []byte {
	// 计算函数签名的 keccak256 哈希值
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(contractABI))
	interfaceId := hash.Sum(nil)[:4] // 只取前4个字节作为 interfaceId
	return interfaceId
}

func calculateSelector(selector string) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(selector))
	return hash.Sum(nil)[:4]
}

func calculateMethodSelector(m abi.Method) []byte {
	//function foo(uint32 a, int b)    =    "foo(uint32,int256)"
	types := make([]string, len(m.Inputs))
	for i, v := range m.Inputs {
		types[i] = v.Type.String()
	}
	functionStr := fmt.Sprintf("%v(%v)", m.Name, strings.Join(types, ","))
	keccak := sha3.NewLegacyKeccak256()
	keccak.Write([]byte(functionStr))
	return keccak.Sum(nil)[:4]
}

func DecodeABI(abiStr string) (*abi.ABI, error) {
	contractAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, err
	}

	return &contractAbi, nil
}

func Call(ctx context.Context, c *rpc.Client, contract string, method abi.Method) (any, error) {
	to := common.HexToAddress(contract)
	data := method.ID
	callData := &ethereum.CallMsg{
		To:   &to,
		Data: data,
	}

	client := ethclient.NewClient(c)
	result, err := client.CallContract(ctx, *callData, nil)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, ErrNoData
	}

	out, err := method.Outputs.Unpack(result)
	if err != nil {
		return nil, err
	}

	return out[0], nil
}

func CallWithInput(ctx context.Context, c *rpc.Client, contract, methodName string, abiO *abi.ABI, args ...any) ([]any, error) {
	to := common.HexToAddress(contract)
	if len(args) == 0 {
		return nil, errors.New("invalid input")
	}

	method := abiO.Methods[methodName]
	data, err := abiO.Pack(methodName, args...)
	if err != nil {
		return nil, err
	}

	callData := &ethereum.CallMsg{
		To:   &to,
		Data: data,
	}

	client := ethclient.NewClient(c)
	result, err := client.CallContract(ctx, *callData, nil)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, ErrNoData
	}

	out, err := method.Outputs.Unpack(result)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func CodeAt(ctx context.Context, c *rpc.Client, contract string) (string, error) {
	client := ethclient.NewClient(c)
	code, err := client.CodeAt(ctx, common.HexToAddress(contract), nil)
	if err != nil {
		return "", err
	}

	fmt.Println(common.Bytes2Hex(code))

	return calcCodeHash(code), nil
}

func calcCodeHash(code []byte) string {
	if len(code) == 0 {
		return "0x"
	}

	return crypto.Keccak256Hash(code).Hex()
}
