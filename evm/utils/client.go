package utils

import (
	"context"

	"github.com/ethereum/go-ethereum/rpc"
)

var (
	ETHClient, _      = rpc.DialContext(context.Background(), "https://eth-mainnet.g.alchemy.com/v2/qUt8liQq0Kh5rGm9VGGfLfSNLFuONhm3")
	BscClient, _      = rpc.DialContext(context.Background(), "https://binance.llamarpc.com")
	ArbitrumClient, _ = rpc.DialContext(context.Background(), "https://arb-mainnet.g.alchemy.com/v2/vJawUvyo1NE02qYepEe2n_AtjrX_gF5q")
	BaseClient, _     = rpc.DialContext(context.Background(), "https://base-mainnet.g.alchemy.com/v2/7XuVI7nNb8mJFfDc_HlEOiFPVNz4lKAz")
	MerlinClient, _   = rpc.DialContext(context.Background(), "https://rpc.merlinchain.io/api")
)
