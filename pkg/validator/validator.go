// pkg/validator/validator.go

package validator

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrInvalidVAT        = errors.New("invalid VAT number")
	ErrInvalidFiscalCode = errors.New("invalid fiscal code")
	ErrInvalidIBAN       = errors.New("invalid IBAN")
	ErrInvalidPhone      = errors.New("invalid phone number")
	ErrInvalidPostalCode = errors.New("invalid postal code")
)

type Validator struct {
	emailRegex      *regexp.Regexp
	phoneRegex      *regexp.Regexp
	postalCodeRegex *regexp.Regexp
}

func NewValidator() *Validator {
	return &Validator{
		emailRegex:      regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`),
		phoneRegex:      regexp.MustCompile(`^\+?[0-9]{6,15}$`),
		postalCodeRegex: regexp.MustCompile(`^[0-9]{5}$`),
	}
}

func (v *Validator) ValidateEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil
	}

	if !v.emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

func (v *Validator) ValidatePhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil
	}

	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	if !v.phoneRegex.MatchString(phone) {
		return ErrInvalidPhone
	}

	return nil
}

func (v *Validator) ValidateItalianPostalCode(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil
	}

	if !v.postalCodeRegex.MatchString(code) {
		return ErrInvalidPostalCode
	}

	return nil
}

func (v *Validator) ValidateItalianVAT(vat string) error {
	vat = strings.TrimSpace(vat)
	if vat == "" {
		return nil
	}

	vat = strings.ReplaceAll(vat, " ", "")
	vat = strings.ToUpper(vat)

	if strings.HasPrefix(vat, "IT") {
		vat = vat[2:]
	}

	if len(vat) != 11 {
		return ErrInvalidVAT
	}

	for _, char := range vat {
		if char < '0' || char > '9' {
			return ErrInvalidVAT
		}
	}

	if !v.validateItalianVATChecksum(vat) {
		return ErrInvalidVAT
	}

	return nil
}

func (v *Validator) validateItalianVATChecksum(vat string) bool {
	if len(vat) != 11 {
		return false
	}

	sum := 0
	for i := 0; i < 10; i++ {
		digit := int(vat[i] - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			product := digit * 2
			if product > 9 {
				product = product/10 + product%10
			}
			sum += product
		}
	}

	checkDigit := (10 - (sum % 10)) % 10
	return int(vat[10]-'0') == checkDigit
}

func (v *Validator) ValidateItalianFiscalCode(code string) error {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == "" {
		return nil
	}

	if len(code) != 16 {
		return ErrInvalidFiscalCode
	}

	for i := 0; i < 6; i++ {
		if code[i] < 'A' || code[i] > 'Z' {
			return ErrInvalidFiscalCode
		}
	}

	for i := 6; i < 8; i++ {
		if code[i] < '0' || code[i] > '9' {
			return ErrInvalidFiscalCode
		}
	}

	if code[8] < 'A' || code[8] > 'Z' {
		return ErrInvalidFiscalCode
	}

	for i := 9; i < 11; i++ {
		if code[i] < '0' || code[i] > '9' {
			return ErrInvalidFiscalCode
		}
	}

	if code[11] < 'A' || code[11] > 'Z' {
		return ErrInvalidFiscalCode
	}

	for i := 12; i < 15; i++ {
		if code[i] < '0' || code[i] > '9' {
			return ErrInvalidFiscalCode
		}
	}

	if code[15] < 'A' || code[15] > 'Z' {
		return ErrInvalidFiscalCode
	}

	return nil
}

func (v *Validator) ValidateIBAN(iban string) error {
	iban = strings.TrimSpace(strings.ToUpper(iban))
	if iban == "" {
		return nil
	}

	iban = strings.ReplaceAll(iban, " ", "")

	if len(iban) < 15 || len(iban) > 34 {
		return ErrInvalidIBAN
	}

	countryCode := iban[0:2]
	if countryCode[0] < 'A' || countryCode[0] > 'Z' || countryCode[1] < 'A' || countryCode[1] > 'Z' {
		return ErrInvalidIBAN
	}

	checkDigits := iban[2:4]
	if checkDigits[0] < '0' || checkDigits[0] > '9' || checkDigits[1] < '0' || checkDigits[1] > '9' {
		return ErrInvalidIBAN
	}

	if countryCode == "IT" && len(iban) != 27 {
		return ErrInvalidIBAN
	}

	return nil
}

func (v *Validator) ValidateSDI(sdi string) error {
	sdi = strings.TrimSpace(strings.ToUpper(sdi))
	if sdi == "" {
		return nil
	}

	if len(sdi) != 7 {
		return errors.New("SDI code must be 7 characters")
	}

	for _, char := range sdi {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return errors.New("SDI code must contain only letters and numbers")
		}
	}

	return nil
}

func (v *Validator) ValidatePEC(pec string) error {
	pec = strings.TrimSpace(strings.ToLower(pec))
	if pec == "" {
		return nil
	}

	if err := v.ValidateEmail(pec); err != nil {
		return errors.New("invalid PEC format")
	}

	validDomains := []string{"pec.it", "legalmail.it", "pecimpresa.it"}
	hasPECDomain := false
	for _, domain := range validDomains {
		if strings.HasSuffix(pec, domain) {
			hasPECDomain = true
			break
		}
	}

	if !hasPECDomain {
		return errors.New("PEC must use a certified email domain")
	}

	return nil
}

type BusinessValidator struct {
	validator *Validator
}

func NewBusinessValidator() *BusinessValidator {
	return &BusinessValidator{
		validator: NewValidator(),
	}
}

func (bv *BusinessValidator) ValidatePrice(price float64) error {
	if price < 0 {
		return errors.New("price cannot be negative")
	}
	if price > 999999.99 {
		return errors.New("price exceeds maximum allowed value")
	}
	return nil
}

func (bv *BusinessValidator) ValidateQuantity(quantity float64) error {
	if quantity < 0 {
		return errors.New("quantity cannot be negative")
	}
	if quantity > 999999 {
		return errors.New("quantity exceeds maximum allowed value")
	}
	return nil
}

func (bv *BusinessValidator) ValidateDiscount(discount float64) error {
	if discount < 0 {
		return errors.New("discount cannot be negative")
	}
	if discount > 100 {
		return errors.New("discount cannot exceed 100%")
	}
	return nil
}

func (bv *BusinessValidator) ValidateDiscountCascade(discounts []float64) error {
	if len(discounts) == 0 {
		return errors.New("discount cascade cannot be empty")
	}

	for i, discount := range discounts {
		if err := bv.ValidateDiscount(discount); err != nil {
			return errors.New("invalid discount at position " + string(rune(i+1)) + ": " + err.Error())
		}
	}

	totalDiscount := 1.0
	for _, discount := range discounts {
		totalDiscount *= (1 - discount/100)
	}

	effectiveDiscount := (1 - totalDiscount) * 100
	if effectiveDiscount > 99 {
		return errors.New("total discount cascade exceeds 99%")
	}

	return nil
}

func (bv *BusinessValidator) ValidateArticleCode(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return errors.New("article code cannot be empty")
	}
	if len(code) > 50 {
		return errors.New("article code too long (max 50 characters)")
	}

	validChars := regexp.MustCompile(`^[A-Z0-9\-_]+$`)
	if !validChars.MatchString(strings.ToUpper(code)) {
		return errors.New("article code contains invalid characters")
	}

	return nil
}

func (bv *BusinessValidator) ValidateCustomerCode(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return errors.New("customer code cannot be empty")
	}
	if len(code) > 20 {
		return errors.New("customer code too long (max 20 characters)")
	}

	validChars := regexp.MustCompile(`^[A-Z0-9]+$`)
	if !validChars.MatchString(strings.ToUpper(code)) {
		return errors.New("customer code contains invalid characters")
	}

	return nil
}

func (bv *BusinessValidator) ValidateSupplierCode(code string) error {
	return bv.ValidateCustomerCode(code)
}

func (bv *BusinessValidator) ValidateDateRange(startDate, endDate string) error {
	if startDate == "" || endDate == "" {
		return errors.New("date range cannot be empty")
	}

	return nil
}

func (bv *BusinessValidator) ValidateFidoLimit(limit float64) error {
	if limit < 0 {
		return errors.New("fido limit cannot be negative")
	}
	if limit > 9999999.99 {
		return errors.New("fido limit exceeds maximum allowed value")
	}
	return nil
}

func (bv *BusinessValidator) ValidateCreditAmount(amount float64) error {
	if amount <= 0 {
		return errors.New("credit amount must be positive")
	}
	if amount > 999999.99 {
		return errors.New("credit amount exceeds maximum allowed value")
	}
	return nil
}

type InputSanitizer struct{}

func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{}
}

func (is *InputSanitizer) SanitizeString(input string) string {
	input = strings.TrimSpace(input)

	replacements := map[string]string{
		"<":  "&lt;",
		">":  "&gt;",
		"&":  "&amp;",
		"\"": "&quot;",
		"'":  "&#39;",
	}

	for old, new := range replacements {
		input = strings.ReplaceAll(input, old, new)
	}

	return input
}

func (is *InputSanitizer) SanitizeArticleCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	validChars := regexp.MustCompile(`[^A-Z0-9\-_]`)
	return validChars.ReplaceAllString(code, "")
}

func (is *InputSanitizer) SanitizeNumericString(input string) string {
	validChars := regexp.MustCompile(`[^0-9.]`)
	return validChars.ReplaceAllString(strings.TrimSpace(input), "")
}

func (is *InputSanitizer) SanitizeAlphanumeric(input string) string {
	validChars := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return validChars.ReplaceAllString(strings.TrimSpace(input), "")
}

func (is *InputSanitizer) TruncateString(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}
	return input[:maxLength]
}

func (is *InputSanitizer) NormalizeWhitespace(input string) string {
	whitespace := regexp.MustCompile(`\s+`)
	input = whitespace.ReplaceAllString(input, " ")
	return strings.TrimSpace(input)
}
