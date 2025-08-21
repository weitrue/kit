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
	printSingaporeXML()
}

func ReplaceXFA() {
	in := "/Users/wpeng/Projects/golang/src/kit/pdf/HongKong_STR_form_O.pdf"
	outFile := "STR_Form_f.pdf"

	// 读取新的 XML 数据
	xmlData, err := ioutil.ReadFile("/Users/wpeng/Projects/golang/src/kit/pdf/template/dataset.xml")
	if err != nil {
		log.Fatal("Error reading XML data:", err)
		return
	}

	lines := strings.Split(string(xmlData), "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanLines = append(cleanLines, trimmed)
		}
	}
	xmlData = []byte(strings.Join(cleanLines, "\n"))

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

	xfaArr, err := ctx.DereferenceArray(xfaObj)
	if !ok {
		log.Fatal("XFA is not an array or couldn't deref")
	}

	// 4) 找到 datasets 流对象
	for i := 0; i < len(xfaArr); i += 2 {
		name := objToNameString(xfaArr[i])
		if strings.EqualFold(strings.TrimSpace(name), "datasets") {
			targetObj := xfaArr[i+1]
			s, _ := ctx.Dereference(targetObj)
			data := s.(types.StreamDict)
			data.Content = xmlData
			err = data.Encode()
			if err != nil {
				log.Fatalf("encode err: %v", err)
				return
			}

			sObjNr, err := ctx.IndRefForNewObject(data)
			if err != nil {
				log.Fatal("IndRefForNewObject error")
				return
			}

			xfaArr[i+1] = *sObjNr
			break
		}
	}

	// 写回 PDF
	err = api.WriteContextFile(ctx, outFile)
	if err != nil {
		return
	}
}

func print(ctx *pdf.Context, xfaArr types.Array) {
	for i := 0; i+1 < len(xfaArr); i += 2 {
		// name entry
		name := objToNameString(xfaArr[i])
		fmt.Printf("Entry %d: name=%s\n", i/2, name)

		// value entry
		val := xfaArr[i+1]
		fmt.Printf("  value type: %T\n", val)

		// if indirect ref, print obj number and stream info
		if indRef, ok := val.(types.IndirectRef); ok {
			objNr := indRef.ObjectNumber.Value()
			gen := indRef.GenerationNumber.Value()
			fmt.Printf("  -> IndirectRef %d %d R\n", objNr, gen)
			obj, err := ctx.Dereference(indRef)
			if err != nil {
				fmt.Printf("     deref error: %v\n", err)
				continue
			}
			switch s := obj.(type) {
			case types.StreamDict:
				fmt.Printf("     stream found: Length=%v, keys=", s.Dict["Length"])
				for k := range s.Dict {
					fmt.Printf("%s ", k)
				}
				fmt.Println()
				if f, ok := s.Dict.Find("Filter"); ok {
					fmt.Printf("     Filter: %T -> %v\n", f, f)
				}
			default:
				fmt.Printf("     dereferenced to type: %T\n", obj)
			}
		} else {
			// direct object - maybe a stream inline (rare)
			obj, err := ctx.Dereference(val)
			if err == nil {
				switch s := obj.(type) {
				case types.StreamDict:
					fmt.Printf("  -> Inline stream: Length=%v, keys=", s.Dict["Length"])
					for k := range s.Dict {
						fmt.Printf("%s ", k)
					}
					fmt.Println()
					if f, ok := s.Dict.Find("Filter"); ok {
						fmt.Printf("     Filter: %T -> %v\n", f, f)
					}
				default:
					fmt.Printf("  -> Inline type: %T\n", obj)
				}
			}
		}
	}
}

