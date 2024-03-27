package utils

import (
	"context"
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

func DecodeABI(abiStr string) (*abi.ABI, error) {
	contractAbi, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, err
	}

	return &contractAbi, nil
}

func Call(ctx context.Context, c *rpc.Client, contract string, method abi.Method, args ...any) (any, error) {
	to := common.HexToAddress(contract)
	data := method.ID
	if len(args) > 0 {
		input, err := method.Inputs.Pack(args)
		if err != nil {
			return nil, err
		}
		data = append(data, input...)
	} else {
		input, err := method.Inputs.Pack()
		if err != nil {
			return nil, err
		}
		data = append(data, input...)
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

	out, err := method.Outputs.Unpack(result)
	if err != nil {
		return nil, err
	}

	return out[0], nil
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
