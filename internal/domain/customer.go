// internal/domain/customer.go

package domain

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrInvalidCustomerData = errors.New("invalid customer data")
	ErrFidoExceeded        = errors.New("fido limit exceeded")
	ErrInvalidDiscountRule = errors.New("invalid discount rule")
)

type CustomerCategory string

const (
	CategoryRetail    CustomerCategory = "retail"
	CategoryWholesale CustomerCategory = "wholesale"
	CategoryWorkshop  CustomerCategory = "workshop"
	CategoryDealer    CustomerCategory = "dealer"
	CategoryVIP       CustomerCategory = "vip"
)

type CreditClass string

const (
	CreditClassA CreditClass = "A"
	CreditClassB CreditClass = "B"
	CreditClassC CreditClass = "C"
	CreditClassD CreditClass = "D"
	CreditClassE CreditClass = "E"
)

type Customer struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Code            string               `bson:"code" json:"code"`
	CompanyName     string               `bson:"company_name" json:"company_name"`
	VATNumber       string               `bson:"vat_number" json:"vat_number"`
	FiscalCode      string               `bson:"fiscal_code" json:"fiscal_code"`
	Category        CustomerCategory     `bson:"category" json:"category"`
	CreditInfo      CreditInfo           `bson:"credit_info" json:"credit_info"`
	ContactInfo     ContactInfo          `bson:"contact_info" json:"contact_info"`
	BillingAddress  Address              `bson:"billing_address" json:"billing_address"`
	ShippingAddress Address              `bson:"shipping_address" json:"shipping_address"`
	DiscountGrid    []DiscountRule       `bson:"discount_grid" json:"discount_grid"`
	Budgets         []primitive.ObjectID `bson:"budgets" json:"budgets"`
	CreditVouchers  []primitive.ObjectID `bson:"credit_vouchers" json:"credit_vouchers"`
	PaymentTerms    PaymentTerms         `bson:"payment_terms" json:"payment_terms"`
	PriceList       string               `bson:"price_list" json:"price_list"`
	IsActive        bool                 `bson:"is_active" json:"is_active"`
	Notes           string               `bson:"notes" json:"notes"`
	Tags            []string             `bson:"tags" json:"tags"`
	CreatedAt       time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time            `bson:"updated_at" json:"updated_at"`
	CreatedBy       string               `bson:"created_by" json:"created_by"`
	UpdatedBy       string               `bson:"updated_by" json:"updated_by"`
}

type CreditInfo struct {
	CreditClass     CreditClass `bson:"credit_class" json:"credit_class"`
	FidoLimit       float64     `bson:"fido_limit" json:"fido_limit"`
	CurrentExposure float64     `bson:"current_exposure" json:"current_exposure"`
	OpenOrders      float64     `bson:"open_orders" json:"open_orders"`
	UnpaidInvoices  float64     `bson:"unpaid_invoices" json:"unpaid_invoices"`
	OverdueAmount   float64     `bson:"overdue_amount" json:"overdue_amount"`
	LastCreditCheck time.Time   `bson:"last_credit_check" json:"last_credit_check"`
	BlockSales      bool        `bson:"block_sales" json:"block_sales"`
	BlockReason     string      `bson:"block_reason" json:"block_reason"`
}

type ContactInfo struct {
	PrimaryContact string `bson:"primary_contact" json:"primary_contact"`
	Phone          string `bson:"phone" json:"phone"`
	Mobile         string `bson:"mobile" json:"mobile"`
	Email          string `bson:"email" json:"email"`
	PEC            string `bson:"pec" json:"pec"`
	Website        string `bson:"website" json:"website"`
	SDICode        string `bson:"sdi_code" json:"sdi_code"`
}

type Address struct {
	Street     string `bson:"street" json:"street"`
	City       string `bson:"city" json:"city"`
	Province   string `bson:"province" json:"province"`
	PostalCode string `bson:"postal_code" json:"postal_code"`
	Country    string `bson:"country" json:"country"`
}

