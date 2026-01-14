// pkg/barcode/barcode.go

package barcode

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/ean"
)

var (
	ErrInvalidBarcodeFormat = errors.New("invalid barcode format")
	ErrInvalidBarcodeData   = errors.New("invalid barcode data")
)

type BarcodeFormat string

const (
	FormatEAN13   BarcodeFormat = "EAN13"
	FormatEAN8    BarcodeFormat = "EAN8"
	FormatCode128 BarcodeFormat = "CODE128"
)

type BarcodeGenerator struct {
	defaultFormat BarcodeFormat
	width         int
	height        int
}

func NewBarcodeGenerator() *BarcodeGenerator {
	return &BarcodeGenerator{
		defaultFormat: FormatEAN13,
		width:         200,
		height:        100,
	}
}

func (bg *BarcodeGenerator) Generate(data string, format BarcodeFormat) (barcode.Barcode, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return nil, ErrInvalidBarcodeData
	}

	var bc barcode.Barcode
	var err error

	switch format {
	case FormatEAN13:
		if len(data) != 13 && len(data) != 12 {
			return nil, errors.New("EAN13 requires 12 or 13 digits")
		}
		if len(data) == 12 {
			data = data + calculateEAN13Checksum(data)
		}
		bc, err = ean.Encode(data)

	case FormatEAN8:
		if len(data) != 8 && len(data) != 7 {
			return nil, errors.New("EAN8 requires 7 or 8 digits")
		}
		if len(data) == 7 {
			data = data + calculateEAN8Checksum(data)
		}
		bc, err = ean.Encode(data)

	case FormatCode128:
		bc, err = code128.Encode(data)

	default:
		return nil, ErrInvalidBarcodeFormat
	}

	if err != nil {
		return nil, err
	}

	bc, err = barcode.Scale(bc, bg.width, bg.height)
	if err != nil {
		return nil, err
	}

	return bc, nil
}

func (bg *BarcodeGenerator) GenerateWithDefaultFormat(data string) (barcode.Barcode, error) {
	return bg.Generate(data, bg.defaultFormat)
}

