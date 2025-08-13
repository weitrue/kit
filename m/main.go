package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/otiai10/gosseract/v2"
	"image"
	"log"
	"strings"
	"time"
)

/*
#cgo CFLAGS: -I/opt/homebrew/include
#cgo LDFLAGS: -L/opt/homebrew/lib -llept
#include <leptonica/allheaders.h>
*/
import "C"

func main() {
	println("CGO works!")
}

func main2() {
	// 示例 base64 图片数据（请替换为你自己的）
	base64Image := "data:image/png;base64,..."

	// 移除前缀（如果有）
	if strings.Contains(base64Image, ",") {
		base64Image = strings.Split(base64Image, ",")[1]
	}

	// 解码 base64
	imgData, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		panic(err)
	}

	// 解码为 image.Image 对象（可选）
	img, format, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		panic(err)
	}
	fmt.Println("图片格式：", format)
	_ = img // 你可以用它做进一步处理，比如灰度化、缩放等

	// OCR 识别
	client := gosseract.NewClient()
	defer client.Close()

	err = client.SetImageFromBytes(imgData)
	if err != nil {
		panic(err)
	}

	// 可选：限制只识别数字
	client.SetWhitelist("0123456789")

	text, err := client.Text()
	if err != nil {
		panic(err)
	}

	fmt.Println("识别结果：", strings.TrimSpace(text))
}

func main1() {
	url := "ws://localhost:9000/trade/solana?apiKey=6959765ca5b94b78ae95cf5f9d0cdb86"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	//defer conn.Close()

	// 构造要发送的 binary 数据
	message := map[string]any{
		"type":      "SUBSCRIBE_SWAP",
		"accounts":  []string{"5peFv5HvpMRGPUpkujeWExq7mwdfENbDgtnpsG9h6vFc", "jtnEAnbJzraBjxM9eKCbhJaUgxhiF2Kxu8hS4AHXzTb", "C6GA4fZDTrxzqRh16F1XbGvYqodmgWT1JHCKHvKVnv8j", "JDd3hy3gQn2V982mi1zqhNqUw1GfV2UL6g76STojCJPN", "12BRrNxzJYMx7cRhuBdhA71AchuxWRcvGydNnDoZpump"},
		"protocols": []string{"jupiter_v6", "raydium_clmm", "pumpfun", "pumpfun_amm", "raydium_amm"},
	}

	byt, _ := json.Marshal(message)

	// 发送 binary message
	err = conn.WriteMessage(websocket.TextMessage, byt)
	if err != nil {
		log.Fatal("write error:", err)
	}

	for {
		// 读取服务器回传的信息
		_, resp, err := conn.ReadMessage()
		if err != nil {
			log.Fatal("read error:", err)
		}
		log.Printf("received: %s", string(resp))

		time.Sleep(time.Second)
	}
}
