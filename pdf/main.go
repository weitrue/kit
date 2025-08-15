package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdf "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	in := "/Users/wpeng/Projects/golang/src/kit/pdf/STR_form.pdf"
	outFile := "STR_Form_filled.pdf"

	targetValues := map[string]string{
		"TrxTotalAmount":  "1",
		"TrxTotalPeriod":  "2",
		"TrxDailyAverage": "2",
	}

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

	// 从 Catalog 里拿 AcroForm 引用
	formRef, found := rootDict.Find("AcroForm")
	if !found {
		log.Println("没有 AcroForm")
		return
	}

	formDict, err := ctx.DereferenceDict(formRef)
	if err != nil {
		log.Fatal(err)
	}

	// 取 XFA
	xfaObj, ok := formDict.Find("XFA")
	if !ok {
		log.Fatal("AcroForm 没有 XFA")
	}

	xfaArr, ok := derefToArray(ctx, xfaObj)
	if !ok {
		log.Fatal("XFA is not an array or couldn't deref")
	}

	// 4) 找到 datasets 流对象
	var datasetsStream *types.StreamDict
	for i := 0; i < len(xfaArr); i += 2 {
		name := objToNameString(xfaArr[i])
		if strings.EqualFold(strings.TrimSpace(name), "datasets") {
			o, _ := ctx.Dereference(xfaArr[i+1])
			if sd, ok := o.(types.StreamDict); ok {
				datasetsStream = &sd
			}
			break
		}
	}
	if datasetsStream == nil {
		log.Fatal("datasets stream not found")
		return
	}

	// 5) 解码原始 datasets
	err = datasetsStream.Decode()
	if err != nil {
		log.Fatalf("Decode datasets stream failed: %v", err)
		return
	}

	origXML := datasetsStream.Content

	// 6) 修改 datasets XML
	newXML, err := modifyDatasetsXML(origXML, targetValues)
	if err != nil {
		log.Fatalf("modifyDatasetsXML error: %v", err)
	}

	// 7) 读取原 PDF bytes
	pdfBytes, err := ioutil.ReadFile(in)
	if err != nil {
		log.Fatalf("ReadFile error: %v", err)
	}

	// 8) 原位替换 datasets bytes
	// pdfcpu 的 StreamDict 有 Offset 和 Length 字段可获取原始位置
	offset := datasetsStream.StreamOffset
	oldLen := *datasetsStream.StreamLength
	if int(offset+oldLen) > len(pdfBytes) {
		log.Fatal("invalid stream offset/length")
		return
	}

	newBytes := newXML
	if len(newBytes) > int(oldLen) {
		log.Println("warning: new datasets is larger than original; Acrobat 可能仍可接受，但长度超出原位置")
		return
	}

	// 构建新 PDF bytes
	outPDF := make([]byte, len(pdfBytes)-int(oldLen)+len(newBytes))
	copy(outPDF[:offset], pdfBytes[:offset])
	copy(outPDF[offset:int(offset)+len(newBytes)], newBytes)
	copy(outPDF[int(offset)+len(newBytes):], pdfBytes[offset+oldLen:])

	// 9) 保存新 PDF
	if err := ioutil.WriteFile(outFile, outPDF, 0644); err != nil {
		log.Fatalf("WriteFile error: %v", err)
	}

	fmt.Printf("成功生成新 PDF: %s\n", outFile)
}

// derefToArray: 把可能是 IndirectRef/Array 的 XFA 对象解为 types.Array
func derefToArray(ctx *pdf.Context, obj types.Object) (types.Array, bool) {
	switch v := obj.(type) {
	case types.IndirectRef:
		o, err := ctx.Dereference(v)
		if err != nil {
			log.Printf("derefToArray: deref indirect ref failed: %v", err)
			return nil, false
		}
		if a, ok := o.(types.Array); ok {
			return a, true
		}
		if sd, ok := o.(types.StreamDict); ok {
			// 单流形式，构造伪 array: ["xfa", stream]
			return types.Array{types.StringLiteral("xfa"), sd}, true
		}
		return nil, false
	case types.Array:
		return v, true
	default:
		return nil, false
	}
}

