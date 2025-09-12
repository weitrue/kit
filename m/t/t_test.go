package t

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"golang.org/x/net/context"
)

func TestTimeoutProcess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("timeout")
			return // break 只会跳出select
		default:
			fmt.Println("running")
			time.Sleep(time.Second)
		}
	}
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

func TestGenerateFile(t *testing.T) {
	if len(os.Args) < 3 {
		fmt.Println("用法: file2go <输入文件> <输出文件.go>")
		return
	}

	inFile := os.Args[1]
	outFile := os.Args[2]

	// 读取输入文件
	data, err := os.ReadFile(inFile)
	if err != nil {
		panic(err)
	}

	// 获取文件名，生成变量名
	base := filepath.Base(inFile)
	varName := toVarName(base)

	// 打开输出文件
	out, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// 写 go 源码
	fmt.Fprintf(out, "package main\n\n")
	fmt.Fprintf(out, "// %s 是由 %s 生成的嵌入数据\n", varName, base)
	fmt.Fprintf(out, "var %s = []byte{\n", varName)

	// 每行最多写 12 个字节
	for _, b := range data {
		fmt.Fprintf(out, "%d,", b)
	}
	fmt.Fprint(out, "\n}\n")

	fmt.Printf("✅ 已生成 %s，变量名: %s\n", outFile, varName)
}

// 将文件名转成合法 Go 变量名
func toVarName(name string) string {
	var result []rune
	for i, r := range name {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9' && i > 0) {
			result = append(result, r)
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}
