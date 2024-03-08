package utils

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithdraw(t *testing.T) {
	type args struct {
		ctx  context.Context
		from string
		to   string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{
				ctx:  context.Background(),
				from: "0xba46dd807DD7A5bBe2eE80b6D0516A088223C574",
				to:   "0xEF87e7024Fe8f2D35fA8Be569a3c788722b2905f",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Withdraw(tt.args.ctx, tt.args.from, tt.args.to)
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
			assert.Nil(t, err)
			transaction, err = signTransaction(transaction, "c29fc418e770e259ebd5e02e6393191898415eabffc5be64cfeaa47c1bddeeaf", big.NewInt(1))
			assert.Nil(t, err)
			bytes, err := transaction.MarshalBinary()
			assert.Nil(t, err)
			fmt.Println(hexutil.Encode(bytes))
		})
	}
}

func TestSignTransaction(t *testing.T) {
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
				sender:   "0xEF87e7024Fe8f2D35fA8Be569a3c788722b2905f",
				to:       "0xEF87e7024Fe8f2D35fA8Be569a3c788722b2905f",
				value:    big.NewInt(0),
				gas:      big.NewInt(5000000000),
				gasLimit: uint64(21000),
				input:    "0x",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction, err := CreateTransaction(tt.args.ctx, tt.args.client, tt.args.sender, tt.args.to, tt.args.value, tt.args.gas, tt.args.gasLimit, tt.args.input)
			assert.Nil(t, err)
			transaction, err = signTransaction(transaction, "14d1159998efb653b3c1c503c5e8d5904897b9c7b9f26b35bce8bab8c9d787dc", big.NewInt(4200))
			assert.Nil(t, err)
			bytes, err := transaction.MarshalBinary()
			assert.Nil(t, err)
			fmt.Println(hexutil.Encode(bytes))

			// 获取当前 gas price
			var gasPrice *big.Int
			if gasPrice, err = tt.args.client.SuggestGasPrice(tt.args.ctx); err != nil {
				log.Fatal(err)
				return
			}

			signer := types.NewEIP155Signer(transaction.ChainId())
			sender, err := signer.Sender(transaction)
			assert.Nil(t, err)
			fmt.Println(sender.String())

			baseFee, err := getBaseFee(tt.args.ctx, tt.args.client, sender, *transaction.To(), gasPrice, transaction.Data())
			assert.Nil(t, err)
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
		{
			name: "",
			args: args{
				ctx:      context.Background(),
				callData: "0xf8ab8085012a05f200827d00948de7eea34a72059324dbbed7f2f49fb2190abd5680b844a9059cbb0000000000000000000000003f8ddd0ed8798e46e2e14dfa4b7c8072805548a400000000000000000000000000000000000000000000000029a2241af62c00008220f4a0c853dabab4eacce6c34f142c9edf3ba7aded1afb9b8c82eb7cb69e7ed140ca21a007ce8a293019835d9bafb8b7ab5e7475cc8f50dac570037352c23c0667e46010",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := decodePreSignTransaction(tt.args.ctx, tt.args.callData)
			assert.Nil(t, err)
			fmt.Println(tx.Nonce())
			signer := types.LatestSignerForChainID(tx.ChainId())
			sender, err := signer.Sender(tx)
			assert.Nil(t, err)
			fmt.Println(sender.String())
		})
	}
}
