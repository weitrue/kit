package main

import (
	"fmt"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"log"
	"os"
	"time"

	"github.com/jung-kurt/gofpdf"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdf "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// ========== 数据结构（按 HKMA STR 常见栏目分组） ==========

type ReporterInfo struct {
	InstitutionName string // 报告机构名称
	InstitutionID   string // 机构/牌照编号（如有）
	Department      string // 部门
	ContactName     string // 联系人
	ContactEmail    string // 邮箱
	ContactPhone    string // 电话
}

type SubjectIDDoc struct {
	Type   string // 证件类型 e.g. "HKID", "Passport"
	Number string // 证件号
}

type SubjectInfo struct {
	Name        string
	Alias       string
	Nationality string
	Address     string
	AccountNo   string
	CustomerID  string
	IDDocs      []SubjectIDDoc
}

type Transaction struct {
	TxID           string
	TxTimeUnix     int64  // 交易时间（Unix 秒）
	Channel        string // 渠道 e.g. "Online/Mobile/Branch"
	Currency       string // 货币
	Amount         string // 金额（字符串可保留高精度）
	Counterparty   string // 对手方名称/账户
	CounterpartyId string // 对手方标识（如账号/钱包地址）
	Description    string // 交易说明/备注
}

type STRReport struct {
	// 基础元信息
	CaseRef        string    // 内部案件编号
	ReportDate     time.Time // 报告日期
	Reporter       ReporterInfo
	Subject        SubjectInfo
	Transactions   []Transaction
	Reason         string // 可疑原因（尽量具体，引用触发规则/阈值）
	ActionsTaken   string // 已采取措施（如冻结/加强尽调/拒绝交易等）
	AttachmentNote string // 附件/证据列表说明（如截图、对账单、聊天记录等）
	DeclarantName  string // 申报人姓名
	DeclarantTitle string // 申报人职务
	DeclarantSign  string // 申报人签名占位（如“/s/ John Chan”）
}

// ========== PDF 渲染 ==========

type STRPDFOptions struct {
	Title             string
	FontRegularTTF    string // UTF-8 字体：常规
	FontBoldTTF       string // UTF-8 字体：粗体
	OutputPath        string
	SetFileTimes      bool
	FileCreateTime    time.Time // 用于 os.Chtimes（文件系统时间）
	FileModifyTime    time.Time
	PageMarginLeftMM  float64
	PageMarginTopMM   float64
	PageMarginRightMM float64
}

func RenderSTRToPDF(r STRReport, opt STRPDFOptions) (string, error) {
	if opt.OutputPath == "" {
		opt.OutputPath = fmt.Sprintf("/Users/wpeng/Projects/golang/src/kit/pdf/HKMA_STR_%s.pdf", safeFileName(r.CaseRef))
	}
	if opt.Title == "" {
		opt.Title = "HKMA Suspicious Transaction Report (e-STR)"
	}
	if opt.PageMarginLeftMM == 0 && opt.PageMarginTopMM == 0 && opt.PageMarginRightMM == 0 {
		opt.PageMarginLeftMM, opt.PageMarginTopMM, opt.PageMarginRightMM = 15, 15, 15
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(opt.PageMarginLeftMM, opt.PageMarginTopMM, opt.PageMarginRightMM)
	pdf.SetAutoPageBreak(true, 12)

	// 字体（中文需要 UTF-8 字体）
	useBold := false
	if opt.FontRegularTTF != "" {
		pdf.AddUTF8Font("Body", "", opt.FontRegularTTF)
		pdf.SetFont("Body", "", 11)
		if opt.FontBoldTTF != "" {
			pdf.AddUTF8Font("BodyBold", "", opt.FontBoldTTF)
			useBold = true
		}
	} else {
		// 无自定义字体则使用内置 Helvetica（仅英文/ASCII）
		pdf.SetFont("Helvetica", "", 11)
	}

	// 页眉页脚
	pdf.SetHeaderFuncMode(func() {
		if useBold {
			pdf.SetFont("BodyBold", "", 12)
		} else {
			pdf.SetFont("Helvetica", "B", 12)
		}
		pdf.CellFormat(0, 8, opt.Title, "", 1, "L", false, 0, "")
		// 细线
		pdf.SetDrawColor(200, 200, 200)
		pdf.Line(opt.PageMarginLeftMM, pdf.GetY(), 210-opt.PageMarginRightMM, pdf.GetY())
		pdf.Ln(2)
	}, true)

	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetDrawColor(200, 200, 200)
		pdf.Line(opt.PageMarginLeftMM, pdf.GetY(), 210-opt.PageMarginRightMM, pdf.GetY())
		pdf.SetY(-10)
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(0, 8, fmt.Sprintf("Case: %s | Page %d/{nb}", r.CaseRef, pdf.PageNo()), "", 0, "R", false, 0, "")
	})
	pdf.AliasNbPages("")

	pdf.AddPage()

	// 标题区
	writeKV(pdf, "Case Reference", r.CaseRef, useBold)
	writeKV(pdf, "Report Date", r.ReportDate.Format("2006-01-02 15:04:05 MST"), useBold)
	pdf.Ln(2)

	// 1. Reporter Information
	section(pdf, "1. Reporter Information", useBold)
	writeKV(pdf, "Institution Name", r.Reporter.InstitutionName, useBold)
	writeKV(pdf, "Institution/License ID", r.Reporter.InstitutionID, useBold)
	writeKV(pdf, "Department", r.Reporter.Department, useBold)
	writeKV(pdf, "Contact Name", r.Reporter.ContactName, useBold)
	writeKV(pdf, "Contact Email", r.Reporter.ContactEmail, useBold)
	writeKV(pdf, "Contact Phone", r.Reporter.ContactPhone, useBold)

	// 2. Subject Information
	section(pdf, "2. Subject (Customer) Information", useBold)
	writeKV(pdf, "Name", r.Subject.Name, useBold)
	writeKV(pdf, "Alias", r.Subject.Alias, useBold)
	writeKV(pdf, "Nationality", r.Subject.Nationality, useBold)
	writeKV(pdf, "Address", r.Subject.Address, useBold)
	writeKV(pdf, "Account No.", r.Subject.AccountNo, useBold)
	writeKV(pdf, "Customer ID", r.Subject.CustomerID, useBold)
	if len(r.Subject.IDDocs) > 0 {
		if useBold {
			pdf.SetFont("BodyBold", "", 11)
		} else {
			pdf.SetFont("Helvetica", "B", 11)
		}
		pdf.CellFormat(0, 6, "Identification Documents:", "", 1, "L", false, 0, "")
		if useBold {
			pdf.SetFont("Body", "", 11)
		} else {
			pdf.SetFont("Helvetica", "", 11)
		}
		for i, idd := range r.Subject.IDDocs {
			writeKV(pdf, fmt.Sprintf("  #%d Type", i+1), idd.Type, useBold)
			writeKV(pdf, fmt.Sprintf("  #%d Number", i+1), idd.Number, useBold)
		}
	}

	// 3. Transaction Details（表格）
	section(pdf, "3. Suspicious Transaction Details", useBold)
	if len(r.Transactions) == 0 {
		writeKV(pdf, "Transactions", "(none)", useBold)
	} else {
		renderTxTable(pdf, r.Transactions, useBold)
	}

	// 4. Reason for Suspicion
	section(pdf, "4. Reason for Suspicion", useBold)
	multiLine(pdf, r.Reason)

	// 5. Actions Taken
	section(pdf, "5. Actions Taken", useBold)
	multiLine(pdf, r.ActionsTaken)

	// 6. Attachment / Evidence Notes
	section(pdf, "6. Attachments / Evidence", useBold)
	multiLine(pdf, r.AttachmentNote)

	// 7. Declarant
	section(pdf, "7. Declarant", useBold)
	writeKV(pdf, "Name", r.DeclarantName, useBold)
	writeKV(pdf, "Title", r.DeclarantTitle, useBold)
	writeKV(pdf, "Signature", r.DeclarantSign, useBold)

	// 导出
	if err := pdf.OutputFileAndClose(opt.OutputPath); err != nil {
		return "", err
	}

	// 可选：设置文件系统时间（创建/修改时间）
	if opt.SetFileTimes {
		// on Unix, atime 与 mtime 可改；“创建时间”严格意义上不可写，但很多场景用 mtime 代表
		at := opt.FileCreateTime
		mt := opt.FileModifyTime
		if at.IsZero() {
			at = time.Now()
		}
		if mt.IsZero() {
			mt = at
		}
		_ = os.Chtimes(opt.OutputPath, at, mt)
	}

	return opt.OutputPath, nil
}