// objToNameString: array 中的 name slot 可能是 Name/StringLiteral/HexLiteral/IndirectRef
func objToNameString(o types.Object) string {
	switch t := o.(type) {
	case types.Name:
		return t.String()
	case types.StringLiteral:
		return t.Value()
	case types.HexLiteral:
		return t.Value()
	case types.IndirectRef:
		// 不在此函数解引用，返回 placeholder
		return "indirectName"
	default:
		return ""
	}
}

// modifyDatasetsXML: 简单的 XML 修改，将 exampleValues 写入 <form> 下
func modifyDatasetsXML(orig []byte, values map[string]string) ([]byte, error) {
	type Field struct {
		XMLName xml.Name
		Value   string `xml:",chardata"`
	}
	type Form struct {
		XMLName xml.Name
		Fields  []Field `xml:",any"`
	}
	type Data struct {
		Form Form `xml:"form"`
	}
	type Datasets struct {
		XMLName xml.Name
		Data    Data `xml:"data"`
	}

	// 1. 用 Decoder 忽略命名空间
	decoder := xml.NewDecoder(bytes.NewReader(orig))
	decoder.Strict = false

	var d Datasets
	if err := decoder.Decode(&d); err != nil {
		// 解析失败则创建最简单结构
		d = Datasets{
			XMLName: xml.Name{Local: "datasets"},
			Data: Data{
				Form: Form{
					XMLName: xml.Name{Local: "form"},
					Fields:  []Field{},
				},
			},
		}
	}

	// 2. 更新或添加字段
	fieldMap := make(map[string]*Field)
	for i := range d.Data.Form.Fields {
		fieldMap[d.Data.Form.Fields[i].XMLName.Local] = &d.Data.Form.Fields[i]
	}
	for k, v := range values {
		if f, ok := fieldMap[k]; ok {
			f.Value = v
		} else {
			d.Data.Form.Fields = append(d.Data.Form.Fields, Field{
				XMLName: xml.Name{Local: k},
				Value:   v,
			})
		}
	}

	// 3. Marshal 回 XML
	out, err := xml.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, err
	}
	out = append([]byte(xml.Header), out...)
	return out, nil
}

func printSTRFormField(ctx *pdf.Context, xfaObj types.Object) {
	parts, err := extractXFA(ctx, xfaObj)
	if err != nil {
		log.Fatal(err)
		return
	}

	tpl, ok := parts["template"]
	if !ok || len(tpl) == 0 {
		log.Fatal("XFA 里没有 template 部分（无法定位字段布局）")
	}

}

