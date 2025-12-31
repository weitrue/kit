package tron

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"testing"
)

var token = "725c49edba7b926708f4fab37ba9a9d9f440365d"
var endpoint = "cold-thrilling-bird.tron-mainnet.quiknode.pro:50051"

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

func TestGetTransactionByID(t *testing.T) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(&auth{token}),
	}
	conn := client.NewGrpcClient(endpoint)
	if err := conn.Start(opts...); err != nil {
		panic(err)
	}
	defer conn.Conn.Close()

	txID := "56e2691a259d686736a5928e72cfd2bb6ad47fca3e467c3e6e59c00c06c24307"
	fmt.Println("Getting transaction with ID:", txID)

	tx, err := conn.GetTransactionByID(txID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	//getRawData(tx)
	fmt.Println("Transaction Response:", tx.GetRawData().RefBlockNum)
	txJSON, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v\n", err)
		return
	}
	fmt.Println(string(txJSON))
}

func TestGetBlockByBum(t *testing.T) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(&auth{token}),
	}

	conn := client.NewGrpcClient(endpoint)
	if err := conn.Start(opts...); err != nil {
		panic(err)
	}
	defer conn.Conn.Close()

	num := int64(78170788)

	block, err := conn.GetBlockByNum(num)
	if err != nil {
		log.Fatalf("Error getting block: %v", err)
	}

	resultJSON, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal block: %v", err)
	}
	fmt.Println(string(resultJSON))
}

type Result struct {
	Hash   string
	Status int
	Data   []byte
	Err    error
}

func requestTransaction(wg *sync.WaitGroup, task chan struct{}, ch chan<- Result, chain, hash string) {
	defer func() {
		wg.Done()
		<-task
	}()

	url := "http://127.0.0.1:9001/api/getTranscation"

	payload := fmt.Sprintf(`{
		"chain": "%s",
		"hash": "%s"
	}`, chain, hash)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		ch <- Result{Hash: hash, Err: err}
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- Result{Hash: hash, Err: err}
		return
	}
	defer resp.Body.Close()
	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- Result{Hash: hash, Err: err}
		return
	}

	ch <- Result{
		Hash:   hash,
		Data:   byts,
		Status: resp.StatusCode,
	}
}

