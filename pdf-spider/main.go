package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// PDFOptions PDF 生成选项
type PDFOptions struct {
	Format          string  `json:"format,omitempty"`          // 页面格式（如 "A4"）
	Width           float64 `json:"width,omitempty"`           // 自定义宽度（单位：points）
	Height          float64 `json:"height,omitempty"`          // 自定义高度（单位：points）
	Landscape       bool    `json:"landscape,omitempty"`       // 是否横向
	PrintBackground bool    `json:"printBackground,omitempty"` // 是否打印背景
	MarginTop       float64 `json:"marginTop,omitempty"`       // 上边距（单位：cm）
	MarginRight     float64 `json:"marginRight,omitempty"`     // 右边距
	MarginBottom    float64 `json:"marginBottom,omitempty"`    // 下边距
	MarginLeft      float64 `json:"marginLeft,omitempty"`      // 左边距
}

// GenerateOptions 生成 PDF 的配置选项
type GenerateOptions struct {
	Timeout        time.Duration `json:"timeout,omitempty"`        // 超时时间，默认 30 秒
	ViewportWidth  int64         `json:"viewportWidth,omitempty"`  // 视口宽度，默认 1920
	ViewportHeight int64         `json:"viewportHeight,omitempty"` // 视口高度，默认 1080
	PDFOptions     *PDFOptions   `json:"pdfOptions,omitempty"`     // PDF 选项
}

// PageDimensions 页面尺寸
type PageDimensions struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// GeneratePDFRequest API 请求结构
type GeneratePDFRequest struct {
	URL            string      `json:"url"`
	OutputPath     string      `json:"outputPath,omitempty"`
	Timeout        int         `json:"timeout,omitempty"` // 毫秒
	ViewportWidth  int64       `json:"viewportWidth,omitempty"`
	ViewportHeight int64       `json:"viewportHeight,omitempty"`
	PDFOptions     *PDFOptions `json:"pdfOptions,omitempty"`
}

