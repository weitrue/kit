package utils

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/keycard-go/hexutils"
	"log"
	"math/big"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestWithdraw(t *testing.T) {
	type args struct {
		ctx        context.Context
		c          *rpc.Client
		from       string
		to         string
		privateKey string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				ctx:        context.Background(),
				c:          ETHClient,
				from:       "0xba46dd807DD7A5bBe2eE80b6D0516A088223C574",
				to:         "0xEF87e7024Fe8f2D35fA8Be569a3c788722b2905f",
				privateKey: "c29fc418e770e259ebd5e02e6393191898415eabffc5be64cfeaa47c1bddeeaf",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Withdraw(tt.args.ctx, tt.args.c, tt.args.from, tt.args.to, tt.args.privateKey)
			assert.Nil(t, err)
		})
	}

}

func TestCreateTransaction(t *testing.T) {
	type args struct {
		ctx      context.Context
		client   *ethclient.Client
		sender   string
		to       string
		value    *big.Int
		gas      *big.Int
		gasLimit uint64
		input    string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.Transaction
		wantErr error
	}{
		{
			name: "",
			args: args{
				ctx:      context.Background(),
				client:   ethclient.NewClient(ETHClient),
				sender:   "0xba46dd807DD7A5bBe2eE80b6D0516A088223C574",
				to:       "0xffc4baf49d1abfc8e1feb86108321f4720298689",
				value:    big.NewInt(0),
				gas:      big.NewInt(2000000000),
				gasLimit: uint64(21000),
				input:    "0x6a7612020000000000000000000000009ba67b0b391d615b3f2fd9ba3c3d9d30c2365c3500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000140000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000026000000000000000000000000000000000000000000000000000000000000000e42a3b93e500000000000000000000000089c735df172a35e070dd300a61bde7443ec32ec4fa7c58710000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000200000000000000000000000021e27a5e5513d6e65c4f830167390997aa84843a000000000000000000000000fbeffffa7bf68bdd2eff3346da00ec2a74a4f5870000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000417456bd8d35ecaf63ca6a8cba6d2c1e8ffaa01788f7730bbf642905c032f7ccfe0f80f86b886ad56f4f892eda773573303847cb1bdc9f9aebfe594392c8f440f31b00000000000000000000000000000000000000000000000000000000000000",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction, err := CreateTransaction(tt.args.ctx, tt.args.client, tt.args.sender, tt.args.to, tt.args.value, tt.args.gas, tt.args.gasLimit, tt.args.input)
			if !assert.Nil(t, err) {
				t.Log(err)
				return
			}
			transaction, err = SignTransaction(transaction, "c29fc418e770e259ebd5e02e6393191898415eabffc5be64cfeaa47c1bddeeaf", big.NewInt(1))
			if !assert.Nil(t, err) {
				t.Log(err)
				return
			}
			bytes, err := transaction.MarshalBinary()
			if !assert.Nil(t, err) {
				t.Log(err)
				return
			}
			fmt.Println(hexutil.Encode(bytes))
		})
	}
}

func TestSignTransaction(t *testing.T) {
	type args struct {
		ctx        context.Context
		client     *ethclient.Client
		sender     string
		to         string
		value      *big.Int
		gas        *big.Int
		gasLimit   uint64
		input      string
		chainId    *big.Int
		privateKey string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.Transaction
		wantErr error
	}{
		{
			name: "ETH",
			args: args{
				ctx:        context.Background(),
				client:     ethclient.NewClient(ETHClient),
				chainId:    big.NewInt(1),
				sender:     "0x6b65dD3537AF63A285e0a008ecbcbE725ccB8fd2",
				to:         "0x6b65dD3537AF63A285e0a008ecbcbE725ccB8fd2",
				value:      big.NewInt(0),
				gas:        big.NewInt(6000000000),
				gasLimit:   uint64(21000),
				input:      "0x",
				privateKey: "b79bffb5b9e2303a9bdd5b0a5638f81705f5f813f5fa7c219257a3b0cffca49d",
			},
		},
		{
			name: "Manta",
			args: args{
				ctx:        context.Background(),
				client:     ethclient.NewClient(ETHClient),
				chainId:    big.NewInt(169),
				sender:     "0x6b65dD3537AF63A285e0a008ecbcbE725ccB8fd2",
				to:         "0x6b65dD3537AF63A285e0a008ecbcbE725ccB8fd2",
				value:      big.NewInt(0),
				gas:        big.NewInt(6000000000),
				gasLimit:   uint64(21000),
				input:      "0x",
				privateKey: "b79bffb5b9e2303a9bdd5b0a5638f81705f5f813f5fa7c219257a3b0cffca49d",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction, err := CreateTransaction(tt.args.ctx, tt.args.client, tt.args.sender, tt.args.to, tt.args.value, tt.args.gas, tt.args.gasLimit, tt.args.input)
			if !assert.Nil(t, err) {
				t.Log(err)
				return
			}
			transaction, err = SignTransaction(transaction, tt.args.privateKey, tt.args.chainId)
			if !assert.Nil(t, err) {
				t.Log(err)
				return
			}
			bytes, err := transaction.MarshalBinary()
			if !assert.Nil(t, err) {
				t.Log(err)
				return
			}
			fmt.Println("pre sign")
			fmt.Println(hexutil.Encode(bytes))

			// 获取当前 gas price
			var gasPrice *big.Int
			if gasPrice, err = tt.args.client.SuggestGasPrice(tt.args.ctx); err != nil {
				log.Fatal(err)
				return
			}

			signer := types.NewLondonSigner(transaction.ChainId())
			sender, err := signer.Sender(transaction)
			if !assert.Nil(t, err) {
				t.Log(err)
				return
			}
			fmt.Println("sender")
			fmt.Println(sender.String())
			baseFee, err := getBaseFee(tt.args.ctx, tt.args.client, sender, *transaction.To(), gasPrice, transaction.Data())
			if !assert.Nil(t, err) {
				t.Log(err)
				return
			}
			fmt.Println("Fee")
			fmt.Println(ToDecimal(baseFee, 18).String())
			fmt.Println(ToDecimal(new(big.Int).Mul(new(big.Int).SetUint64(transaction.Gas()), transaction.GasPrice()), 18).String())
		})
	}
}

