package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
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
