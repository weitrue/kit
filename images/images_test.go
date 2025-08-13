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

}