func Test_decodeTransactionByPreSign(t *testing.T) {
	type args struct {
		ctx      context.Context
		callData string
	}
	tests := []struct {
		name string
		args args
	}{
		//{
		//	name: "",
		//	args: args{
		//		ctx:      context.Background(),
		//		callData: "0xf8650484b2d05e00825208943d497994fff1d1ace609b83b8ac440b5d6f04cf603808220f3a0f987a52e0241e525e3e52ade0553cc9232c7e3033225ba93b3eced81df6e62e5a058e5b5bccf724c1622f87c5e7e8e499cd63b970c1ddbb7ce5f6bed67679b3aec",
		//	},
		//},
		//{
		//	name: "manta",
		//	args: args{
		//		ctx:      context.Background(),
		//		callData: "0xf866808501de9027e8825208946b65dd3537af63a285e0a008ecbcbe725ccb8fd28080820175a080b8076c2d40b89983c232ac00565d8f3c79e25049bca6ddb6ceddd1ead4567fa0798669d54d8bc0d5810dde0c4ba756bf12f94bc396d534c8335ecc619387c759",
		//	},
		//},
		{
			name: "bsc",
			args: args{
				ctx:      context.Background(),
				callData: "0x410cc3a300000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000002400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ae50fbca863fc28d9b5833e38ddb040d507583fd00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000014401d5062a00000000000000000000000095bdb13cc363998f3adec92985bab00e1b62531a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004418c882a50000000000000000000000007089324d008db4a49a676c224f6aec5b15a5e8d7000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ae50fbca863fc28d9b5833e38ddb040d507583fd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000124134008d300000000000000000000000095bdb13cc363998f3adec92985bab00e1b62531a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004418c882a50000000000000000000000007089324d008db4a49a676c224f6aec5b15a5e8d700000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := decodePreSignTransaction(tt.args.callData)
			if !assert.Nil(t, err) {
				t.Log(err)
				return
			}
			fmt.Println(tx.Nonce(), tx.ChainId().String())
			signer := types.LatestSignerForChainID(tx.ChainId())
			sender, err := signer.Sender(tx)
			if !assert.Nil(t, err) {
				t.Log(err)
				return
			}
			fmt.Println(sender.String())
		})
	}
}

func TestName1(t *testing.T) {
	str := "b530f66688e58ad957574aa09e98cf209b40b782562ee3a9aaaa685f34f5be632618f4c64805526a3092d41f25597ccfe4dd82166644607b"
	decode, err := hex.DecodeString(str)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for len(decode) > 0 {
		// 解码第一个 UTF-8 编码字符
		r, size := utf8.DecodeRune(decode)
		if r == utf8.RuneError && size == 1 {
			// 如果遇到无效的 UTF-8 编码字符，则返回 false
			fmt.Println("Invalid UTF-8 encoding")
			return
		}
		// 跳过已解码的字节
		decode = decode[size:]
	}
	fmt.Println(utf8.Valid(decode))

	fmt.Println(string(decode))
	byt := hexutils.HexToBytes(str)
	fmt.Println(string(byt))
}
