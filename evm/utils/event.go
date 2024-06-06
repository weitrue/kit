package utils

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

func SubscribeContractEvents(ctx context.Context, client *ethclient.Client, contractAddress common.Address, fromBlock, toBlock *big.Int, topics []string) error {
	topic := make([]common.Hash, 0)
	for _, v := range topics {
		topic = append(topic, common.HexToHash(v))
	}
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Topics:    [][]common.Hash{topic},
	}

	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		return err
	}

	for _, v := range logs {
		fmt.Println(common.Bytes2Hex(v.Data[:32]))
	}

	return nil
}