func (bg *BarcodeGenerator) GeneratePNG(data string, format BarcodeFormat) ([]byte, error) {
	bc, err := bg.Generate(data, format)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, bc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (bg *BarcodeGenerator) GenerateBase64(data string, format BarcodeFormat) (string, error) {
	pngData, err := bg.GeneratePNG(data, format)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(pngData), nil
}

func calculateEAN13Checksum(data string) string {
	if len(data) != 12 {
		return "0"
	}

	sum := 0
	for i, char := range data {
		digit := int(char - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}

	checksum := (10 - (sum % 10)) % 10
	return fmt.Sprintf("%d", checksum)
}

func calculateEAN8Checksum(data string) string {
	if len(data) != 7 {
		return "0"
	}

	sum := 0
	for i, char := range data {
		digit := int(char - '0')
		if i%2 == 0 {
			sum += digit * 3
		} else {
			sum += digit
		}
	}

	checksum := (10 - (sum % 10)) % 10
	return fmt.Sprintf("%d", checksum)
}

func (bg *BarcodeGenerator) ValidateEAN13(code string) bool {
	if len(code) != 13 {
		return false
	}

	for _, char := range code {
		if char < '0' || char > '9' {
			return false
		}
	}

	checksum := code[12:]
	data := code[:12]
	expectedChecksum := calculateEAN13Checksum(data)

	return checksum == expectedChecksum
}

func (bg *BarcodeGenerator) ValidateEAN8(code string) bool {
	if len(code) != 8 {
		return false
	}

	for _, char := range code {
		if char < '0' || char > '9' {
			return false
		}
	}

	checksum := code[7:]
	data := code[:7]
	expectedChecksum := calculateEAN8Checksum(data)

	return checksum == expectedChecksum
}

type LabelPrinter struct {
	labelWidth  int
	labelHeight int
	dpi         int
}

func NewLabelPrinter(widthMM, heightMM, dpi int) *LabelPrinter {
	return &LabelPrinter{
		labelWidth:  widthMM,
		labelHeight: heightMM,
		dpi:         dpi,
	}
}

func (lp *LabelPrinter) GenerateZPL(articleCode, description, barcode string, price float64) string {
	var zpl strings.Builder

	zpl.WriteString("^XA\n")
	zpl.WriteString("^FO20,20^A0N,30,30^FD" + articleCode + "^FS\n")

	if len(description) > 30 {
		description = description[:30]
	}
	zpl.WriteString("^FO20,60^A0N,20,20^FD" + description + "^FS\n")

	zpl.WriteString("^FO20,100^BY2^BCN,70,Y,N,N^FD" + barcode + "^FS\n")

	priceStr := fmt.Sprintf("EUR %.2f", price)
	zpl.WriteString("^FO20,190^A0N,25,25^FD" + priceStr + "^FS\n")

	zpl.WriteString("^XZ\n")

	return zpl.String()
}

func (lp *LabelPrinter) GenerateZPLBatch(labels []LabelData) string {
	var zpl strings.Builder

	for _, label := range labels {
		zpl.WriteString(lp.GenerateZPL(label.ArticleCode, label.Description, label.Barcode, label.Price))
	}

	return zpl.String()
}

func (lp *LabelPrinter) GenerateShelfLabel(articleCode, description, location string) string {
	var zpl strings.Builder

	zpl.WriteString("^XA\n")
	zpl.WriteString("^FO20,20^A0N,40,40^FD" + articleCode + "^FS\n")

	if len(description) > 25 {
		description = description[:25]
	}
	zpl.WriteString("^FO20,70^A0N,25,25^FD" + description + "^FS\n")

	zpl.WriteString("^FO20,110^A0N,30,30^FDLocation: " + location + "^FS\n")

	zpl.WriteString("^XZ\n")

	return zpl.String()
}

type LabelData struct {
	ArticleCode string
	Description string
	Barcode     string
	Price       float64
}

type BarcodeScanner struct {
	buffer []rune
}

func NewBarcodeScanner() *BarcodeScanner {
	return &BarcodeScanner{
		buffer: make([]rune, 0),
	}
}

func (bs *BarcodeScanner) ProcessInput(input rune) (string, bool) {
	if input == '\n' || input == '\r' {
		if len(bs.buffer) > 0 {
			result := string(bs.buffer)
			bs.buffer = bs.buffer[:0]
			return result, true
		}
		return "", false
	}

	bs.buffer = append(bs.buffer, input)
	return "", false
}

func (bs *BarcodeScanner) Reset() {
	bs.buffer = bs.buffer[:0]
}

func (bs *BarcodeScanner) GetBuffer() string {
	return string(bs.buffer)
}

type BarcodeValidator struct{}

func NewBarcodeValidator() *BarcodeValidator {
	return &BarcodeValidator{}
}

func (bv *BarcodeValidator) Validate(code string, format BarcodeFormat) (bool, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return false, ErrInvalidBarcodeData
	}

	switch format {
	case FormatEAN13:
		if len(code) != 13 {
			return false, errors.New("EAN13 must be 13 digits")
		}
		for _, char := range code {
			if char < '0' || char > '9' {
				return false, errors.New("EAN13 must contain only digits")
			}
		}
		checksum := code[12:]
		data := code[:12]
		expectedChecksum := calculateEAN13Checksum(data)
		if checksum != expectedChecksum {
			return false, errors.New("invalid EAN13 checksum")
		}
		return true, nil

	case FormatEAN8:
		if len(code) != 8 {
			return false, errors.New("EAN8 must be 8 digits")
		}
		for _, char := range code {
			if char < '0' || char > '9' {
				return false, errors.New("EAN8 must contain only digits")
			}
		}
		checksum := code[7:]
		data := code[:7]
		expectedChecksum := calculateEAN8Checksum(data)
		if checksum != expectedChecksum {
			return false, errors.New("invalid EAN8 checksum")
		}
		return true, nil

	case FormatCode128:
		if len(code) == 0 {
			return false, errors.New("CODE128 cannot be empty")
		}
		return true, nil

	default:
		return false, ErrInvalidBarcodeFormat
	}
}

func (bv *BarcodeValidator) DetectFormat(code string) BarcodeFormat {
	code = strings.TrimSpace(code)

	if len(code) == 13 {
		isNumeric := true
		for _, char := range code {
			if char < '0' || char > '9' {
				isNumeric = false
				break
			}
		}
		if isNumeric {
			return FormatEAN13
		}
	}

	if len(code) == 8 {
		isNumeric := true
		for _, char := range code {
			if char < '0' || char > '9' {
				isNumeric = false
				break
			}
		}
		if isNumeric {
			return FormatEAN8
		}
	}

	return FormatCode128
}

func (bv *BarcodeValidator) SuggestCorrection(code string, format BarcodeFormat) (string, error) {
	code = strings.TrimSpace(code)

	switch format {
	case FormatEAN13:
		if len(code) == 12 {
			return code + calculateEAN13Checksum(code), nil
		}
		if len(code) == 13 {
			data := code[:12]
			return data + calculateEAN13Checksum(data), nil
		}
		return "", errors.New("cannot suggest correction for this code length")

	case FormatEAN8:
		if len(code) == 7 {
			return code + calculateEAN8Checksum(code), nil
		}
		if len(code) == 8 {
			data := code[:7]
			return data + calculateEAN8Checksum(data), nil
		}
		return "", errors.New("cannot suggest correction for this code length")

	default:
		return code, nil
	}
}