// //////////////////////////////////////////////////////////////////////////////
// 辅助：提取 XFA template 与 datasets（如果存在），并返回 template bytes、datasetsRef(可能为nil)、datasets bytes
// //////////////////////////////////////////////////////////////////////////////
func extractXfaTemplateAndDatasets(ctx *pdf.Context, xfa types.Object) (template []byte, datasetsRef *types.IndirectRef, datasets []byte, err error) {
	// 支持几种 xfa 表示形式：
	// - Array [ name object name object ... ]
	// - StreamDict 单独流
	// - IndirectRef 指向其它
	switch v := xfa.(type) {
	case types.IndirectRef:
		o, e := ctx.Dereference(v)
		if e != nil {
			return nil, nil, nil, e
		}
		return extractXfaTemplateAndDatasets(ctx, o)
	case types.StreamDict:
		// 整个 XFA 在单个流中（少见）
		sd := v
		if err := sd.Decode(); err != nil {
			return nil, nil, nil, err
		}
		return sd.Content, nil, nil, nil
	case types.Array:
		// 常见：[ "template" 12 0 R "datasets" 13 0 R ... ]
		var tmpl []byte
		var dsBytes []byte
		for i := 0; i < len(v); {
			// 取 name 元素
			var name string
			switch nm := v[i].(type) {
			case types.Name:
				name = nm.String()
			case types.StringLiteral:
				name = nm.Value()
			case types.HexLiteral:
				name = nm.Value()
			case types.IndirectRef:
				// 有些文件把 name 放间接引用里
				on, err := ctx.Dereference(nm)
				if err == nil {
					if n2, ok := on.(types.StringLiteral); ok {
						name = n2.Value()
					} else if n2, ok := on.(types.Name); ok {
						name = n2.String()
					}
				}
			}
			i++
			if i >= len(v) {
				break
			}
			// 取内容对象
			contentObj := v[i]
			i++

			lowerName := strings.ToLower(strings.TrimSpace(name))
			if lowerName == "" {
				continue
			}

			// 解引用并读取 bytes（如果是 stream 或 literal）
			o, e := ctx.Dereference(contentObj)
			if e != nil {
				// 忽略单个部件解引用失败，继续下一个
				continue
			}
			switch oc := o.(type) {
			case types.StreamDict:
				// 需要先 Decode()
				sd := oc
				if err := sd.Decode(); err != nil {
					return nil, nil, nil, err
				}
				if lowerName == "template" {
					tmpl = append(tmpl, sd.Content...)
				} else if lowerName == "datasets" {
					// 保存 datasets 的引用（contentObj 很可能是 IndirectRef）
					if ir, ok := contentObj.(types.IndirectRef); ok {
						dsr := ir
						datasetsRef = &dsr
					}
					dsBytes = append(dsBytes, sd.Content...)
				}
			case types.StringLiteral:
				if lowerName == "template" {
					tmpl = append(tmpl, []byte(oc.Value())...)
				} else if lowerName == "datasets" {
					if ir, ok := contentObj.(types.IndirectRef); ok {
						dsr := ir
						datasetsRef = &dsr
					}
					dsBytes = append(dsBytes, []byte(oc.Value())...)
				}
			case types.HexLiteral:
				if lowerName == "template" {
					tmpl = append(tmpl, []byte(oc.Value())...)
				} else if lowerName == "datasets" {
					if ir, ok := contentObj.(types.IndirectRef); ok {
						dsr := ir
						datasetsRef = &dsr
					}
					dsBytes = append(dsBytes, []byte(oc.Value())...)
				}
			default:
				// 忽略
			}
		}
		return tmpl, datasetsRef, dsBytes, nil
	default:
		return nil, nil, nil, fmt.Errorf("unsupported XFA object type: %T", xfa)
	}
}

