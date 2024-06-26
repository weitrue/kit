package utils

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestSubscribeContractEvents(t *testing.T) {
	type args struct {
		ctx             context.Context
		client          *ethclient.Client
		contractAddress common.Address
		fromBlock       *big.Int
		toBlock         *big.Int
		topics          []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		//{
		//	name: "add module",
		//	args: args{
		//		ctx:             context.Background(),
		//		client:          ethclient.NewClient(BscClient),
		//		contractAddress: common.HexToAddress("0x7CC71473bbc02b6d234f8d4a92D0E7572316B741"),
		//		fromBlock:       big.NewInt(34487152),
		//		toBlock:         big.NewInt(34487160),
		//		topics:          []string{"0x5983cdcaa370320b76fe01a3a32a0430e6a13b9f47a55e806afb13b5aef95a12"},
		//	},
		//	wantErr: nil,
		//},
		//{
		//	name: "delete module",
		//	args: args{
		//		ctx:             context.Background(),
		//		client:          ethclient.NewClient(BscClient),
		//		contractAddress: common.HexToAddress("0xA281c299f9E85d15F2C743C727ee34551E13d37E"),
		//		fromBlock:       big.NewInt(34339714),
		//		toBlock:         big.NewInt(34339715),
		//		topics:          []string{"0xaab4fa2b463f581b2b32cb3b7e3b704b9ce37cc209b5fb4d77e593ace4054276"},
		//	},
		//	wantErr: nil,
		//},
		{
			name: "update module",
			args: args{
				ctx:             context.Background(),
				client:          ethclient.NewClient(BscClient),
				contractAddress: common.HexToAddress("0xA281c299f9E85d15F2C743C727ee34551E13d37E"),
				fromBlock:       big.NewInt(34365090),
				toBlock:         big.NewInt(34365090),
				topics:          []string{"0x442e715f626346e8c54381002da614f62bee8d27386535b2521ec8540898556e"},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SubscribeContractEvents(tt.args.ctx, tt.args.client, tt.args.contractAddress, tt.args.fromBlock, tt.args.toBlock, tt.args.topics)
			assert.Nil(t, err)
		})
	}
}
