package utils

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	ETHClient, _ = rpc.DialContext(context.Background(), "https://rpc.phalcon.blocksec.com/rpc_25d3496386cd49d79e0055b9708af0a6")
	Sender       = common.HexToAddress("0xba46dd807DD7A5bBe2eE80b6D0516A088223C574")
)
