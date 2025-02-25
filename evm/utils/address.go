package utils

import (
	"context"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shopspring/decimal"
	"reflect"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

const (
	zeroAddress         = "0x0000000000000000000000000000000000000000"
	NativeTokenDecimals = 18
)

// PublicKeyBytesToAddress zero address return if key is invalid
func PublicKeyBytesToAddress(publicKey []byte) common.Address {
	if len(publicKey) < 12 {
		return common.HexToAddress(zeroAddress)
	}

	var buf []byte

	hash := sha3.NewLegacyKeccak256()
	hash.Write(publicKey[1:]) // remove EC prefix 04
	buf = hash.Sum(nil)
	address := buf[12:]

	return common.HexToAddress(hex.EncodeToString(address))
}

// IsValidAddress validate hex address
func IsValidAddress(addr any) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	switch v := addr.(type) {
	case string:
		return re.MatchString(v)
	case common.Address:
		return re.MatchString(v.Hex())
	default:
		return false
	}
}

// IsZeroAddress validate if it's a 0 address
func IsZeroAddress(addr any) bool {
	var address common.Address
	switch v := addr.(type) {
	case string:
		address = common.HexToAddress(v)
	case common.Address:
		address = v
	default:
		return false
	}

	zeroAddressBytes := common.FromHex(zeroAddress)
	addressBytes := address.Bytes()
	return reflect.DeepEqual(addressBytes, zeroAddressBytes)
}

func GetBalance(ctx context.Context, c *rpc.Client, address string) (string, error) {
	client := ethclient.NewClient(c)
	balanceAt, err := client.BalanceAt(ctx, common.HexToAddress(address), nil)
	if err != nil {
		return "", err
	}

	return decimal.NewFromBigInt(balanceAt, -NativeTokenDecimals).String(), nil
}

func GetBalanceWithDecimals(ctx context.Context, c *rpc.Client, address string) (decimal.Decimal, error) {
	client := ethclient.NewClient(c)
	balanceAt, err := client.BalanceAt(ctx, common.HexToAddress(address), nil)
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromBigInt(balanceAt, 0), nil
}