type DiscountRule struct {
	ID              primitive.ObjectID `bson:"id" json:"id"`
	Priority        int                `bson:"priority" json:"priority"`
	Precodice       string             `bson:"precodice" json:"precodice"`
	Family          string             `bson:"family" json:"family"`
	Classification  string             `bson:"classification" json:"classification"`
	ArticleCode     string             `bson:"article_code" json:"article_code"`
	DiscountPercent float64            `bson:"discount_percent" json:"discount_percent"`
	DiscountCascade []float64          `bson:"discount_cascade" json:"discount_cascade"`
	ValidFrom       time.Time          `bson:"valid_from" json:"valid_from"`
	ValidTo         time.Time          `bson:"valid_to" json:"valid_to"`
	MinQuantity     float64            `bson:"min_quantity" json:"min_quantity"`
	IsActive        bool               `bson:"is_active" json:"is_active"`
}

type PaymentTerms struct {
	Method          string  `bson:"method" json:"method"`
	DaysNet         int     `bson:"days_net" json:"days_net"`
	DaysEndMonth    bool    `bson:"days_end_month" json:"days_end_month"`
	CashDiscount    float64 `bson:"cash_discount" json:"cash_discount"`
	InstallmentPlan bool    `bson:"installment_plan" json:"installment_plan"`
}

func NewCustomer(code, companyName, createdBy string) (*Customer, error) {
	if strings.TrimSpace(code) == "" {
		return nil, errors.New("customer code cannot be empty")
	}
	if strings.TrimSpace(companyName) == "" {
		return nil, errors.New("company name cannot be empty")
	}

	now := time.Now()
	return &Customer{
		ID:          primitive.NewObjectID(),
		Code:        strings.ToUpper(strings.TrimSpace(code)),
		CompanyName: strings.TrimSpace(companyName),
		Category:    CategoryRetail,
		CreditInfo: CreditInfo{
			CreditClass:     CreditClassC,
			FidoLimit:       5000.0,
			CurrentExposure: 0,
			BlockSales:      false,
		},
		ContactInfo:     ContactInfo{},
		BillingAddress:  Address{Country: "IT"},
		ShippingAddress: Address{Country: "IT"},
		DiscountGrid:    []DiscountRule{},
		Budgets:         []primitive.ObjectID{},
		CreditVouchers:  []primitive.ObjectID{},
		PaymentTerms:    PaymentTerms{Method: "bank_transfer", DaysNet: 30},
		PriceList:       "standard",
		IsActive:        true,
		Tags:            []string{},
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       createdBy,
		UpdatedBy:       createdBy,
	}, nil
}

func (c *Customer) Validate() error {
	if strings.TrimSpace(c.Code) == "" {
		return errors.New("customer code cannot be empty")
	}
	if strings.TrimSpace(c.CompanyName) == "" {
		return errors.New("company name cannot be empty")
	}
	if c.CreditInfo.FidoLimit < 0 {
		return errors.New("fido limit cannot be negative")
	}
	return nil
}

func (c *Customer) UpdateExposure(unpaidInvoices, openOrders float64) {
	c.CreditInfo.UnpaidInvoices = unpaidInvoices
	c.CreditInfo.OpenOrders = openOrders
	c.CreditInfo.CurrentExposure = unpaidInvoices + openOrders
	c.CreditInfo.LastCreditCheck = time.Now()
	c.UpdatedAt = time.Now()
}

func (c *Customer) GetFidoUsagePercent() float64 {
	if c.CreditInfo.FidoLimit == 0 {
		return 0
	}
	return (c.CreditInfo.CurrentExposure / c.CreditInfo.FidoLimit) * 100
}

func (c *Customer) IsFidoWarning(warningThreshold float64) bool {
	usage := c.GetFidoUsagePercent()
	return usage >= warningThreshold && usage < 100
}

func (c *Customer) IsFidoBlocked(blockThreshold float64) bool {
	usage := c.GetFidoUsagePercent()
	return usage >= blockThreshold || c.CreditInfo.BlockSales
}

func (c *Customer) CanMakePurchase(amount float64, warningThreshold, blockThreshold float64) (bool, string) {
	if !c.IsActive {
		return false, "customer is not active"
	}

	if c.CreditInfo.BlockSales {
		return false, c.CreditInfo.BlockReason
	}

	newExposure := c.CreditInfo.CurrentExposure + amount
	newUsage := (newExposure / c.CreditInfo.FidoLimit) * 100

	if newUsage >= blockThreshold {
		return false, "purchase would exceed fido limit"
	}

	if newUsage >= warningThreshold {
		return true, "warning: approaching fido limit"
	}

	return true, ""
}

