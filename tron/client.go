package tron

import (
	"context"
	"crypto/tls"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"time"
)

func NewTronProvider(url, key string) (*client.GrpcClient, error) {
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1 << 30)),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(&auth{key}),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  1 * time.Second,  // 初始重试间隔
				Multiplier: 1.6,              // 每次递增倍数
				MaxDelay:   15 * time.Second, // 最大重试间隔
			},
			MinConnectTimeout: 5 * time.Second, // 最小连接超时
		}),
	}

	grpcClient := client.NewGrpcClient(url)
	if err := grpcClient.Start(opts...); err != nil {
		return nil, err
	}
	return grpcClient, nil
}

type auth struct {
	token string
}

func (a *auth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"x-token": a.token,
	}, nil
}

func (a *auth) RequireTransportSecurity() bool {
	return false
}