func printHongKongXML() {
	in := "/Users/wpeng/Projects/golang/src/kit/pdf/template/HongKong_STR_form.pdf"
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
	in := "/Users/wpeng/Projects/golang/src/kit/pdf/Singapore_STR_O.pdf"
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

func demo1() {
	in := "/Users/wpeng/Projects/golang/src/kit/pdf/template/HongKong_STR_form.pdf"
	outFile := "STR_Form_filled.pdf"

	targetValues := map[string]string{
		"TrxTotalAmount":  "<TrxTotalAmount>100.00</TrxTotalAmount>",
		"TrxTotalPeriod":  "<TrxTotalPeriod>1.00000000</TrxTotalPeriod>",
		"TrxDailyAverage": "<TrxDailyAverage>100.00</TrxDailyAverage>",
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
	var templateStream *types.StreamDict
	var formsStream *types.StreamDict
	count := 0
	for i := 0; i < len(xfaArr); i += 2 {
		name := objToNameString(xfaArr[i])
		if strings.EqualFold(strings.TrimSpace(name), "datasets") {
			o, _ := ctx.Dereference(xfaArr[i+1])
			if sd, ok := o.(types.StreamDict); ok {
				datasetsStream = &sd
			}

			count++
		}

		if strings.EqualFold(strings.TrimSpace(name), "template") {
			o, _ := ctx.Dereference(xfaArr[i+1])
			if sd, ok := o.(types.StreamDict); ok {
				templateStream = &sd
			}
			count++
		}

		if strings.EqualFold(strings.TrimSpace(name), "form") {
			o, _ := ctx.Dereference(xfaArr[i+1])
			if sd, ok := o.(types.StreamDict); ok {
				formsStream = &sd
			}
			count++
		}

		if count == 3 {
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

	if templateStream == nil {
		log.Fatal("template stream not found")
		return
	}

	err = templateStream.Decode()
	if err != nil {
		log.Fatalf("Decode template stream failed: %v", err)
		return
	}

	if formsStream == nil {
		log.Fatal("template stream not found")
		return
	}

	err = formsStream.Decode()
	if err != nil {
		log.Fatalf("Decode template stream failed: %v", err)
		return
	}

	origDataXML := datasetsStream.Content
	templateXML := templateStream.Content
	form := formsStream.Content

	fmt.Println(templateXML)
	fmt.Println(form)
	fmt.Println(origDataXML)
	fmt.Println(targetValues)

	// 6) 修改 datasets XML
	newXML := newDatasetsXML()

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

func newDatasetsXML() []byte {
	s := `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<xfa:datasets xmlns:xfa="http://www.xfa.org/schema/xfa-data/1.0/">
    <xfa:data>
        <STR>
            <VersionNo>3.0</VersionNo>
            <mergeFormReload>FALSE</mergeFormReload>
            <CertExpiryDate>
                <body xmlns="http://www.w3.org/1999/xhtml">
                    <p style="margin-top:0in;margin-bottom:0in;">
                        <span style="xfa-spacerun:yes"> </span>
                    </p>
                </body>
            </CertExpiryDate>
            <STRHeader>
                <InstitutionID/>
                <RptOfficerID/>
                <Cap000>false</Cap000>
                <isRelatedToExistingInvestigation>false</isRelatedToExistingInvestigation>
                <isRelatedToPreviousDisclosure>false</isRelatedToPreviousDisclosure>
                <NumOfSubject/>
                <NumOfOrganization/>
                <NumOfAccount/>
                <NumOfPhone/>
                <NumOfAddress/>
                <NumOfTransaction/>
            </STRHeader>
            <EntityPersonList>
                <EntityPersonDetail>
                    <PersonID>20250818160027069</PersonID>
                    <ChiName>
                        <ChineseCommercialCode>
                            <FourDigitCode/>
                        </ChineseCommercialCode>
                    </ChiName>
                    <HKID>
                        <Identifier/>
                    </HKID>
                    <Sex>U</Sex>
                    <EmailList>
                        <EmailDetail>
                            <Email/>
                        </EmailDetail>
                    </EmailList>
                </EntityPersonDetail>
            </EntityPersonList>
            <EntityOrganisationList>
                <EntityOrganisationDetail>
                    <OrgID>20250818160027070</OrgID>
                    <LocalCompanyIndicator>false</LocalCompanyIndicator>
                    <LocalBusinessRegNumber>
                        <BusinessRegistrationNumber/>
                    </LocalBusinessRegNumber>
                    <OverseasCompanyIndicator>false</OverseasCompanyIndicator>
                    <EmailList>
                        <EmailDetail>
                            <Email/>
                        </EmailDetail>
                    </EmailList>
                </EntityOrganisationDetail>
            </EntityOrganisationList>
            <EntityAccountList>
                <EntityAccountDetail>
                    <AccID>20250818160027071</AccID>
                    <AccCurrency>HKD</AccCurrency>
                </EntityAccountDetail>
            </EntityAccountList>
            <EntityPhoneList>
                <EntityPhoneDetail>
                    <PhoneID>20250818160027070</PhoneID>
                    <PhoneNumDetail>
                        <SubscriberNumber/>
                    </PhoneNumDetail>
                </EntityPhoneDetail>
            </EntityPhoneList>
            <EntityAddressList>
                <EntityAddressDetail>
                    <AddressID>20250818160027070</AddressID>
                    <AddressDetail>
                        <AddressIndicator/>
                        <FreeAddress>
                            <Line1/>
                            <Line2/>
                            <Line3/>
                        </FreeAddress>
                        <FixedAddress>
                            <FlatNumber/>
                            <Floor/>
                            <BlockNumber/>
                            <BuildingEng/>
                            <BuildingChi/>
                            <EstVillage/>
                            <StreetNumFrom/>
                            <StreetNumTo/>
                            <StreetNameEng/>
                            <StreetNameChi/>
                            <District/>
                            <Area/>
                            <Country/>
                        </FixedAddress>
                    </AddressDetail>
                </EntityAddressDetail>
            </EntityAddressList>
            <EntityTrxList>
                <EntityTrxDetail>
                    <TrxID>1</TrxID>
                    <TrxDateTimeFrom>2025-08-18 00:00:00</TrxDateTimeFrom>
                    <TrxDateTimeTo>2025-08-19 00:00:00</TrxDateTimeTo>
                    <TrxSubject/>
                    <SubjectID>
                        <PersonID>20250818160027069</PersonID>
                        <OrgID/>
                        <AccID/>
                    </SubjectID>
                    <TrxType>TRW</TrxType>
                    <TrxAmount>-100.00</TrxAmount>
                    <TrxCurrency>HKD</TrxCurrency>
                    <Counterpart>
                        <CounterpartSubjectID>
                            <PersonID/>
                            <OrgID/>
                            <AccID/>
                            <CounterpartAccNum/>
                            <BeneficiaryOrPayer/>
                        </CounterpartSubjectID>
                    </Counterpart>
                </EntityTrxDetail>
            </EntityTrxList>
            <TrxTotalAmount>100.00</TrxTotalAmount>
            <TrxTotalPeriod>1.00000000</TrxTotalPeriod>
            <TrxDailyAverage>100.00</TrxDailyAverage>
            <SuspectedCrime>
                <SuspectedCrimeList>
                    <SuspectedCrimeDetail>
                        <ReasonCode>TERR</ReasonCode>
                    </SuspectedCrimeDetail>
                    <SuspectedCrimeDetail>
                        <ReasonCode>TRHM</ReasonCode>
                    </SuspectedCrimeDetail>
                </SuspectedCrimeList>
            </SuspectedCrime>
            <SuspiciousIndicator>
                <SuspiciousIndicatorList>
                    <SuspiciousIndicatorDetail>
                        <ReasonCode>ITBA</ReasonCode>
                    </SuspiciousIndicatorDetail>
                </SuspiciousIndicatorList>
            </SuspiciousIndicator>
            <OpenSourceInformation>
                <WebsiteList>
                    <WebsiteDetail>
                        <Website/>
                    </WebsiteDetail>
                </WebsiteList>
            </OpenSourceInformation>
            <AttachmentList>
                <AttachmentDetail>
                    <Sequence>1</Sequence>
                    <FileName/>
                    <FileSize/>
                    <FileID/>
                    <FileContent/>
                </AttachmentDetail>
            </AttachmentList>
        </STR>
    </xfa:data>
    <dd:dataDescription xmlns:dd="http://ns.adobe.com/data-description/" dd:name="STR">
        <STR>
            <VersionNo/>
            <STRNum dd:minOccur="0" dd:nullType="exclude"/>
            <SubmissionNum dd:minOccur="0" dd:nullType="exclude"/>
            <SubmissionDate dd:minOccur="0" dd:nullType="exclude"/>
            <AckDate dd:minOccur="0" dd:nullType="exclude"/>
            <ConsentDate dd:minOccur="0" dd:nullType="exclude"/>
            <ConsentFlag dd:minOccur="0" dd:nullType="exclude"/>
            <ConsentRemark dd:minOccur="0" dd:nullType="exclude"/>
            <mergeFormReload/>
            <CertExpiryDate dd:minOccur="0" dd:nullType="exclude"/>
            <STRHeader>
                <InstitutionID/>
                <RptOfficerID/>
                <InstitutionName dd:minOccur="0" dd:nullType="exclude"/>
                <RptOfficerName dd:minOccur="0" dd:nullType="exclude"/>
                <OrgReference dd:minOccur="0" dd:nullType="exclude"/>
                <PreTransaction dd:minOccur="0" dd:nullType="exclude"/>
                <Phone dd:minOccur="0" dd:nullType="exclude"/>
                <Fax dd:minOccur="0" dd:nullType="exclude"/>
                <Email dd:minOccur="0" dd:nullType="exclude"/>
                <Cap405 dd:minOccur="0" dd:nullType="exclude"/>
                <Cap455 dd:minOccur="0" dd:nullType="exclude"/>
                <Cap575 dd:minOccur="0" dd:nullType="exclude"/>
                <Cap000 dd:minOccur="0" dd:nullType="exclude"/>
                <UrgentCase dd:minOccur="0" dd:nullType="exclude"/>
                <isRelatedToExistingInvestigation/>
                <isRelatedToPreviousDisclosure dd:minOccur="0" dd:nullType="exclude"/>
                <JFIUNo dd:minOccur="0" dd:nullType="exclude"/>
                <UrRefNo dd:minOccur="0" dd:nullType="exclude"/>
                <ExistCaseRef dd:minOccur="0" dd:nullType="exclude"/>
                <InvUnit dd:minOccur="0" dd:nullType="exclude"/>
                <NumOfSubject/>
                <NumOfOrganization/>
                <NumOfAccount/>
                <NumOfPhone/>
                <NumOfAddress/>
                <NumOfTransaction/>
            </STRHeader>
            <EntityPersonList dd:minOccur="0">
                <EntityPersonDetail dd:maxOccur="-1" dd:minOccur="0">
                    <PersonID dd:minOccur="0" dd:nullType="exclude"/>
                    <FamilyName dd:minOccur="0" dd:nullType="exclude"/>
                    <GivenName dd:minOccur="0" dd:nullType="exclude"/>
                    <MiddleName dd:minOccur="0" dd:nullType="exclude"/>
                    <ChiName dd:minOccur="0">
                        <Name dd:minOccur="0" dd:nullType="exclude"/>
                        <ChineseCommercialCode dd:maxOccur="-1" dd:minOccur="0">
                            <FourDigitCode/>
                        </ChineseCommercialCode>
                    </ChiName>
                    <HKID dd:minOccur="0">
                        <Identifier/>
                        <CheckDigit dd:minOccur="0" dd:nullType="exclude"/>
                    </HKID>
                    <IdentityType dd:minOccur="0" dd:nullType="exclude"/>
                    <OtherIdentityTypeDescription dd:minOccur="0" dd:nullType="exclude"/>
                    <IdentityNum dd:minOccur="0" dd:nullType="exclude"/>
                    <Country dd:minOccur="0" dd:nullType="exclude"/>
                    <DOB dd:minOccur="0" dd:nullType="exclude"/>
                    <Sex dd:minOccur="0" dd:nullType="exclude"/>
                    <Occupation dd:minOccur="0" dd:nullType="exclude"/>
                    <Nature dd:minOccur="0" dd:nullType="exclude"/>
                    <OtherIdList dd:minOccur="0" dd:nullType="exclude">
                        <OtherIdDetail dd:maxOccur="-1" dd:minOccur="0" dd:nullType="exclude">
                            <IdentityType dd:minOccur="0" dd:nullType="exclude"/>
                            <OtherIdentityTypeDescription dd:minOccur="0" dd:nullType="exclude"/>
                            <IdentityNum dd:minOccur="0" dd:nullType="exclude"/>
                            <Country dd:minOccur="0" dd:nullType="exclude"/>
                        </OtherIdDetail>
                    </OtherIdList>
                    <EmailList>
                        <EmailDetail dd:maxOccur="-1" dd:minOccur="0">
                            <Email/>
                        </EmailDetail>
                    </EmailList>
                    <PhoneNumList dd:minOccur="0" dd:nullType="exclude">
                        <PhoneNumDetail dd:maxOccur="-1" dd:minOccur="0" dd:nullType="exclude">
                            <ChildID dd:minOccur="0" dd:nullType="exclude"/>
                            <ChildRole dd:minOccur="0" dd:nullType="exclude"/>
                            <OtherChildRoleDescription dd:minOccur="0" dd:nullType="exclude"/>
                        </PhoneNumDetail>
                    </PhoneNumList>
                    <AddressList dd:minOccur="0" dd:nullType="exclude">
                        <AddressDetail dd:maxOccur="-1" dd:minOccur="0" dd:nullType="exclude">
                            <ChildID dd:minOccur="0" dd:nullType="exclude"/>
                            <ChildRole dd:minOccur="0" dd:nullType="exclude"/>
                            <OtherChildRoleDescription dd:minOccur="0" dd:nullType="exclude"/>
                        </AddressDetail>
                    </AddressList>
                    <AddInfo dd:minOccur="0" dd:nullType="exclude"/>
                </EntityPersonDetail>
            </EntityPersonList>
            <EntityOrganisationList dd:minOccur="0">
                <EntityOrganisationDetail dd:maxOccur="-1" dd:minOccur="0">
                    <OrgID/>
                    <OrgEngName dd:minOccur="0" dd:nullType="exclude"/>
                    <OrgChiName dd:minOccur="0" dd:nullType="exclude"/>
                    <DateOfInc dd:minOccur="0" dd:nullType="exclude"/>
                    <LocalCompanyIndicator dd:minOccur="0" dd:nullType="exclude"/>
                    <LocalBusinessRegNumber dd:minOccur="0">
                        <BusinessRegistrationNumber/>
                        <BranchNumber dd:minOccur="0" dd:nullType="exclude"/>
                    </LocalBusinessRegNumber>
                    <LocalCompanyRegNumber dd:minOccur="0" dd:nullType="exclude"/>
                    <LocalPublicNumber dd:minOccur="0" dd:nullType="exclude"/>
                    <OverseasCompanyIndicator dd:minOccur="0" dd:nullType="exclude"/>
                    <OverseasCompanyCountry dd:minOccur="0" dd:nullType="exclude"/>
                    <OverseasRegNumber dd:minOccur="0" dd:nullType="exclude"/>
                    <NGOIndicator dd:minOccur="0" dd:nullType="exclude"/>
                    <CharityIndicator dd:minOccur="0" dd:nullType="exclude"/>
                    <Nature dd:minOccur="0" dd:nullType="exclude"/>
                    <BusinessNature dd:minOccur="0" dd:nullType="exclude"/>
                    <EmailList>
                        <EmailDetail dd:maxOccur="-1" dd:minOccur="0">
                            <Email/>
                        </EmailDetail>
                    </EmailList>
                    <PhoneNumList dd:minOccur="0" dd:nullType="exclude">
                        <PhoneNumDetail dd:maxOccur="-1" dd:minOccur="0" dd:nullType="exclude">
                            <ChildID dd:minOccur="0" dd:nullType="exclude"/>
                            <ChildRole dd:minOccur="0" dd:nullType="exclude"/>
                            <OtherChildRoleDescription dd:minOccur="0" dd:nullType="exclude"/>
                        </PhoneNumDetail>
                    </PhoneNumList>
                    <AddressList dd:minOccur="0" dd:nullType="exclude">
                        <AddressDetail dd:maxOccur="-1" dd:minOccur="0" dd:nullType="exclude">
                            <ChildID dd:minOccur="0" dd:nullType="exclude"/>
                            <ChildRole dd:minOccur="0" dd:nullType="exclude"/>
                            <OtherChildRoleDescription dd:minOccur="0" dd:nullType="exclude"/>
                        </AddressDetail>
                    </AddressList>
                    <PersonList dd:minOccur="0" dd:nullType="exclude">
                        <PersonDetail dd:maxOccur="-1" dd:minOccur="0" dd:nullType="exclude">
                            <ChildID dd:minOccur="0" dd:nullType="exclude"/>
                            <ChildRole dd:minOccur="0" dd:nullType="exclude"/>
                            <OtherChildRoleDescription dd:minOccur="0" dd:nullType="exclude"/>
                        </PersonDetail>
                    </PersonList>
                    <OrganisaionList dd:minOccur="0" dd:nullType="exclude">
                        <OrganisaionDetail dd:maxOccur="-1" dd:minOccur="0" dd:nullType="exclude">
                            <ChildID dd:minOccur="0" dd:nullType="exclude"/>
                            <ChildRole dd:minOccur="0" dd:nullType="exclude"/>
                            <OtherChildRoleDescription dd:minOccur="0" dd:nullType="exclude"/>
                        </OrganisaionDetail>
                    </OrganisaionList>
                    <AddInfo dd:minOccur="0" dd:nullType="exclude"/>
                </EntityOrganisationDetail>
            </EntityOrganisationList>
            <EntityAccountList dd:minOccur="0">
                <EntityAccountDetail dd:maxOccur="-1" dd:minOccur="0">
                    <AccID/>
                    <AccInstitution dd:minOccur="0" dd:nullType="exclude"/>
                    <AccNum dd:minOccur="0" dd:nullType="exclude"/>
                    <AccType dd:minOccur="0" dd:nullType="exclude"/>
                    <OtherAccTypeDescription dd:minOccur="0" dd:nullType="exclude"/>
                    <AccOpenDate dd:minOccur="0" dd:nullType="exclude"/>
                    <AccCloseDate dd:minOccur="0" dd:nullType="exclude"/>
                    <AccCurrency dd:minOccur="0" dd:nullType="exclude"/>
                    <AccBalance dd:minOccur="0" dd:nullType="exclude"/>
                    <AccBalanceDate dd:minOccur="0" dd:nullType="exclude"/>
                    <PersonList dd:minOccur="0" dd:nullType="exclude">
                        <PersonDetail dd:maxOccur="-1" dd:minOccur="0" dd:nullType="exclude">
                            <ChildID dd:minOccur="0" dd:nullType="exclude"/>
                            <ChildRole dd:minOccur="0" dd:nullType="exclude"/>
                            <OtherChildRoleDescription dd:minOccur="0" dd:nullType="exclude"/>
                        </PersonDetail>
                    </PersonList>
                    <OrganisaionList dd:minOccur="0" dd:nullType="exclude">
                        <OrganisaionDetail dd:maxOccur="-1" dd:minOccur="0" dd:nullType="exclude">
                            <ChildID dd:minOccur="0" dd:nullType="exclude"/>
                            <ChildRole dd:minOccur="0" dd:nullType="exclude"/>
                            <OtherChildRoleDescription dd:minOccur="0" dd:nullType="exclude"/>
                        </OrganisaionDetail>
                    </OrganisaionList>
                    <AddInfo dd:minOccur="0" dd:nullType="exclude"/>
                </EntityAccountDetail>
            </EntityAccountList>
            <EntityPhoneList dd:minOccur="0">
                <EntityPhoneDetail dd:maxOccur="-1" dd:minOccur="0">
                    <PhoneID/>
                    <PhoneNumDetail dd:maxOccur="-1" dd:minOccur="0">
                        <CountryCode dd:minOccur="0" dd:nullType="exclude"/>
                        <NationalDestinationCode dd:minOccur="0" dd:nullType="exclude"/>
                        <SubscriberNumber/>
                        <ExtensionNumber dd:minOccur="0" dd:nullType="exclude"/>
                    </PhoneNumDetail>
                </EntityPhoneDetail>
            </EntityPhoneList>
            <EntityAddressList dd:minOccur="0">
                <EntityAddressDetail dd:maxOccur="-1" dd:minOccur="0">
                    <AddressID/>
                    <AddressDetail dd:maxOccur="-1" dd:minOccur="0">
                        <AddressIndicator/>
                        <FreeAddress>
                            <Line1/>
                            <Line2/>
                            <Line3/>
                        </FreeAddress>
                        <FixedAddress>
                            <FlatNumber/>
                            <Floor/>
                            <BlockNumber/>
                            <BuildingEng/>
                            <BuildingChi/>
                            <EstVillage/>
                            <StreetNumFrom/>
                            <StreetNumTo/>
                            <StreetNameEng/>
                            <StreetNameChi/>
                            <District/>
                            <Area/>
                            <Country/>
                        </FixedAddress>
                    </AddressDetail>
                </EntityAddressDetail>
            </EntityAddressList>
            <EntityTrxList dd:minOccur="0">
                <EntityTrxDetail dd:maxOccur="-1" dd:minOccur="0">
                    <TrxID/>
                    <TrxDateTimeFrom/>
                    <TrxDateTimeTo/>
                    <TrxSubject/>
                    <SubjectID>
                        <PersonID/>
                        <OrgID/>
                        <AccID/>
                    </SubjectID>
                    <TrxType/>
                    <OtherTrxTypeDescription dd:minOccur="0" dd:nullType="exclude"/>
                    <TrxAmount/>
                    <TrxCurrency/>
                    <AccBalance dd:minOccur="0" dd:nullType="exclude"/>
                    <Branch dd:minOccur="0" dd:nullType="exclude"/>
                    <MatchingSeq dd:minOccur="0" dd:nullType="exclude"/>
                    <Counterpart>
                        <CounterpartSubjectID>
                            <PersonID/>
                            <OrgID/>
                            <AccID/>
                            <CounterpartAccNum/>
                            <BeneficiaryOrPayer/>
                        </CounterpartSubjectID>
                    </Counterpart>
                    <CounterpartDetail dd:minOccur="0" dd:nullType="exclude"/>
                </EntityTrxDetail>
            </EntityTrxList>
            <TrxTotalAmount dd:minOccur="0" dd:nullType="exclude"/>
            <TrxTotalPeriod dd:minOccur="0" dd:nullType="exclude"/>
            <TrxDailyAverage dd:minOccur="0" dd:nullType="exclude"/>
            <SuspectedCrime dd:minOccur="0" dd:nullType="exclude">
                <SuspectedCrimeList dd:minOccur="0" dd:nullType="exclude">
                    <SuspectedCrimeDetail dd:maxOccur="-1" dd:minOccur="0" dd:nullType="exclude">
                        <ReasonCode dd:minOccur="0" dd:nullType="exclude"/>
                        <OtherReasonDescription dd:minOccur="0" dd:nullType="exclude"/>
                    </SuspectedCrimeDetail>
                </SuspectedCrimeList>
                <AddInfo dd:minOccur="0" dd:nullType="exclude"/>
            </SuspectedCrime>
            <SuspiciousIndicator dd:minOccur="0" dd:nullType="exclude">
                <SuspiciousIndicatorList dd:minOccur="0" dd:nullType="exclude">
                    <SuspiciousIndicatorDetail dd:maxOccur="-1" dd:minOccur="0" dd:nullType="exclude">
                        <ReasonCode dd:minOccur="0" dd:nullType="exclude"/>
                        <OtherReasonDescription dd:minOccur="0" dd:nullType="exclude"/>
                    </SuspiciousIndicatorDetail>
                </SuspiciousIndicatorList>
                <AddInfo dd:minOccur="0" dd:nullType="exclude"/>
            </SuspiciousIndicator>
            <OpenSourceInformation dd:minOccur="0">
                <WebsiteList>
                    <WebsiteDetail dd:maxOccur="-1" dd:minOccur="0">
                        <Website/>
                    </WebsiteDetail>
                </WebsiteList>
                <AddInfo dd:minOccur="0" dd:nullType="exclude"/>
            </OpenSourceInformation>
            <AttachmentList>
                <AttachmentDetail dd:maxOccur="-1" dd:minOccur="0">
                    <Sequence/>
                    <FileName/>
                    <FileSize/>
                    <FileID/>
                    <FileContent/>
                </AttachmentDetail>
            </AttachmentList>
        </STR>
    </dd:dataDescription>
</xfa:datasets>`

	return []byte(s)
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