// GeneratePDFResponse API 响应结构
type GeneratePDFResponse struct {
	Success bool   `json:"success"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// GeneratePDFReport 爬取页面并生成 PDF 报告
//
// 主要功能：
// 1. 等待所有网络请求完成
// 2. 等待所有 Canvas 元素绘制完成
// 3. 自适应页面尺寸生成 PDF
//
// 参数：
//   - targetURL: 要爬取的页面 URL
//   - outputPath: PDF 输出路径（默认为 ./report.pdf）
//   - options: 配置选项
//
// 返回：生成的 PDF 文件路径和错误信息
func GeneratePDFReport(targetURL, outputPath string, options *GenerateOptions) (string, error) {
	// 设置默认值
	if outputPath == "" {
		outputPath = "./report.pdf"
	}
	if options == nil {
		options = &GenerateOptions{}
	}
	if options.Timeout == 0 {
		options.Timeout = 30 * time.Second
	}
	if options.ViewportWidth == 0 {
		options.ViewportWidth = 1920
	}
	if options.ViewportHeight == 0 {
		options.ViewportHeight = 1080
	}
	if options.PDFOptions == nil {
		options.PDFOptions = &PDFOptions{
			PrintBackground: true,
			MarginTop:       1.0,
			MarginRight:     1.0,
			MarginBottom:    1.0,
			MarginLeft:      1.0,
		}
	}

	log.Println("[1/6] 正在启动浏览器...")

	// 创建 Chrome 实例
	allocCtx, cancel := chromedp.NewExecAllocator(
		context.Background(),
		append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-setuid-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
		)...,
	)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置超时
	ctx, cancel = context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	log.Printf("[2/6] 设置视口: %dx%d\n", options.ViewportWidth, options.ViewportHeight)

	// 设置视口
	if err := chromedp.Run(ctx,
		chromedp.EmulateViewport(options.ViewportWidth, options.ViewportHeight),
	); err != nil {
		return "", fmt.Errorf("设置视口失败: %w", err)
	}

	log.Printf("[3/6] 正在加载页面: %s\n", targetURL)

	// 加载页面并等待网络空闲
	if err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
	); err != nil {
		log.Printf("   ⚠ 页面加载出现问题: %v\n", err)
	}

	log.Println("   ✓ 页面加载完成")

	chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		time.Sleep(time.Second * 1)
		return nil
	}))

	// 等待 Canvas 绘制完成
	log.Println("[4/6] 等待 Canvas 图表绘制完成...")
	if err := waitForCanvasReady(ctx, options.Timeout); err != nil {
		log.Printf("   ⚠ Canvas 等待过程出错: %v\n", err)
		// 不阻塞流程，继续执行
	}

	// 获取页面实际内容尺寸
	log.Println("[5/6] 检测页面尺寸并生成 PDF...")
	dimensions, err := getPageDimensions(ctx)
	if err != nil {
		return "", fmt.Errorf("获取页面尺寸失败: %w", err)
	}
	log.Printf("   内容尺寸: %.0fx%.0fpx\n", dimensions.Width, dimensions.Height)

	// 根据内容尺寸生成最优 PDF 配置
	finalPDFOptions := generateOptimalPDFOptions(dimensions, options.PDFOptions)

	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return "", fmt.Errorf("创建输出目录失败: %w", err)
		}
	}

	// 生成 PDF
	log.Println("[6/6] 正在生成 PDF...")
	var pdfBuf []byte

	params := page.PrintToPDF().
		WithPrintBackground(finalPDFOptions.PrintBackground)

	// 设置边距（转换为英寸）
	if finalPDFOptions.MarginTop > 0 {
		params = params.WithMarginTop(finalPDFOptions.MarginTop / 2.54)
	}
	if finalPDFOptions.MarginRight > 0 {
		params = params.WithMarginRight(finalPDFOptions.MarginRight / 2.54)
	}
	if finalPDFOptions.MarginBottom > 0 {
		params = params.WithMarginBottom(finalPDFOptions.MarginBottom / 2.54)
	}
	if finalPDFOptions.MarginLeft > 0 {
		params = params.WithMarginLeft(finalPDFOptions.MarginLeft / 2.54)
	}

	// 设置页面格式或自定义尺寸
	if finalPDFOptions.Format != "" {
		params = params.WithPaperWidth(0).WithPaperHeight(0)
	} else if finalPDFOptions.Width > 0 && finalPDFOptions.Height > 0 {
		// 转换 points 到英寸 (1 inch = 72 points)
		params = params.WithPaperWidth(finalPDFOptions.Width / 72).
			WithPaperHeight(finalPDFOptions.Height / 72)
	}

	params = params.WithLandscape(finalPDFOptions.Landscape)

	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		pdfBuf, _, err = params.Do(ctx)
		return err
	})); err != nil {
		return "", fmt.Errorf("生成 PDF 失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(outputPath, pdfBuf, 0644); err != nil {
		return "", fmt.Errorf("写入 PDF 文件失败: %w", err)
	}

	log.Printf("✓ PDF 报告已生成: %s\n", outputPath)
	return outputPath, nil
}

// waitForCanvasReady 等待页面所有 Canvas 元素绘制完成
func waitForCanvasReady(ctx context.Context, maxWaitTime time.Duration) error {
	// JavaScript 代码：检查 Canvas 是否就绪
	js := `
(function() {
	const canvases = Array.from(document.querySelectorAll('canvas'));
	
	if (canvases.length === 0) {
		return { count: 0, ready: 0, message: '未发现 Canvas 元素', allReady: true };
	}
	
	// 检查 Canvas 是否有实际绘制内容
	const hasCanvasContent = (canvas) => {
		if (!canvas || canvas.width === 0 || canvas.height === 0) {
			return false;
		}
		
		try {
			const ctx2d = canvas.getContext('2d', { willReadFrequently: true });
			if (ctx2d) {
				// 采样检查：检查中心点和四个角
				const checkPoints = [
					{ x: Math.floor(canvas.width / 2), y: Math.floor(canvas.height / 2) },
					{ x: 10, y: 10 },
					{ x: canvas.width - 10, y: 10 },
					{ x: 10, y: canvas.height - 10 },
					{ x: canvas.width - 10, y: canvas.height - 10 }
				];
				
				for (const point of checkPoints) {
					if (point.x >= 0 && point.x < canvas.width && 
						point.y >= 0 && point.y < canvas.height) {
						const pixel = ctx2d.getImageData(point.x, point.y, 1, 1);
						if (pixel.data[3] > 0) {
							return true;
						}
					}
				}
				return false;
			}
			
			// 检查是否是 WebGL Canvas
			const ctxWebGL = canvas.getContext('webgl') || 
							canvas.getContext('webgl2') || 
							canvas.getContext('experimental-webgl');
			if (ctxWebGL) {
				return canvas.width > 0 && canvas.height > 0;
			}
			
			return false;
		} catch (e) {
			return canvas.width > 0 && canvas.height > 0;
		}
	};
	
	// 统计有内容的 Canvas
	let readyCount = 0;
	canvases.forEach(canvas => {
		if (hasCanvasContent(canvas)) {
			readyCount++;
		}
	});
	
	const allReady = readyCount === canvases.length && canvases.length > 0;
	
	return {
		count: canvases.length,
		ready: readyCount,
		allReady: allReady,
		message: readyCount + '/' + canvases.length + ' 个 Canvas 已就绪'
	};
})();
`

	// 使用轮询机制，正确使用 maxWaitTime 参数
	startTime := time.Now()
	checkInterval := 200 * time.Millisecond // 每 200ms 检查一次
	var lastResult map[string]interface{}

	for {
		// 检查是否超时
		elapsed := time.Since(startTime)
		if elapsed >= maxWaitTime {
			log.Printf("   ⚠ Canvas 等待超时 (%.1fs), 当前状态: %v\n", elapsed.Seconds(), lastResult["message"])
			break
		}

		// 执行检查
		var result map[string]interface{}
		if err := chromedp.Run(ctx, chromedp.Evaluate(js, &result)); err != nil {
			return fmt.Errorf("Canvas 检查失败: %w", err)
		}

		lastResult = result

		// 获取检查结果
		count, _ := result["count"].(float64)
		ready, _ := result["ready"].(float64)
		allReady, _ := result["allReady"].(bool)

		// 如果没有 Canvas 或所有 Canvas 都已就绪，立即返回
		if count == 0 {
			if message, ok := result["message"].(string); ok {
				log.Printf("   ✓ %s\n", message)
			}
			return nil
		}

		if allReady {
			if message, ok := result["message"].(string); ok {
				log.Printf("   ✓ %s (耗时: %.1fs)\n", message, elapsed.Seconds())
			}
			return nil
		}

		// 第一次检测时输出发现的 Canvas 数量
		if elapsed < checkInterval {
			log.Printf("   发现 %.0f 个 Canvas 元素，等待绘制完成...\n", count)
		}

		// 定期输出进度
		if int(elapsed/time.Second) > int((elapsed-checkInterval)/time.Second) {
			log.Printf("   进度: %.0f/%.0f Canvas 已就绪 (%.1fs)\n", ready, count, elapsed.Seconds())
		}

		// 等待一段时间后再次检查
		time.Sleep(checkInterval)
	}

	// 超时后仍返回最后的状态
	if lastResult != nil {
		if message, ok := lastResult["message"].(string); ok {
			log.Printf("   ✓ %s\n", message)
		}
	}

	return nil
}

// getPageDimensions 获取页面实际内容尺寸
func getPageDimensions(ctx context.Context) (*PageDimensions, error) {
	js := `
(function() {
	const body = document.body;
	const html = document.documentElement;
	
	const width = Math.max(
		body.scrollWidth,
		html.scrollWidth,
		body.offsetWidth,
		html.offsetWidth,
		body.clientWidth,
		html.clientWidth
	);
	
	const height = Math.max(
		body.scrollHeight,
		html.scrollHeight,
		body.offsetHeight,
		html.offsetHeight,
		body.clientHeight,
		html.clientHeight
	);
	
	return { width, height };
})();
`

	var dimensions PageDimensions
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &dimensions)); err != nil {
		return nil, err
	}

	return &dimensions, nil
}

// generateOptimalPDFOptions 根据页面内容尺寸生成最优 PDF 配置
//
// 策略：
// 1. 内容宽度 <= 210mm (A4宽度): 使用 A4 纵向
// 2. 内容宽度 <= 297mm (A4长度): 使用 A4 横向
// 3. 内容宽度 > 297mm: 使用自定义尺寸，按内容自适应
func generateOptimalPDFOptions(dimensions *PageDimensions, basePDFOptions *PDFOptions) *PDFOptions {
	// 像素转换为毫米 (1 inch = 96px = 25.4mm)
	widthMm := (dimensions.Width / 96) * 25.4
	heightMm := (dimensions.Height / 96) * 25.4

	log.Printf("   内容尺寸: %.0fmm x %.0fmm\n", widthMm, heightMm)

	// 如果用户已指定格式或尺寸，优先使用
	if basePDFOptions.Format != "" || (basePDFOptions.Width > 0 && basePDFOptions.Height > 0) {
		log.Println("   使用预设 PDF 配置")
		return basePDFOptions
	}

	// A4 尺寸：210mm x 297mm
	const A4Width = 210.0
	const A4Height = 297.0

	pdfOptions := *basePDFOptions

	if widthMm <= A4Width && heightMm <= A4Height {
		// 适合 A4 纵向
		pdfOptions.Format = "A4"
		pdfOptions.Landscape = false
		pdfOptions.Width = 595.0  // A4 宽度（points）
		pdfOptions.Height = 842.0 // A4 高度（points）
		log.Println("   使用 A4 纵向格式")
	} else if widthMm <= A4Height && heightMm <= A4Width*2 {
		// 适合 A4 横向
		pdfOptions.Format = "A4"
		pdfOptions.Landscape = true
		pdfOptions.Width = 842.0  // A4 横向宽度
		pdfOptions.Height = 595.0 // A4 横向高度
		log.Println("   使用 A4 横向格式")
	} else {
		// 使用自定义尺寸，按内容自适应
		// 转换为 PDF points (1 inch = 72 points)
		widthPoints := (dimensions.Width / 96) * 72
		heightPoints := (dimensions.Height / 96) * 72

		// 限制最大尺寸，避免 PDF 过大
		const maxSize = 2000.0 // points (约 70cm)

		pdfOptions.Width = min(widthPoints, maxSize)
		pdfOptions.Height = min(heightPoints, maxSize)
		pdfOptions.Landscape = dimensions.Width > dimensions.Height

		orientation := "纵向"
		if pdfOptions.Landscape {
			orientation = "横向"
		}
		log.Printf("   使用自定义尺寸: %.0fx%.0f points (%s)\n",
			pdfOptions.Width, pdfOptions.Height, orientation)
	}

	return &pdfOptions
}

// StartServer 启动 HTTP 服务，提供 PDF 生成 API 接口
//
// 接口说明：
//   - POST /generate-pdf - 生成 PDF 报告
//     Body: { url, outputPath?, timeout?, viewportWidth?, viewportHeight?, pdfOptions? }
//   - GET /generate-pdf?url=<URL>&output=<PATH>&timeout=<MS>
//   - GET /health - 健康检查
func StartServer(port int) error {
	http.HandleFunc("/generate-pdf", handleGeneratePDF)
	http.HandleFunc("/health", handleHealth)

	addr := fmt.Sprintf(":%d", port)
	log.Println("\n🚀 PDF 报告生成服务已启动")
	log.Printf("   端口: %d\n", port)
	log.Println("\n📖 接口文档:")
	log.Println("   POST /generate-pdf - 生成 PDF 报告")
	log.Println("        Body: { url, outputPath?, timeout?, viewportWidth?, viewportHeight?, pdfOptions? }")
	log.Println("   GET  /generate-pdf?url=<URL>&output=<PATH>&timeout=<MS>")
	log.Println("   GET  /health - 健康检查\n")

	return http.ListenAndServe(addr, nil)
}

// handleGeneratePDF 处理 PDF 生成请求
func handleGeneratePDF(w http.ResponseWriter, r *http.Request) {
	// CORS 支持
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// POST 请求
	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "读取请求体失败")
			return
		}
		defer r.Body.Close()

		var req GeneratePDFRequest
		if err := json.Unmarshal(body, &req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "解析 JSON 失败")
			return
		}

		if req.URL == "" {
			writeJSONError(w, http.StatusBadRequest, "缺少必需参数: url")
			return
		}

		options := &GenerateOptions{
			ViewportWidth:  req.ViewportWidth,
			ViewportHeight: req.ViewportHeight,
			PDFOptions:     req.PDFOptions,
		}
		if req.Timeout > 0 {
			options.Timeout = time.Duration(req.Timeout) * time.Millisecond
		}

		outputPath := req.OutputPath
		if outputPath == "" {
			outputPath = "./report.pdf"
		}

		pdfPath, err := GeneratePDFReport(req.URL, outputPath, options)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSONSuccess(w, pdfPath, "PDF 报告生成成功")
		return
	}

	// GET 请求
	if r.Method == "GET" {
		query := r.URL.Query()
		targetURL := query.Get("url")
		if targetURL == "" {
			writeJSONError(w, http.StatusBadRequest, "缺少必需参数: url")
			return
		}

		decodedURL, err := url.QueryUnescape(targetURL)
		if err != nil {
			decodedURL = targetURL
		}

		outputPath := query.Get("output")
		if outputPath == "" {
			outputPath = "./report.pdf"
		}

		options := &GenerateOptions{}
		if timeoutStr := query.Get("timeout"); timeoutStr != "" {
			if timeout, err := strconv.Atoi(timeoutStr); err == nil {
				options.Timeout = time.Duration(timeout) * time.Millisecond
			}
		}

		pdfPath, err := GeneratePDFReport(decodedURL, outputPath, options)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSONSuccess(w, pdfPath, "PDF 报告生成成功")
		return
	}

	writeJSONError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
}

// handleHealth 健康检查
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "pdf-report-generator",
	})
}

// writeJSONError 写入 JSON 错误响应
func writeJSONError(w http.ResponseWriter, statusCode int, errorMsg string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(GeneratePDFResponse{
		Success: false,
		Error:   errorMsg,
	})
}

// writeJSONSuccess 写入 JSON 成功响应
func writeJSONSuccess(w http.ResponseWriter, path, message string) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GeneratePDFResponse{
		Success: true,
		Path:    path,
		Message: message,
	})
}

// min 返回两个 float64 中的较小值
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// 主函数示例
func main() {
	// 示例 1: 直接生成 PDF
	// pdfPath, err := GeneratePDFReport(
	// 	"https://example.com/report",
	// 	"./report.pdf",
	// 	nil,
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println("PDF 已生成:", pdfPath)

	// 示例 2: 启动 HTTP 服务
	port := 3000
	if err := StartServer(port); err != nil {
		log.Fatal(err)
	}
}