// ========== 工具渲染函数 ==========

func section(pdf *gofpdf.Fpdf, title string, useBold bool) {
	pdf.Ln(2)
	if useBold {
		pdf.SetFont("BodyBold", "", 12)
	} else {
		pdf.SetFont("Helvetica", "B", 12)
	}
	pdf.CellFormat(0, 7, title, "", 1, "L", false, 0, "")
	pdf.SetDrawColor(30, 144, 255)
	y := pdf.GetY()
	pdf.Line(15, y, 195, y)
	if useBold {
		pdf.SetFont("Body", "", 11)
	} else {
		pdf.SetFont("Helvetica", "", 11)
	}
	pdf.Ln(1)
}

func writeKV(pdf *gofpdf.Fpdf, k, v string, useBold bool) {
	// label
	if useBold {
		pdf.SetFont("BodyBold", "", 11)
	} else {
		pdf.SetFont("Helvetica", "B", 11)
	}
	pdf.CellFormat(50, 6, k, "", 0, "L", false, 0, "")
	// value
	if useBold {
		pdf.SetFont("Body", "", 11)
	} else {
		pdf.SetFont("Helvetica", "", 11)
	}
	pdf.MultiCell(0, 6, v, "", "L", false)
}

func multiLine(pdf *gofpdf.Fpdf, text string) {
	if text == "" {
		text = "(none)"
	}
	pdf.MultiCell(0, 6, text, "", "L", false)
}

