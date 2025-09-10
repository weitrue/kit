package images

import (
	"bytes"
	"encoding/base64"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/otiai10/gosseract/v2"
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestGenerateImage(t *testing.T) {
	type args struct {
		x    int
		data []ShareContent
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr error
	}{
		{
			name: "test",
			args: args{
				x: 36,
				data: []ShareContent{
					{
						FontColor:  color.White,
						PointFixed: 120,
						FontSize:   32,
						Text:       "CRAZYFROG",
					},
					{
						FontColor: color.RGBA{
							R: 0x22,
							G: 0xC5,
							B: 0x5E,
							A: 0xFF,
						},
						PointFixed: 80,
						FontSize:   64,
						Text:       fmt.Sprintf("%.1fX", 2.5),
					},
					{
						FontColor: color.RGBA{
							R: 0x22,
							G: 0xC5,
							B: 0x5E,
							A: 0xFF,
						},
						PointFixed: 36,
						FontSize:   24,
						Text:       fmt.Sprintf("+%.2f SOL", 10.01),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateImage(tt.args.x, tt.args.data)
			if nil != tt.wantErr {
				t.Errorf("GenerateImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			fmt.Println(string(got))

			bot, _ := tgbotapi.NewBotAPI("7819699755:AAFdg4Oa72JIO5xk4dg_zdcaOtP-T49cvaw")
			chatID := int64(-4762900415)
			// 创建图片消息
			photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{
				Name:  "image.jpg", // 图片文件名
				Bytes: got,
			})

			msg := `28.59% HFHT/SOL 📈

Share token with your Reflink:
https://t.me/easycoinai_bot?start=r-btcapostle-EbMT8cnhPXrTNkxnfvNC5dqCqFcGVW2cM5r1oUpMpump`
			photo.Caption = msg

			bot.Send(photo)
		})
	}
}

func TestDecodeImage(t *testing.T) {
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
