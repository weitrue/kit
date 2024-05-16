package utils

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	BscForkClient, _  = rpc.DialContext(context.Background(), "https://rpc.phalcon.blocksec.com/rpc_25d3496386cd49d79e0055b9708af0a6")
	ETHClient, _      = rpc.DialContext(context.Background(), "https://eth-mainnet.g.alchemy.com/v2/qUt8liQq0Kh5rGm9VGGfLfSNLFuONhm3")
	BscClient, _      = rpc.DialContext(context.Background(), "https://binance.llamarpc.com")
	ArbitrumClient, _ = rpc.DialContext(context.Background(), "https://arb-mainnet.g.alchemy.com/v2/vJawUvyo1NE02qYepEe2n_AtjrX_gF5q")
	BaseClient, _     = rpc.DialContext(context.Background(), "https://base-mainnet.g.alchemy.com/v2/7XuVI7nNb8mJFfDc_HlEOiFPVNz4lKAz")
	MerlinClient, _   = rpc.DialContext(context.Background(), "https://rpc.merlinchain.io/api")
	Sender            = common.HexToAddress("0xba46dd807DD7A5bBe2eE80b6D0516A088223C574")
	EthForkClient, _  = rpc.DialContext(context.Background(), "https://rpc-dev.phalcon.blocksec.com/rpc_672bb0b921e64497a127b102e9d35db6")
	MantaClient, _    = rpc.DialContext(context.Background(), "https://pacific-rpc.manta.network/http")
)
