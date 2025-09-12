package number

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/shopspring/decimal"
	"math"
)

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