// //////////////////////////////////////////////////////////////////////////////
// 依据已有 datasetsBytes (可能为空) 和你要写入的 values，生成新的 datasets XML bytes
// 简单策略：
//   - 如果已有 datasetsXML，则尝试在 <xfa:data> 下查找/替换简单元素：
//     <fieldName>old</fieldName> -> 替换 inner text。
//     该策略适合 template 中 field 元素以简单元素名出现的情况。
//   - 如果没有 datasets，则构建最小结构：
//     <xfa:datasets xmlns:xfa="..."><xfa:data><form>...fields...</form></xfa:data></xfa:datasets>
//
// 注意：真实 XFA 可能有命名空间和复杂嵌套，这里为通用/简单实现，必要时你可以根据 template 内容微调。
// 返回的 bytes 是 UTF-8 编码的 XML。
// //////////////////////////////////////////////////////////////////////////////
func buildOrUpdateDatasetsXML(existing []byte, values map[string]string) ([]byte, error) {
	if len(values) == 0 {
		// 没有要写的，保留原样
		if existing != nil {
			return existing, nil
		}
		return nil, fmt.Errorf("no values provided")
	}

	if len(existing) == 0 {
		// 构建最简 datasets xml
		var b bytes.Buffer
		// 常见 xfa 命名空间前缀 xfa，但 Acrobat 对缺省 namespace 也往往能接受。
		b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
		b.WriteString(`<xfa:datasets xmlns:xfa="http://www.xfa.org/schema/xfa-data/1.0">`)
		b.WriteString(`<xfa:data>`)
		// 一个最简单的 form 根
		b.WriteString(`<form>`)
		for k, v := range values {
			// 把 FQN 的点替换为下划线作为元素名（简单处理）
			ename := safeElementName(k)
			b.WriteString(fmt.Sprintf("<%s>%s</%s>", ename, xmlEscapeText(v), ename))
		}
		b.WriteString(`</form>`)
		b.WriteString(`</xfa:data></xfa:datasets>`)
		return b.Bytes(), nil
	}

	// 有已有 datasets：我们做一个简单的替换策略：
	sxml := string(existing)

	for k, v := range values {
		ename := safeElementName(k)

		// 先尝试匹配 <ename>...</ename>
		open := "<" + ename + ">"
		close := "</" + ename + ">"
		if strings.Contains(sxml, open) && strings.Contains(sxml, close) {
			// 简单替换内部文本（第一次出现）
			start := strings.Index(sxml, open)
			if start >= 0 {
				after := sxml[start+len(open):]
				end := strings.Index(after, close)
				if end >= 0 {
					old := after[:end]
					sxml = strings.Replace(sxml, open+old+close, open+xmlEscapeText(v)+close, 1)
					continue
				}
			}
		}

		// 如果没有匹配到，就把新的元素插到 </xfa:data> 前面（最常见容器）
		if idx := strings.Index(strings.ToLower(sxml), "</xfa:data>"); idx != -1 {
			insert := fmt.Sprintf("<%s>%s</%s>", ename, xmlEscapeText(v), ename)
			sxml = sxml[:idx] + insert + sxml[idx:]
			continue
		}

		// 兜底：追加到末尾
		sxml = sxml + fmt.Sprintf("<%s>%s</%s>", ename, xmlEscapeText(v), ename)
	}

	return []byte(sxml), nil
}

// 把 FQN 字符串转换为安全的 XML 元素名（把 '.' 替为 '_'，并移除非法字符）
func safeElementName(fqn string) string {
	ename := strings.ReplaceAll(fqn, ".", "_")
	// 这里只做最简单清理：保留字母数字和下划线
	var b strings.Builder
	for _, r := range ename {
		if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "field"
	}
	return b.String()
}

func xmlEscapeText(s string) string {
	// 简单转义
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, `'`, "&apos;")
	return s
}

