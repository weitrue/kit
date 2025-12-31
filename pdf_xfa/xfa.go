package pdf_xfa

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdf "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"io"
	"log"
	"os"
	"strings"
)

func printHongKongXML() {
	in := "~/Projects/golang/src/kit/pdf_xfa/template/HongKong_STR_form.pdf"
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
	for i := 0; i < len(xfaArr); i += 2 {
		name := objToNameString(xfaArr[i])
		if name == "datasets" {
			o, _ := ctx.Dereference(xfaArr[i+1])
			if data, ok := o.(types.StreamDict); ok {
				err = data.Decode()
				if err != nil {
					log.Fatal(err)
					return
				}
				printReasonCodes(data.Content)
			}
		}

	}
}

func printSingaporeXML() {
	in := "~/Projects/golang/src/kit/pdf_xfa/template/Singapore_STR_form.pdf"
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
	for i := 0; i < len(xfaArr); i += 2 {
		name := objToNameString(xfaArr[i])
		if name == "datasets" {
			o, _ := ctx.Dereference(xfaArr[i+1])
			if data, ok := o.(types.StreamDict); ok {
				err = data.Decode()
				if err != nil {
					log.Fatal(err)
					return
				}
				fmt.Println(string(data.Content))
			}
		}

	}
}

func printReasonCodes(xmlData []byte) error {
	dec := xml.NewDecoder(bytes.NewReader(xmlData))
	var (
		inSuspectedCrimeDetail      bool
		inSuspiciousIndicatorDetail bool
		suspectedCodes              []string
		indicatorCodes              []string
	)

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "SuspectedCrimeDetail":
				inSuspectedCrimeDetail = true
			case "SuspiciousIndicatorDetail":
				inSuspiciousIndicatorDetail = true
			case "ReasonCode":
				var v string
				if err := dec.DecodeElement(&v, &se); err != nil {
					return err
				}
				v = strings.TrimSpace(v)
				if v == "" {
					continue // 跳过 schema 里那些自闭合的空 ReasonCode
				}
				if inSuspectedCrimeDetail {
					suspectedCodes = append(suspectedCodes, v)
				} else if inSuspiciousIndicatorDetail {
					indicatorCodes = append(indicatorCodes, v)
				}
			}

		case xml.EndElement:
			switch se.Name.Local {
			case "SuspectedCrimeDetail":
				inSuspectedCrimeDetail = false
			case "SuspiciousIndicatorDetail":
				inSuspiciousIndicatorDetail = false
			}
		}
	}

	// 打印
	if len(suspectedCodes) > 0 {
		fmt.Printf("SuspectedCrime ReasonCode: %v\n", suspectedCodes)
	}
	if len(indicatorCodes) > 0 {
		fmt.Printf("SuspiciousIndicator ReasonCode: %v\n", indicatorCodes)
	}
	return nil
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
