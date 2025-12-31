package t

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
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

var html = `<!doctype html>
<html>
<head>
<meta charset="utf-8"/>
<style>
  @page { size: A4; margin: 20mm; }
  body { font-family: Arial, sans-serif; }

  /* 普通内容 */
  .content { width: 100%; }
  .section { margin-bottom: 16px; }

  /* 强制最后一页单独成页 */
  .last-page {
    /* 这些指令尽量都加，兼容不同浏览器/引擎 */
    page-break-before: always;
    break-before: page;
    -webkit-region-break-before: always;
    page-break-inside: avoid;
    break-inside: avoid;
    /* 让这个元素撑满整页（可选） */
    display: block;
  }

  /* 打印专用隐藏/显示 */
  @media print {
    .only-screen { display: none; }
  }
</style>
</head>
<body>
  <div class="content">
    <h1>Main Document</h1>
    <div class="section">Lots of content... (simulate pages)</div>
    <div class="section">More content...</div>
    <div class="section">Even more content...</div>
    <!-- repeat to make multiple pages for demo -->

    <!-- 这是被强制为单独最后一页的内容 -->
    <div class="last-page">
      <h2>Appendix</h2>
      <p>This content will start on a new page because of page-break-before: always.</p>
    </div>
 	<!-- 这是被强制为单独最后一页的内容 -->
    <div class="last-page">
      <h2>Appendix / Last Page</h2>
      <p>This content will start on a new page because of page-break-before: always.</p>
    </div>
  </div>

  <!-- 当渲染完成后设置window.isRenderDone=true，这样可以用chromedp等待 -->
  <script>
    // 如果页面是动态加载数据的话，实际项目里需要在数据渲染完时显式设置这个变量。
    window.isRenderDone = true;
  </script>
</body>
</html>
`