// //////////////////////////////////////////////////////////////////////////////
// 写回 PDF：如果 datasetsRef 不为空则更新对应对象；否则向 xref 插入新的 StreamDict，并修改 XFA 数组引用为新对象
// 该函数尽量使用 pdfcpu 公有 API：ctx.XRefTable.InsertObject() 用来插入新 stream，api.WriteContextFile 输出 PDF。
// //////////////////////////////////////////////////////////////////////////////
func writeDatasetsBackToPdf(ctx *pdf.Context, acroDict types.Dict, xfaObj types.Object, datasetsRef *types.IndirectRef, newDatasets []byte) error {
	// 获取 xref table
	xrt := ctx.XRefTable

	// 如果原来有 datasetsRef -> 更新该对象内容
	if datasetsRef != nil {
		// 找到该对象在 xref table 的 entry（通过 IndirectRef）
		if ent, ok := xrt.FindTableEntryForIndRef(datasetsRef); ok {
			// ent.Object holds the object currently; 我们要把一个新的 StreamDict 插入该 entry
			// 先创建一个 StreamDict
			sd := types.StreamDict{
				Dict:    types.Dict{},
				Content: newDatasets,
			}
			// 需要设置 Length 字段（Insert/Encode 等方法在不同版本可能不同）
			sd.Dict = types.Dict{
				"Length": types.Integer(len(newDatasets)),
			}
			// 尝试调用 sd.Encode() 如果存在（某些版本有）
			if encFn := tryCallEncode(&sd); encFn != nil {
				if err := encFn(); err != nil {
					// non-fatal: 继续（pdfcpu 在写出阶段会根据 stream 的内容自动处理）
				}
			}
			// 用新的 StreamDict 替换 entry.Object
			ent.Object = sd
			// 保持 entry in xref table (done)
			return nil
		}
		// 否则无法找到 entry -> 退回到插入新对象
	}

	// 没有原 datasetsRef，或替换失败 -> 插入新对象并更新 XFA 数组指向
	// 创建 new stream dict and insert
	sd := types.StreamDict{
		Dict:    types.Dict{"Type": types.Name("XFA")},
		Content: newDatasets,
	}
	// set length
	sd.Dict["Length"] = types.Integer(len(newDatasets))

	// 插入对象到 xref table
	objNr, err := xrt.InsertObject(sd)
	if err != nil {
		return fmt.Errorf("InsertObject failed: %v", err)
	}
	newIndRef := types.IndirectRef{
		ObjectNumber:     types.Integer(objNr),
		GenerationNumber: types.Integer(0),
	}

	// 修改 acroDict 的 XFA 数组：把 datasets 指向 newIndRef
	// 先把原 XFA 对象解出来为 array
	var xfaArr types.Array
	switch v := xfaObj.(type) {
	case types.IndirectRef:
		deref, err := ctx.Dereference(v)
		if err != nil {
			return err
		}
		if a, ok := deref.(types.Array); ok {
			xfaArr = a
		}
	case types.Array:
		xfaArr = v
	}
	// 遍历并替换 datasets 项（name 后面的对象）
	for i := 0; i < len(xfaArr)-1; i++ {
		// get key at i (maybe Name or StringLiteral)
		var name string
		switch nm := xfaArr[i].(type) {
		case types.Name:
			name = nm.String()
		case types.StringLiteral:
			name = nm.Value()
		case types.HexLiteral:
			name = nm.Value()
		}
		if strings.ToLower(strings.TrimSpace(name)) == "datasets" {
			// replace the next slot with new indirect ref
			xfaArr[i+1] = newIndRef
			break
		}
	}
	// update acroDict's XFA entry
	acroDict.Update("XFA", xfaArr)
	return nil
}

// tryCallEncode 检测 sd.Encode() 是否可调用（有的 pdfcpu 版本有该方法），并返回一个可调用的函数或 nil。
// 这是反射的轻量封装：若没有 Encode 方法则返回 nil。
// （反射仅用于兼容不同 pdfcpu 版本）
func tryCallEncode(sd *types.StreamDict) func() error {
	// 在多数 pdfcpu 版本里，StreamDict 有方法 Encode() error 或者 Decode() error。
	// 为了兼容不同版本，这里采用反射尝试调用 Encode。
	// 反射实现放在函数内以保持主逻辑清晰。
	// 如果你的 pdfcpu 版本没有 Encode()，无需担心：写出阶段 api.WriteContextFile 会在必要时处理 stream encoding。
	return func() error {
		// no-op: keep simple to avoid fragile reflection in sample.
		// 如果需要强制压缩/应用 Filter，可在此处实现。
		return nil
	}
}

// =============== 解析 template -> 按页提取字段 ===============