func (c *Customer) BlockSales(reason string) {
	c.CreditInfo.BlockSales = true
	c.CreditInfo.BlockReason = reason
	c.UpdatedAt = time.Now()
}

func (c *Customer) UnblockSales() {
	c.CreditInfo.BlockSales = false
	c.CreditInfo.BlockReason = ""
	c.UpdatedAt = time.Now()
}

func (c *Customer) AddDiscountRule(rule DiscountRule) error {
	if rule.DiscountPercent < 0 || rule.DiscountPercent > 100 {
		return ErrInvalidDiscountRule
	}

	rule.ID = primitive.NewObjectID()
	rule.IsActive = true

	c.DiscountGrid = append(c.DiscountGrid, rule)
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Customer) RemoveDiscountRule(ruleID primitive.ObjectID) {
	for i, rule := range c.DiscountGrid {
		if rule.ID == ruleID {
			c.DiscountGrid = append(c.DiscountGrid[:i], c.DiscountGrid[i+1:]...)
			c.UpdatedAt = time.Now()
			return
		}
	}
}

func (c *Customer) GetApplicableDiscount(article *Article, quantity float64) *DiscountRule {
	now := time.Now()
	var bestRule *DiscountRule

	for i := range c.DiscountGrid {
		rule := &c.DiscountGrid[i]

		if !rule.IsActive {
			continue
		}

		if !rule.ValidFrom.IsZero() && now.Before(rule.ValidFrom) {
			continue
		}

		if !rule.ValidTo.IsZero() && now.After(rule.ValidTo) {
			continue
		}

		if rule.MinQuantity > 0 && quantity < rule.MinQuantity {
			continue
		}

		matches := false

		if rule.ArticleCode != "" && rule.ArticleCode == article.Code {
			matches = true
		} else if rule.Precodice != "" && rule.Precodice == article.Precodice {
			matches = true
		} else if rule.Family != "" && rule.Family == article.Family {
			matches = true
		} else if rule.Classification != "" {
			for _, cls := range article.Classification {
				if cls == rule.Classification {
					matches = true
					break
				}
			}
		}

		if matches {
			if bestRule == nil || rule.Priority > bestRule.Priority {
				bestRule = rule
			}
		}
	}

	return bestRule
}

func (c *Customer) CalculateFinalPrice(article *Article, quantity float64, basePrice float64) float64 {
	rule := c.GetApplicableDiscount(article, quantity)
	if rule == nil {
		return basePrice
	}

	finalPrice := basePrice

	if len(rule.DiscountCascade) > 0 {
		for _, discount := range rule.DiscountCascade {
			finalPrice = finalPrice * (1 - discount/100)
		}
	} else {
		finalPrice = basePrice * (1 - rule.DiscountPercent/100)
	}

	return finalPrice
}

func (c *Customer) GetAvailableFido() float64 {
	available := c.CreditInfo.FidoLimit - c.CreditInfo.CurrentExposure
	if available < 0 {
		return 0
	}
	return available
}

func (c *Customer) HasOverduePayments() bool {
	return c.CreditInfo.OverdueAmount > 0
}

func (c *Customer) SetCreditClass(class CreditClass, fidoLimit float64) {
	c.CreditInfo.CreditClass = class
	c.CreditInfo.FidoLimit = fidoLimit
	c.UpdatedAt = time.Now()
}

func (c *Customer) AddTag(tag string) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return
	}

	for _, t := range c.Tags {
		if t == tag {
			return
		}
	}

	c.Tags = append(c.Tags, tag)
	c.UpdatedAt = time.Now()
}

func (c *Customer) RemoveTag(tag string) {
	for i, t := range c.Tags {
		if t == tag {
			c.Tags = append(c.Tags[:i], c.Tags[i+1:]...)
			c.UpdatedAt = time.Now()
			return
		}
	}
}

func (c *Customer) IsVIP() bool {
	return c.Category == CategoryVIP
}

func (c *Customer) GetCreditRating() string {
	switch c.CreditInfo.CreditClass {
	case CreditClassA:
		return "Excellent"
	case CreditClassB:
		return "Good"
	case CreditClassC:
		return "Average"
	case CreditClassD:
		return "Poor"
	case CreditClassE:
		return "High Risk"
	default:
		return "Unknown"
	}
}