func TestBenchCall(t *testing.T) {
	hashes := []string{
		"98f02bfa2b36493c007ee6d0ed0a7c8c14892b10a080c00dcebdff437c98ed47",
		"56e2691a259d686736a5928e72cfd2bb6ad47fca3e467c3e6e59c00c06c24307",
		"79462c7e2d1eecd7d7ffebe8f5babdf7ba59c6d2720f38fe44d289165ccf0bb7",
		"34bdeed9bbbca15dfcd9bb9e7fc840305f79818d78843dc01a26d0c2d5a80e56",
		"a4e3c363b01bf46ce7217e305d27a34ffe19c0778574069683165ebdd1194a65",
		"341fbf08d7be0995a5c826ddd2e46d8cd0c7a6c04ed6ebfed17a10decf73dd1d",
		"18165612bcd6b254d0230d08aef6e510b8d1799275414d96b4b606c555f25325",
		"2e5edfd2a3b4098c364a9a60209ca31a0bd29c1d3e2944543c342ba4d4285c53",
		"63ab03ff67c26ff08e9c706ed45377d0a285fe72423afd663b9fbe168de83da1",
		"98f02bfa2b36493c007ee6d0ed0a7c8c14892b10a080c00dcebdff437c98ed47",
		"56e2691a259d686736a5928e72cfd2bb6ad47fca3e467c3e6e59c00c06c24307",
		"79462c7e2d1eecd7d7ffebe8f5babdf7ba59c6d2720f38fe44d289165ccf0bb7",
		"34bdeed9bbbca15dfcd9bb9e7fc840305f79818d78843dc01a26d0c2d5a80e56",
		"a4e3c363b01bf46ce7217e305d27a34ffe19c0778574069683165ebdd1194a65",
		"341fbf08d7be0995a5c826ddd2e46d8cd0c7a6c04ed6ebfed17a10decf73dd1d",
		"18165612bcd6b254d0230d08aef6e510b8d1799275414d96b4b606c555f25325",
		"2e5edfd2a3b4098c364a9a60209ca31a0bd29c1d3e2944543c342ba4d4285c53",
		"63ab03ff67c26ff08e9c706ed45377d0a285fe72423afd663b9fbe168de83da1",
		"98f02bfa2b36493c007ee6d0ed0a7c8c14892b10a080c00dcebdff437c98ed47",
		"56e2691a259d686736a5928e72cfd2bb6ad47fca3e467c3e6e59c00c06c24307",
		"79462c7e2d1eecd7d7ffebe8f5babdf7ba59c6d2720f38fe44d289165ccf0bb7",
		"34bdeed9bbbca15dfcd9bb9e7fc840305f79818d78843dc01a26d0c2d5a80e56",
		"a4e3c363b01bf46ce7217e305d27a34ffe19c0778574069683165ebdd1194a65",
		"341fbf08d7be0995a5c826ddd2e46d8cd0c7a6c04ed6ebfed17a10decf73dd1d",
		"18165612bcd6b254d0230d08aef6e510b8d1799275414d96b4b606c555f25325",
		"2e5edfd2a3b4098c364a9a60209ca31a0bd29c1d3e2944543c342ba4d4285c53",
		"63ab03ff67c26ff08e9c706ed45377d0a285fe72423afd663b9fbe168de83da1",
		"98f02bfa2b36493c007ee6d0ed0a7c8c14892b10a080c00dcebdff437c98ed47",
		"56e2691a259d686736a5928e72cfd2bb6ad47fca3e467c3e6e59c00c06c24307",
		"79462c7e2d1eecd7d7ffebe8f5babdf7ba59c6d2720f38fe44d289165ccf0bb7",
		"34bdeed9bbbca15dfcd9bb9e7fc840305f79818d78843dc01a26d0c2d5a80e56",
		"a4e3c363b01bf46ce7217e305d27a34ffe19c0778574069683165ebdd1194a65",
		"341fbf08d7be0995a5c826ddd2e46d8cd0c7a6c04ed6ebfed17a10decf73dd1d",
		"18165612bcd6b254d0230d08aef6e510b8d1799275414d96b4b606c555f25325",
		"2e5edfd2a3b4098c364a9a60209ca31a0bd29c1d3e2944543c342ba4d4285c53",
		"63ab03ff67c26ff08e9c706ed45377d0a285fe72423afd663b9fbe168de83da1",
		"98f02bfa2b36493c007ee6d0ed0a7c8c14892b10a080c00dcebdff437c98ed47",
		"56e2691a259d686736a5928e72cfd2bb6ad47fca3e467c3e6e59c00c06c24307",
		"79462c7e2d1eecd7d7ffebe8f5babdf7ba59c6d2720f38fe44d289165ccf0bb7",
		"34bdeed9bbbca15dfcd9bb9e7fc840305f79818d78843dc01a26d0c2d5a80e56",
		"a4e3c363b01bf46ce7217e305d27a34ffe19c0778574069683165ebdd1194a65",
		"341fbf08d7be0995a5c826ddd2e46d8cd0c7a6c04ed6ebfed17a10decf73dd1d",
		"18165612bcd6b254d0230d08aef6e510b8d1799275414d96b4b606c555f25325",
		"2e5edfd2a3b4098c364a9a60209ca31a0bd29c1d3e2944543c342ba4d4285c53",
		"63ab03ff67c26ff08e9c706ed45377d0a285fe72423afd663b9fbe168de83da1",
	}

	resultCh := make(chan Result, len(hashes)*5)
	var wg sync.WaitGroup
	task := make(chan struct{}, 15)

	for i := 0; i < 5; i++ {
		wg.Add(len(hashes))
		for _, h := range hashes {
			task <- struct{}{}
			go requestTransaction(&wg, task, resultCh, "tron", h)
		}
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for res := range resultCh {
		if res.Err != nil {
			fmt.Println("error:", res.Hash, res.Err)
			continue
		}
		fmt.Println("ok:", res.Hash, res.Status, len(res.Data))
	}
}
