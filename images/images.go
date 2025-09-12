package images

import (
	"bytes"
	"fmt"
	"github.com/golang/freetype"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
)

type ShareContent struct {
	FontSize   float64
	PointFixed float64
	Text       string
	FontColor  color.Color
}

func GenerateImage(x int, data []ShareContent) ([]byte, error) {
	bgImage, _, err := image.Decode(bytes.NewReader(background))
	if err != nil {
		return nil, fmt.Errorf("failed to decode background image: %v", err)
	}

	// 创建一个新的 RGBA 图像
	outputImage := image.NewRGBA(bgImage.Bounds())
	draw.Draw(outputImage, bgImage.Bounds(), bgImage, image.Point{}, draw.Src)

	fontParsed, err := freetype.ParseFont(Font)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(fontParsed)
	c.SetClip(outputImage.Bounds())
	c.SetDst(outputImage)

	pt := freetype.Pt(x, 0) // 起始位置
	for _, setting := range data {
		pt.Y += c.PointToFixed(setting.PointFixed)    // 每行间距
		c.SetFontSize(setting.FontSize)               // 字体大小
		c.SetSrc(image.NewUniform(setting.FontColor)) // 字体颜色
		_, err = c.DrawString(setting.Text, pt)
		if err != nil {
			return nil, fmt.Errorf("failed to draw string: %v", err)
		}
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, outputImage, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