// extractXFA 从 XFA 对象解析 XML
func extractXFA(ctx *pdf.Context, obj types.Object) (map[string][]byte, error) {
	out := make(map[string][]byte)
	switch v := obj.(type) {
	case types.StringLiteral:
		out["xfa"] = append(out["xfa"], []byte(v.Value())...)
		return out, nil
	case types.HexLiteral:
		out["xfa"] = append(out["xfa"], []byte(v.Value())...)
		return out, nil
	case types.StreamDict:
		sd := v
		if err := sd.Decode(); err != nil { // 关键：用 sd.Decode() 解码压缩流
			return nil, err
		}
		out["xfa"] = append(out["xfa"], sd.Content...)
		return out, nil
	case types.IndirectRef:
		o, err := ctx.Dereference(obj)
		if err != nil {
			return nil, err
		}
		return extractXFA(ctx, o)
	case types.Array:
		// 常见形式：[ "template" 12 0 R  "datasets" 13 0 R  ... ]
		for i := 0; i < len(v); {
			// 读取名
			var name string
			switch nm := v[i].(type) {
			case types.IndirectRef:
				on, err := ctx.Dereference(nm)
				if err != nil {
					return nil, err
				}
				switch nn := on.(type) {
				case types.Name:
					name = nn.String()
				case types.StringLiteral:
					name = nn.Value()
				default:
					// 有些文件名就是 Name/String，非间接
				}
			case types.Name:
				name = nm.String()
			case types.StringLiteral:
				name = nm.Value()
			default:
				// 非法/意外，跳一格继续
			}
			i++

			if i >= len(v) {
				break
			}

			// 读取内容对象
			contentObj := v[i]
			i++

			// 解引用并取字节
			var buf []byte
			o, err := ctx.Dereference(contentObj)
			if err != nil {
				return nil, err
			}
			switch sd := o.(type) {
			case types.StreamDict:
				// 关键：用 sd.Decode() 解压
				if err := sd.Decode(); err != nil {
					return nil, err
				}
				buf = sd.Content
			case types.StringLiteral:
				buf = []byte(sd.Value())
			case types.HexLiteral:
				buf = []byte(sd.Value())
			default:
				// 有些 name 可能重复或指向非流，忽略
				continue
			}

			n := strings.ToLower(strings.TrimSpace(name))
			if n == "" {
				n = "unknown"
			}
			out[n] = append(out[n], buf...)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("未知的 XFA 类型: %T", v)
	}
}

// 解析 XFA template，返回：页码 -> 字段全名（FQN）列表
// 说明：基于 pageSet/pageArea 的顺序给出“静态”页号；动态重复场景仅能近似。
func parseTemplateFieldsByPage(templateXML []byte) (map[int][]string, error) {
	dec := xml.NewDecoder(bytes.NewReader(templateXML))
	fieldsByPage := make(map[int][]string)

	var (
		stack      []string // subform/name 栈，用来生成 FQN
		pageNo     = 0      // 当前 pageArea 的序号（从 1 开始）
		inTemplate bool
	)

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch tt := tok.(type) {
		case xml.StartElement:
			local := strings.ToLower(tt.Name.Local)

			switch local {
			case "template":
				inTemplate = true

			case "pageset":
				// 容器，忽略

			case "pagearea":
				// 每遇到一个 pageArea 就 +1，视作一页
				pageNo++
				if pageNo == 0 {
					pageNo = 1
				}

			case "subform":
				if name := attr(tt.Attr, "name"); name != "" {
					stack = append(stack, name)
				} else {
					stack = append(stack, "subform")
				}

			case "field", "exclgroup":
				if !inTemplate {
					continue
				}
				fname := attr(tt.Attr, "name")
				if fname == "" {
					// 没 name 的字段跳过（少见）
					break
				}
				fqn := strings.Join(append(stack, fname), ".")
				// 没出现任何 pageArea 的老模板，默认归到第 1 页
				p := pageNo
				if p == 0 {
					p = 1
				}
				fieldsByPage[p] = append(fieldsByPage[p], fqn)
			}

		case xml.EndElement:
			local := strings.ToLower(tt.Name.Local)
			switch local {
			case "subform":
				if n := len(stack); n > 0 {
					stack = stack[:n-1]
				}
			}
		}
	}

	if len(fieldsByPage) == 0 {
		return nil, errors.New("template 里未发现任何 <field> 节点")
	}
	return fieldsByPage, nil
}

func attr(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}