func TestPDF(t *testing.T) {
	// 1. 创建上下文
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// 设置超时时间，避免死等
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 以 data URL 导航（也可以把 HTML 写到临时文件然后 file://）
	dataURL := "data:text/html," + url.PathEscape(html)

	// 要加载的动态页面
	//url := "https://app.blocksec.com/explorer/tx/eth/0x6c3cf48e9c8b22f60fc2b702b03f93926ee8624639b100fe851b757947e61759"

	var pdfBuf []byte
	var readyState string

	// 2. 执行任务
	err := chromedp.Run(ctx,
		chromedp.Navigate(dataURL),

		chromedp.Evaluate(`document.readyState`, &readyState),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 等待直到 readyState == "complete"
			for readyState != "complete" {
				time.Sleep(100 * time.Millisecond)
				if err := chromedp.Evaluate(`document.readyState`, &readyState).Do(ctx); err != nil {
					return err
				}
			}

			time.Sleep(200 * time.Millisecond)
			return nil
		}),

		// 等待页面加载完成，可以根据具体页面调整
		chromedp.WaitReady("body", chromedp.ByQuery),
		// 打印为 PDF
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			now := time.Now().UTC().Format(time.DateTime)
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithDisplayHeaderFooter(true).
				WithPreferCSSPageSize(true).
				WithHeaderTemplate(`
				<div style="padding-left: 36px; padding-right: 36px; width: 100vw;">
				<div style="display: flex; align-items: center; justify-content: space-between; width: 100%; border-bottom: 1px solid #e3e8ef; padding-bottom: 16px;">
                <div>
                    <svg xmlns="http://www.w3.org/2000/svg" width="68" height="20" viewBox="0 0 68 20" fill="none">
                      <path d="M27.8048 15.2654V18.3387H26.4189V10.9962H29.1888C30.7631 10.9962 32.0646 11.3312 32.0646 13.0938C32.0646 14.8564 30.7749 15.2751 29.0729 15.2751C28.7981 15.2751 28.3427 15.2751 27.8048 15.2654ZM28.9669 14.1221C29.8758 14.1221 30.7121 14.0481 30.7121 13.0938C30.7121 12.1395 29.9406 12.1609 28.9669 12.1609H27.8048V14.1221H28.9669Z" fill="#121926"></path>
                      <path d="M36.1673 12.936C37.2451 12.936 38.0813 13.5437 38.0813 14.7395V18.3367H36.7818V15.6627C36.7818 14.6558 36.4127 14.2039 35.6393 14.2039C34.7932 14.2039 34.0747 14.6558 34.0747 15.6627V18.3367H32.785V10.7216H34.0747V13.2398C34.0747 13.6488 33.8627 14.1416 33.7155 14.5292L33.9373 14.6129C34.0845 14.2253 34.222 13.68 34.54 13.4171C34.9836 13.0392 35.4803 12.9341 36.1673 12.9341V12.936Z" fill="#121926"></path>
                      <path d="M41.6011 18.5062C40.0268 18.5062 38.9157 17.248 38.9157 15.6841C38.9157 14.1202 40.0248 12.936 41.6011 12.936C42.3098 12.936 42.8594 13.1775 43.2403 13.5865C43.4621 13.8163 43.5053 14.2468 43.6113 14.5506L43.8331 14.4766C43.7389 14.1728 43.5485 13.8047 43.5485 13.4911L43.5387 13.0918H44.8382V18.3367H43.5387L43.5485 17.9063C43.5485 17.5908 43.7389 17.2149 43.8331 16.9208L43.6015 16.8468C43.5072 17.1409 43.4641 17.5713 43.252 17.8109C42.8614 18.2413 42.3118 18.5023 41.6031 18.5023L41.6011 18.5062ZM41.6757 17.3104C42.5434 17.3104 43.1244 16.5449 43.1559 15.6841C43.1559 14.7609 42.6376 14.1319 41.6757 14.1319C40.7982 14.1319 40.2172 14.7921 40.2172 15.6841C40.2172 16.5761 40.7982 17.3104 41.6757 17.3104Z" fill="#121926"></path>
                      <path
                        d="M47.4058 10.7333V16.5761C47.4058 17.0065 47.5864 17.1429 48.0713 17.1429H48.727V18.3387H47.9437C46.7384 18.3387 46.1043 17.8868 46.1043 17.0689V10.7333H47.4058Z"
                        fill="#121926"></path>
                      <path
                        d="M52.3213 12.936C54.0664 12.936 54.9969 14.0052 55.1009 14.981H53.715C53.6836 14.8758 53.3126 14.1416 52.2879 14.1416C51.3358 14.1416 50.7548 14.7298 50.745 15.6724C50.745 16.6053 51.3574 17.2772 52.3311 17.287C53.3362 17.287 53.717 16.5527 53.7484 16.4378H55.1324C54.9635 17.4662 54.0645 18.4945 52.3723 18.5042C50.6272 18.4945 49.465 17.3922 49.465 15.6821C49.465 13.9721 50.6075 12.9341 52.3193 12.9341L52.3213 12.936Z"
                        fill="#121926"></path>
                      <path
                        d="M58.8347 12.936C60.6643 12.936 61.8578 14.089 61.8578 15.7055C61.8578 17.322 60.6623 18.5062 58.8347 18.5062C57.0071 18.5062 55.8214 17.3941 55.8214 15.7055C55.8214 14.0169 57.0169 12.936 58.8347 12.936ZM58.8347 17.2675C60.0086 17.2675 60.4837 16.5644 60.4837 15.7055C60.4837 14.8466 60.0086 14.1747 58.8347 14.1747C57.6608 14.1747 57.1956 14.8466 57.1956 15.7055C57.1956 16.5644 57.6706 17.2675 58.8347 17.2675Z"
                        fill="#121926"></path>
                      <path
                        d="M66.086 12.936C67.1637 12.936 68 13.5437 68 14.7395V18.3367H66.7005V15.6627C66.7005 14.6558 66.3314 14.2039 65.558 14.2039C64.7119 14.2039 63.9934 14.6558 63.9934 15.6627V18.3367H62.7037V13.0918H63.9934V13.1756C63.9934 13.5943 63.7814 14.0987 63.6244 14.498L63.8462 14.5915C64.0052 14.1922 64.1524 13.6274 64.4901 13.3645C64.9337 13.0178 65.399 12.9341 66.086 12.9341V12.936Z"
                        fill="#121926"></path>
                      <path
                        d="M6.38962 7.46907C8.78257 6.33166 10.3668 5.78828 11.6388 5.35202C13.435 4.73657 14.6187 4.32953 17.0156 2.60589C15.8613 2.08588 14.5795 1.79763 13.2308 1.79763C8.9514 1.79763 5.35706 4.70347 4.34413 8.6357C4.98801 8.20138 5.67114 7.81185 6.39158 7.46907H6.38962Z"
                        fill="#121926"></path>
                      <path
                        d="M20.981 12.7218C19.7482 14.8174 18.1444 16.3599 16.2148 17.3084C15.2509 17.7836 14.3204 18.1186 13.3742 18.3348C12.5399 18.5257 11.7115 18.6172 10.8419 18.6172C9.7897 18.6172 8.84744 18.4828 7.94247 18.3367C9.43635 19.3846 11.262 20 13.2309 20C18.2976 20 22.4042 15.9256 22.4042 10.8988C22.4042 10.5249 22.3807 10.1587 22.3375 9.79646C21.9527 10.8638 21.5012 11.8415 20.983 12.7198L20.981 12.7218Z"
                        fill="#121926"></path>
                      <path
                        d="M22.526 0C22.5005 0.0292142 22.4749 0.0564812 22.4494 0.0856953C21.5484 1.08871 20.5531 2.00604 19.4852 2.82988C13.9278 7.12047 13.5529 5.84478 7.31432 8.81099C0.645844 11.9817 0 18.2939 0 18.2939C6.20716 13.7501 9.00451 19.1781 15.7652 15.8555C23.001 12.2992 22.526 0 22.526 0ZM12.8462 15.0317C12.8462 15.0317 10.4179 15.6101 7.25346 14.5448C7.25346 14.5448 5.90092 14.1182 4.82517 14.1182C4.82517 14.1182 4.98614 12.1472 9.62286 10.0263C9.6209 10.075 9.61893 10.1217 9.61893 10.1704C9.61893 11.9311 11.0559 13.3567 12.8305 13.3567C14.4716 13.3567 15.8241 12.1356 16.0185 10.5619C15.3314 10.8599 14.5305 10.5775 14.2125 9.91528C14.0338 9.54329 14.0436 9.12845 14.2046 8.77593C14.3361 8.48769 14.5697 8.24034 14.8838 8.09232C14.9309 8.06895 14.98 8.04947 15.0271 8.03389C15.2784 7.94625 15.5238 7.83913 15.7633 7.72422C16.5622 7.34249 17.3317 6.89454 18.0757 6.41738C18.967 5.84673 19.8778 5.16507 20.559 4.42303C20.559 4.42303 19.3301 13.8436 12.8462 15.0336V15.0317Z"
                        fill="#121926"></path>
                      <path
                        d="M34.0924 7.80798V8.44289H31.3127L30.6728 7.80798V5.28777H31.3127V7.80798H34.0924Z"
                        fill="#121926"></path>
                      <path
                        d="M46.6323 6.34921L46.1415 6.83611L46.6323 7.32108V8.44289H45.9923V7.52946L45.6115 7.15164H43.6269V8.44289H42.9869V5.28777H43.6269V6.51671H45.6115L45.9923 6.14083V5.28777H46.6323V6.34921Z"
                        fill="#121926"></path>
                      <path
                        d="M52.2565 5.92461V6.51669H54.9714V7.15161H52.2565V7.74563H54.9714V8.38055H52.2565V8.38249H52.1937L51.6165 7.80796V5.28968H54.9714V5.92461H52.2565Z"
                        fill="#121926"></path>
                      <path
                        d="M58.917 7.80799V7.81382L58.2829 8.4429H56.1413L55.5013 7.80799V5.85648L56.1413 5.22155H58.279L58.917 5.85648H56.1413V7.80799H58.917Z"
                        fill="#121926"></path>
                      <path
                        d="M50.9766 7.80796L50.9785 7.8878L50.419 8.44287H47.231V7.80796H50.3994V7.15161H47.7807L47.231 6.60822V5.85645L47.8082 5.28968H50.9785L50.9766 5.92461H47.871V6.51669H50.5447L50.9766 6.95295V7.80796Z"
                        fill="#121926"></path>
                      <path
                        d="M37.6769 5.28777H35.1819L34.5419 5.85647V7.80798L35.2486 8.44289H37.6769L38.3169 7.80798V5.85647L37.6769 5.28777ZM37.6769 7.80798H35.1819V5.92268H37.6769V7.80798Z"
                        fill="#121926"></path>
                      <path
                        d="M30.0562 5.85648L29.4163 5.22156H26.2813V8.38253H29.4163L30.0562 7.74761V7.12243L29.7716 6.83612L30.0562 6.54788V5.85648ZM29.4163 7.74761H26.9212V7.15165H29.4163V7.74761ZM29.4163 6.51672H26.9212V5.85648H29.4163V6.51672Z"
                        fill="#121926"></path>
                      <path
                        d="M39.5555 7.80798H42.3351L41.6952 8.44289H39.6222L38.9155 7.80798V5.85647L39.5555 5.28777H41.6952L42.3351 5.92268H39.5555V7.80798Z"
                        fill="#121926"></path>
                    </svg>
                  </div>
                <div style="font-size: 10px; color: #697586;">
                  Generated at: ` + now + ` (UTC)
                </div>
              </div>
            </div>`).
				WithFooterTemplate(`<div style="width: 100vw;padding-left: 36px; padding-right: 36px;">
<div style="display: flex; align-items: center; width: 100%;  justify-content: space-between; border-top: 1px solid #f8fafc; padding-top: 12px; padding-bottom: 12px;">
  <div style="font-size: 10px; color: #697586;">
    Generated by Phalcon at ` + now + ` (UTC)
  </div>
  <div style="font-size: 10px; color: #697586;">Page <span class=pageNumber></span> of <span class=totalPages></span></div>
</div>`).
				WithPaperWidth(6.2). // A4 宽度，单位：英寸
				WithPaperHeight(11.69 / 8.27 * 6.2). // A4 高度，单位：英寸
				WithMarginTop(0.7).
				WithMarginBottom(0.7).
				WithMarginLeft(0).
				WithMarginRight(0).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 3. 保存 PDF 文件
	fileName := "output.pdf"
	if err := ioutil.WriteFile(fileName, pdfBuf, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Println("PDF saved to", fileName)
}
