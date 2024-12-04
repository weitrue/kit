package t

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	humanize "github.com/dustin/go-humanize"
	"github.com/shopspring/decimal"
	"github.com/weitrue/kit/evm/utils"
	"golang.org/x/net/context"
)

func TestName(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://api.dexscreener.com/latest/dex/tokens/So11111111111111111111111111111111111111112", nil)
	if err != nil {
		return
	}
	//for k, v := range headerConfig {
	//	req.Header.Set(k, v)
	//}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("http request failed:", string(body))
		return
	}

	fmt.Println(string(body))
}

func TestName1(t *testing.T) {
	hash := md5.Sum([]byte("hrbxywp@163.com"))
	fmt.Println(hex.EncodeToString(hash[:]))
}

func TestName2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("timeout")
			break
		default:
			fmt.Println("running")
			time.Sleep(time.Second)
		}
	}
}

type Message struct {
	Type string                   `json:"type"`
	Data MantaNotificationMessage `json:"data"`
}

type FBTCNotificationMessage struct {
	Type int64 `json:"type"` // 1: Risk Tx, 2: Cross-Chain Tx, 3: Mint/Redeem
	FBTCNotificationWebhook
}

type FBTCNotificationWebhook struct {
	Timestamp   string  `json:"timestamp"`
	TokenName   string  `json:"tokenName"`
	SourceChain string  `json:"sourceChain,omitempty"` // Type=2
	TargetChain string  `json:"targetChain,omitempty"` // Type=2
	Chain       string  `json:"chain,omitempty"`       // Type!=2
	Amount      float64 `json:"amount,omitempty"`      // Type!=!
	Brief       string  `json:"brief,omitempty"`       // type=1
	Details     string  `json:"details"`
	TxHash      string  `json:"txHash"`
	From        string  `json:"from"`
	To          string  `json:"to"`
}

type MantaNotificationMessage struct {
	Type int64 `json:"type"` // 1: Blocked Tx, 2: L1->L2
	MantaNotificationWebhook
}

type MantaNotificationWebhook struct {
	Timestamp     string `json:"timestamp"`
	Source        int64  `json:"source"` // 1: L2 Sequencer, 2: L1->L2
	Thread        int64  `json:"thread"`
	PotentialLoss string `json:"potentialLoss"`
	FromAddress   string `json:"fromAddress"`
	ToAddress     string `json:"toAddress"`
	TxHash        string `json:"txHash"`
}

func TestKafkaProduce(t *testing.T) {
	// 配置Kafka生产者
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Retry.Backoff = time.Second
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	// 连接Kafka broker
	brokers := []string{"localhost:9092"} // 这里是Kafka broker的地址
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Failed to start Sarama producer: %v", err)
	}
	defer producer.Close()

	// 构建消息
	topic := "example_topic"
	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder("Hello Kafka!"),
	}

	// 发送消息
	partition, offset, err := producer.SendMessage(message)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	fmt.Printf("Message sent to partition %d with offset %d\n", partition, offset)
}

func TestKafkaConsumer(t *testing.T) {
	// 创建消费者配置
	config := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092", // Kafka 服务器地址
		"group.id":          "myGroup",        // 消费者组 ID
		"auto.offset.reset": "earliest",       // 从最早的消息开始消费
	}

	// 创建消费者
	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		log.Fatalf("Failed to create consumer: %s\n", err)
	}
	defer consumer.Close()

	// 订阅主题
	topics := []string{"example_topic"} // 替换为你的主题名称
	consumer.SubscribeTopics(topics, nil)

	fmt.Println("Start consuming messages...")

	for {
		// 读取消息
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else {
			// 处理读取消息时的错误
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}

func TestName3(t *testing.T) {
	fmt.Println(Number("1e-02"))
}
func Number(word string) string {
	num, err := decimal.NewFromString(word)
	if err != nil {
		return word
	}
	number := num.InexactFloat64()

	if number < 1 {
		return fmt.Sprintf("%.4g", number)
	} else if number >= 1000 && number < math.MaxInt64 {
		return humanize.Commaf(number)
	} else if number >= math.MaxInt64 {
		return addThousandsSeparator(word)
	} else if math.Floor(number) != number {
		return fmt.Sprintf("%.2f", number)
	} else {
		return word
	}
}

func addThousandsSeparator(s string) string {
	// 检查是否是负数
	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	}

	// 查找小数点位置
	dotIndex := -1
	for i, c := range s {
		if c == '.' {
			dotIndex = i
			break
		}
	}

	// 添加千分位分隔符
	var result string
	if dotIndex == -1 {
		// 整数
		for i := len(s) - 1; i >= 0; i-- {
			result = string(s[i]) + result
			if (len(s)-i)%3 == 0 && i != 0 {
				result = "," + result
			}
		}
	} else {
		// 浮点数
		for i := dotIndex - 1; i >= 0; i-- {
			result = string(s[i]) + result
			if (dotIndex-i)%3 == 0 && i != 0 {
				result = "," + result
			}
		}
		result += s[dotIndex:]
	}

	// 添加负号
	if negative {
		result = "-" + result
	}

	return result
}

func TestSolanaDataDecode(t *testing.T) {
	now := time.Now()
	start := now.AddDate(0, 0, -7)
	fmt.Println(getIndex(start.Add(time.Minute*10), start))

	am := utils.ToDecimal(int64(17818185), 6)
	d := decimal.NewFromInt(204114280).Div(am).Round(9)
	fmt.Println(d)
}

func getIndex(cur, start time.Time) int {
	return int(cur.Sub(start).Minutes())
}