func renderTxTable(pdf *gofpdf.Fpdf, txs []Transaction, useBold bool) {
	// 表头
	if useBold {
		pdf.SetFont("BodyBold", "", 10)
	} else {
		pdf.SetFont("Helvetica", "B", 10)
	}
	headers := []string{"#",
		"Tx ID", "Time", "Channel", "Currency", "Amount",
		"Counterparty", "Counterparty ID", "Description"}
	widths := []float64{8, 30, 28, 20, 18, 22, 32, 35, 0}

	for i, h := range headers {
		w := widths[i]
		pdf.CellFormat(w, 7, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// 内容
	if useBold {
		pdf.SetFont("Body", "", 9)
	} else {
		pdf.SetFont("Helvetica", "", 9)
	}
	for i, t := range txs {
		row := []string{
			fmt.Sprintf("%d", i+1),
			t.TxID,
			time.Unix(t.TxTimeUnix, 0).Format("2006-01-02 15:04"),
			t.Channel,
			t.Currency,
			t.Amount,
			t.Counterparty,
			t.CounterpartyId,
			t.Description,
		}
		for col, val := range row {
			w := widths[col]
			// 最后一列用 MultiCell 占满
			if col == len(row)-1 {
				x, y := pdf.GetX(), pdf.GetY()
				pdf.MultiCell(0, 6, val, "1", "L", false)
				pdf.SetXY(x+0, y) // 下一行从整行宽度后开始换行
			} else {
				pdf.CellFormat(w, 6, val, "1", 0, "L", false, 0, "")
			}
		}
		pdf.Ln(-1)
	}
}

func safeFileName(s string) string {
	if s == "" {
		return "no_ref"
	}
	// 简化：替换不安全字符
	out := s
	for _, bad := range []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|'} {
		out = stringReplaceAllRune(out, bad, '_')
	}
	return out
}
func stringReplaceAllRune(s string, old rune, new rune) string {
	runes := []rune(s)
	for i, r := range runes {
		if r == old {
			runes[i] = new
		}
	}
	return string(runes)
}

// ========== 示例演示 ==========
func main1() {
	path := "/Users/wpeng/Projects/golang/src/kit/pdf/fonts/NotoSansSC-Regular.ttf"
	_, err := os.Stat(path)
	if err != nil {
		fmt.Println("Stat error:", err)
	} else {
		fmt.Println("File exists!")
	}
}

func main2() {
	report := STRReport{
		CaseRef:    "STR-2025-0001",
		ReportDate: time.Now(),
		Reporter: ReporterInfo{
			InstitutionName: "ABC Bank (HK)",
			InstitutionID:   "BK12345",
			Department:      "AML Compliance",
			ContactName:     "Chan Tai Man",
			ContactEmail:    "chan.tm@abcbank.com",
			ContactPhone:    "+852-1234-5678",
		},
		Subject: SubjectInfo{
			Name:        "LEE SIU MING",
			Alias:       "LEE S.M.",
			Nationality: "HKSAR",
			Address:     "Room 1201, XXX Building, Central, Hong Kong",
			AccountNo:   "012-345678-001",
			CustomerID:  "CUST-889900",
			IDDocs: []SubjectIDDoc{
				{Type: "HKID", Number: "A123456(7)"},
			},
		},
		Transactions: []Transaction{
			{
				TxID:           "TX-001",
				TxTimeUnix:     time.Now().Add(-6 * time.Hour).Unix(),
				Channel:        "Online",
				Currency:       "HKD",
				Amount:         "985,000.00",
				Counterparty:   "XYZ LIMITED",
				CounterpartyId: "Acct 11223344",
				Description:    "Multiple large transfers in short period; mismatch with KYC profile.",
			},
			{
				TxID:           "TX-002",
				TxTimeUnix:     time.Now().Add(-2 * time.Hour).Unix(),
				Channel:        "Branch",
				Currency:       "USD",
				Amount:         "180,000.00",
				Counterparty:   "JOHN DOE",
				CounterpartyId: "Acct 55667788",
				Description:    "Cash deposit followed by outward remittance to high-risk jurisdiction.",
			},
		},
		Reason: `Triggered by rules:
- Rapid movement of high-value funds
- Transaction patterns inconsistent with declared business nature
- Counterparty in higher-risk geography`,
		ActionsTaken: `- Temporarily held outgoing payment pending review
- Conducted EDD and contacted customer for explanation
- Filed STR to HKMA e-STR platform`,
		AttachmentNote: `Attached: account statements (last 90 days), online banking logs, onboarding KYC file, branch CCTV snapshot (reference only).`,
		DeclarantName:  "WONG KA HO",
		DeclarantTitle: "AMLO",
		DeclarantSign:  "/s/ WONG KA HO",
	}

	// 字体（如需中文，请下载 NotoSansCJK 或思源黑体等 TTF/OTF）
	fontDir := "/Users/wpeng/Projects/golang/src/kit/pdf/fonts"
	_ = os.MkdirAll(fontDir, 0755)
	opt := STRPDFOptions{
		Title: "HKMA Suspicious Transaction Report (e-STR) — Internal Generated",
		//FontRegularTTF: "/Users/wpeng/Projects/golang/src/kit/pdf/fonts/NotoSansSC-Regular.ttf", // 若无中文可留空
		//FontBoldTTF:    "/Users/wpeng/Projects/golang/src/kit/pdf/fonts/NotoSansSC-Bold.ttf",
		OutputPath:     "HKMA_STR_demo.pdf",
		SetFileTimes:   true,
		FileCreateTime: time.Now().Add(-1 * time.Hour),
		FileModifyTime: time.Now(),
	}
	out, err := RenderSTRToPDF(report, opt)
	if err != nil {
		panic(err)
	}
	fmt.Println("PDF generated:", out)
}

// Example of filling PDF formdata with a form.

// fillFields loads field data from `jsonPath` and used to fill in form data in `inputPath` and outputs
// as PDF in `outputPath`. The output PDF form is flattened.

func main() {
	in := "/Users/wpeng/Projects/golang/src/kit/pdf/STR_form.pdf"
	//out := "data.pdf"
	//data := "/Users/wpeng/Projects/golang/src/kit/pdf/sample_form.json"

	conf := pdf.NewDefaultConfiguration()
	conf.ValidationMode = pdf.ValidationRelaxed // 遇到不规范PDF不直接报错
	conf.Cmd = pdf.VALIDATE

	f, err := os.Open(in)
	if err != nil {
		return
	}
	defer f.Close()
	ctx, err := api.ReadContext(f, conf)
	if err != nil {
		log.Fatalf("ReadContextFile error: %v", err)
	}

	rootDict, err := ctx.XRefTable.Catalog()
	if err != nil {
		log.Fatalf("catalog error: %v", err)
	}

	pagesRef := rootDict.IndirectRefEntry("Pages")
	if pagesRef == nil {
		log.Fatalf("Root 没有 /Pages")
	}
	fmt.Println(pagesRef)

	d, err := ctx.DereferenceDict(*pagesRef)
	if err != nil || d == nil {
		return
	}

	kids := d.ArrayEntry("Kids")
	arr, err := ctx.DereferenceArray(kids)
	if err != nil {
		return
	}

	fmt.Println(arr)

}

// 递归展开页树，收集所有 Page 的字典
func collectPages(ctx *pdf.Context, node types.Object, out *[]types.Dict) error {
	d, err := ctx.DereferenceDict(node) // node 可为 Dict 或 IndirectRef(值)
	if err != nil || d == nil {
		return fmt.Errorf("Dereference 页树节点失败: %v", err)
	}
	if t := d.NameEntry("Type"); t != nil && *t == "Page" {
		*out = append(*out, d)
		return nil
	}
	// Pages 节点：下探 Kids
	kids := d.ArrayEntry("Kids")
	arr, err := ctx.DereferenceArray(kids)
	if err != nil {
		return fmt.Errorf("Dereference Kids 失败: %v", err)
	}
	for _, k := range arr {
		if err := collectPages(ctx, k, out); err != nil {
			return err
		}
	}
	return nil
}
func form1() {
	in := "/Users/wpeng/Projects/golang/src/kit/pdf/sample_form.pdf"
	//out := "data.pdf"
	//data := "/Users/wpeng/Projects/golang/src/kit/pdf/sample_form.json"

	conf := pdf.NewDefaultConfiguration()
	conf.ValidationMode = pdf.ValidationRelaxed // 遇到不规范PDF不直接报错
	conf.Cmd = pdf.VALIDATE

	f, err := os.Open(in)
	if err != nil {
		return
	}
	defer f.Close()
	ctx, err := api.ReadContext(f, conf)
	if err != nil {
		log.Fatalf("ReadContextFile error: %v", err)
	}

	rootDict, err := ctx.XRefTable.Catalog()
	if err != nil {
		log.Fatalf("catalog error: %v", err)
	}

	// 2. 找 AcroForm
	acroFormRef := rootDict["AcroForm"]
	if acroFormRef == nil {
		log.Fatalf("没有找到 AcroForm")
	}

	acroFormDict, err := ctx.DereferenceDict(acroFormRef)
	if err != nil {
		log.Fatalf("Dereference AcroForm error: %v", err)
	}

	// 3. 找 Fields 数组
	fields := acroFormDict.ArrayEntry("Fields")
	if fields == nil {
		log.Fatalf("AcroForm 没有 Fields")
	}

	arr, err := ctx.DereferenceArray(fields)
	if err != nil {
		log.Fatalf("DereferenceArray error: %v", err)
	}

	fmt.Println("AcroForm 字段列表：")
	for _, f := range arr {
		processField(ctx, f)
	}
}

func processField(ctx *pdf.Context, obj types.Object) {
	dict, err := ctx.DereferenceDict(obj)
	if err != nil {
		log.Printf("Dereference field error: %v", err)
		return
	}
	if dict == nil {
		return
	}

	// 取字段名
	if t := dict.StringEntry("T"); t != nil {
		fmt.Printf("字段名: %s\n", *t)
	}

	// 取字段值
	if v := dict.StringEntry("V"); v != nil {
		fmt.Printf("  字段值: %s\n", *v)
	}

	// 有子字段（Kids）
	if kids := dict.ArrayEntry("Kids"); kids != nil {
		arr, _ := ctx.DereferenceArray(kids)
		for _, kid := range arr {
			processField(ctx, kid)
		}
	}
}

func ReadField() {
	in := "/Users/wpeng/Projects/golang/src/kit/pdf/sample_form.pdf"
	//out := "data.pdf"
	//data := "/Users/wpeng/Projects/golang/src/kit/pdf/sample_form.json"

	conf := pdf.NewDefaultConfiguration()
	f0, err := os.Open(in)
	if err != nil {
		log.Fatal(err)
		return
	}

	data, err := api.FormFields(f0, conf)
	if err != nil {
		log.Fatal(err)
	}

	for _, field := range data {
		fmt.Println(field.ID, field.Name, field.V, field.AltName)
	}
}

//// 获取PDF所有表单字段
//func getFormFields(inputPath string) ([]*form.Field, error) {
//	// 加载PDF文件
//	f, err := os.Open(inputPath)
//	if err != nil {
//		return nil, fmt.Errorf("无法打开PDF文件: %v", err)
//	}
//	defer f.Close()
//
//	// 解析PDF
//	ctx, err := api.ReadContext(f, pdf.NewDefaultConfiguration())
//	if err != nil {
//		return nil, fmt.Errorf("PDF解析失败: %v", err)
//	}
//
//	// 获取AcroForm（交互式表单）
//	if ctx.Form == nil {
//		return nil, fmt.Errorf("该PDF没有交互式表单字段")
//	}
//
//	return ctx.Fields, nil
//}
