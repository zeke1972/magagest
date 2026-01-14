// internal/domain/promotion.go

package domain

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrPromotionNotFound    = errors.New("promotion not found")
	ErrPromotionExpired     = errors.New("promotion expired")
	ErrPromotionNotActive   = errors.New("promotion not active")
	ErrInvalidPromotionRule = errors.New("invalid promotion rule")
)

type PromotionType string

const (
	PromotionTypePercentDiscount PromotionType = "percent_discount"
	PromotionTypeFixedPrice      PromotionType = "fixed_price"
	PromotionTypeNxM             PromotionType = "nxm"
	PromotionTypeBundle          PromotionType = "bundle"
	PromotionTypeFreeShipping    PromotionType = "free_shipping"
)

type Promotion struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Code          string              `bson:"code" json:"code"`
	Name          string              `bson:"name" json:"name"`
	Description   string              `bson:"description" json:"description"`
	Type          PromotionType       `bson:"type" json:"type"`
	IsActive      bool                `bson:"is_active" json:"is_active"`
	ValidFrom     time.Time           `bson:"valid_from" json:"valid_from"`
	ValidTo       time.Time           `bson:"valid_to" json:"valid_to"`
	Priority      int                 `bson:"priority" json:"priority"`
	Rules         PromotionRules      `bson:"rules" json:"rules"`
	Applicability ApplicabilityRules  `bson:"applicability" json:"applicability"`
	Conditions    PromotionConditions `bson:"conditions" json:"conditions"`
	Limits        PromotionLimits     `bson:"limits" json:"limits"`
	Statistics    PromotionStats      `bson:"statistics" json:"statistics"`
	CreatedAt     time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time           `bson:"updated_at" json:"updated_at"`
	CreatedBy     string              `bson:"created_by" json:"created_by"`
	UpdatedBy     string              `bson:"updated_by" json:"updated_by"`
}

type PromotionRules struct {
	DiscountPercent float64         `bson:"discount_percent" json:"discount_percent"`
	FixedPrice      float64         `bson:"fixed_price" json:"fixed_price"`
	BuyQuantity     int             `bson:"buy_quantity" json:"buy_quantity"`
	GetQuantity     int             `bson:"get_quantity" json:"get_quantity"`
	BundleArticles  []BundleArticle `bson:"bundle_articles" json:"bundle_articles"`
	BundlePrice     float64         `bson:"bundle_price" json:"bundle_price"`
}

type BundleArticle struct {
	ArticleID   primitive.ObjectID `bson:"article_id" json:"article_id"`
	ArticleCode string             `bson:"article_code" json:"article_code"`
	Quantity    float64            `bson:"quantity" json:"quantity"`
	IsRequired  bool               `bson:"is_required" json:"is_required"`
}

type ApplicabilityRules struct {
	ArticleCodes       []string             `bson:"article_codes" json:"article_codes"`
	Precodici          []string             `bson:"precodici" json:"precodici"`
	Families           []string             `bson:"families" json:"families"`
	Classifications    []string             `bson:"classifications" json:"classifications"`
	Categories         []string             `bson:"categories" json:"categories"`
	CustomerCategories []CustomerCategory   `bson:"customer_categories" json:"customer_categories"`
	SpecificCustomers  []primitive.ObjectID `bson:"specific_customers" json:"specific_customers"`
	ExcludedArticles   []string             `bson:"excluded_articles" json:"excluded_articles"`
}

type PromotionConditions struct {
	MinQuantity       float64 `bson:"min_quantity" json:"min_quantity"`
	MaxQuantity       float64 `bson:"max_quantity" json:"max_quantity"`
	MinAmount         float64 `bson:"min_amount" json:"min_amount"`
	MaxAmount         float64 `bson:"max_amount" json:"max_amount"`
	RequiresCoupon    bool    `bson:"requires_coupon" json:"requires_coupon"`
	CouponCode        string  `bson:"coupon_code" json:"coupon_code"`
	CombineWithOthers bool    `bson:"combine_with_others" json:"combine_with_others"`
}

type PromotionLimits struct {
	MaxUsageTotal       int `bson:"max_usage_total" json:"max_usage_total"`
	MaxUsagePerCustomer int `bson:"max_usage_per_customer" json:"max_usage_per_customer"`
	MaxUsagePerDay      int `bson:"max_usage_per_day" json:"max_usage_per_day"`
}

type PromotionStats struct {
	TotalUsages    int            `bson:"total_usages" json:"total_usages"`
	UsagesToday    int            `bson:"usages_today" json:"usages_today"`
	LastUsageDate  time.Time      `bson:"last_usage_date" json:"last_usage_date"`
	CustomerUsages map[string]int `bson:"customer_usages" json:"customer_usages"`
	TotalRevenue   float64        `bson:"total_revenue" json:"total_revenue"`
	TotalDiscount  float64        `bson:"total_discount" json:"total_discount"`
}

func NewPromotion(code, name string, promotionType PromotionType, validFrom, validTo time.Time, createdBy string) (*Promotion, error) {
	if strings.TrimSpace(code) == "" {
		return nil, errors.New("promotion code cannot be empty")
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("promotion name cannot be empty")
	}
	if validFrom.After(validTo) {
		return nil, errors.New("valid_from must be before valid_to")
	}

	now := time.Now()
	return &Promotion{
		ID:        primitive.NewObjectID(),
		Code:      strings.ToUpper(strings.TrimSpace(code)),
		Name:      strings.TrimSpace(name),
		Type:      promotionType,
		IsActive:  true,
		ValidFrom: validFrom,
		ValidTo:   validTo,
		Priority:  0,
		Rules:     PromotionRules{},
		Applicability: ApplicabilityRules{
			ArticleCodes:       []string{},
			Precodici:          []string{},
			Families:           []string{},
			Classifications:    []string{},
			Categories:         []string{},
			CustomerCategories: []CustomerCategory{},
			SpecificCustomers:  []primitive.ObjectID{},
			ExcludedArticles:   []string{},
		},
		Conditions: PromotionConditions{
			CombineWithOthers: false,
		},
		Limits: PromotionLimits{},
		Statistics: PromotionStats{
			CustomerUsages: make(map[string]int),
		},
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}, nil
}

func (p *Promotion) Validate() error {
	if strings.TrimSpace(p.Code) == "" {
		return errors.New("promotion code cannot be empty")
	}
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("promotion name cannot be empty")
	}
	if p.ValidFrom.After(p.ValidTo) {
		return errors.New("valid_from must be before valid_to")
	}

	switch p.Type {
	case PromotionTypePercentDiscount:
		if p.Rules.DiscountPercent <= 0 || p.Rules.DiscountPercent > 100 {
			return ErrInvalidPromotionRule
		}
	case PromotionTypeFixedPrice:
		if p.Rules.FixedPrice <= 0 {
			return ErrInvalidPromotionRule
		}
	case PromotionTypeNxM:
		if p.Rules.BuyQuantity <= 0 || p.Rules.GetQuantity < 0 {
			return ErrInvalidPromotionRule
		}
	case PromotionTypeBundle:
		if len(p.Rules.BundleArticles) < 2 {
			return errors.New("bundle must contain at least 2 articles")
		}
	}

	return nil
}

func (p *Promotion) IsValid(now time.Time) bool {
	if !p.IsActive {
		return false
	}
	if now.Before(p.ValidFrom) {
		return false
	}
	if !p.ValidTo.IsZero() && now.After(p.ValidTo) {
		return false
	}
	return true
}

func (p *Promotion) IsExpired(now time.Time) bool {
	if p.ValidTo.IsZero() {
		return false
	}
	return now.After(p.ValidTo)
}

func (p *Promotion) IsApplicableToArticle(article *Article) bool {
	if len(p.Applicability.ExcludedArticles) > 0 {
		for _, excluded := range p.Applicability.ExcludedArticles {
			if excluded == article.Code {
				return false
			}
		}
	}

	if len(p.Applicability.ArticleCodes) > 0 {
		for _, code := range p.Applicability.ArticleCodes {
			if code == article.Code {
				return true
			}
		}
		return false
	}

	if len(p.Applicability.Precodici) > 0 {
		for _, precodice := range p.Applicability.Precodici {
			if precodice == article.Precodice {
				return true
			}
		}
		return false
	}

	if len(p.Applicability.Families) > 0 {
		for _, family := range p.Applicability.Families {
			if family == article.Family {
				return true
			}
		}
		return false
	}

	if len(p.Applicability.Classifications) > 0 {
		for _, classification := range p.Applicability.Classifications {
			for _, artClass := range article.Classification {
				if classification == artClass {
					return true
				}
			}
		}
		return false
	}

	if len(p.Applicability.Categories) > 0 {
		for _, category := range p.Applicability.Categories {
			if category == article.Category {
				return true
			}
		}
		return false
	}

	return len(p.Applicability.ArticleCodes) == 0 &&
		len(p.Applicability.Precodici) == 0 &&
		len(p.Applicability.Families) == 0 &&
		len(p.Applicability.Classifications) == 0 &&
		len(p.Applicability.Categories) == 0
}

func (p *Promotion) IsApplicableToCustomer(customer *Customer) bool {
	if len(p.Applicability.SpecificCustomers) > 0 {
		for _, custID := range p.Applicability.SpecificCustomers {
			if custID == customer.ID {
				return true
			}
		}
		return false
	}

	if len(p.Applicability.CustomerCategories) > 0 {
		for _, category := range p.Applicability.CustomerCategories {
			if category == customer.Category {
				return true
			}
		}
		return false
	}

	return true
}

func (p *Promotion) CanBeUsed(customerID string, quantity float64, amount float64) (bool, string) {
	if p.Limits.MaxUsageTotal > 0 && p.Statistics.TotalUsages >= p.Limits.MaxUsageTotal {
		return false, "promotion usage limit reached"
	}

	if p.Limits.MaxUsagePerCustomer > 0 {
		if usage, exists := p.Statistics.CustomerUsages[customerID]; exists {
			if usage >= p.Limits.MaxUsagePerCustomer {
				return false, "customer usage limit reached"
			}
		}
	}

	if p.Limits.MaxUsagePerDay > 0 && p.Statistics.UsagesToday >= p.Limits.MaxUsagePerDay {
		return false, "daily usage limit reached"
	}

	if p.Conditions.MinQuantity > 0 && quantity < p.Conditions.MinQuantity {
		return false, "minimum quantity not met"
	}

	if p.Conditions.MaxQuantity > 0 && quantity > p.Conditions.MaxQuantity {
		return false, "maximum quantity exceeded"
	}

	if p.Conditions.MinAmount > 0 && amount < p.Conditions.MinAmount {
		return false, "minimum amount not met"
	}

	if p.Conditions.MaxAmount > 0 && amount > p.Conditions.MaxAmount {
		return false, "maximum amount exceeded"
	}

	return true, ""
}

func (p *Promotion) CalculateDiscount(basePrice float64, quantity float64) float64 {
	switch p.Type {
	case PromotionTypePercentDiscount:
		return basePrice * quantity * (p.Rules.DiscountPercent / 100)

	case PromotionTypeFixedPrice:
		originalTotal := basePrice * quantity
		newTotal := p.Rules.FixedPrice * quantity
		return originalTotal - newTotal

	case PromotionTypeNxM:
		sets := int(quantity) / p.Rules.BuyQuantity
		freeItems := sets * p.Rules.GetQuantity
		return basePrice * float64(freeItems)

	default:
		return 0
	}
}

func (p *Promotion) RecordUsage(customerID string, revenue, discount float64) {
	p.Statistics.TotalUsages++
	p.Statistics.UsagesToday++
	p.Statistics.LastUsageDate = time.Now()
	p.Statistics.TotalRevenue += revenue
	p.Statistics.TotalDiscount += discount

	if p.Statistics.CustomerUsages == nil {
		p.Statistics.CustomerUsages = make(map[string]int)
	}
	p.Statistics.CustomerUsages[customerID]++

	p.UpdatedAt = time.Now()
}

func (p *Promotion) ResetDailyUsage() {
	p.Statistics.UsagesToday = 0
	p.UpdatedAt = time.Now()
}

func (p *Promotion) Activate() {
	p.IsActive = true
	p.UpdatedAt = time.Now()
}

func (p *Promotion) Deactivate() {
	p.IsActive = false
	p.UpdatedAt = time.Now()
}

func (p *Promotion) GetEffectivenessRate() float64 {
	if p.Statistics.TotalRevenue == 0 {
		return 0
	}
	return (p.Statistics.TotalRevenue - p.Statistics.TotalDiscount) / p.Statistics.TotalRevenue * 100
}

func (p *Promotion) DaysUntilExpiry() int {
	if p.ValidTo.IsZero() {
		return -1
	}
	duration := time.Until(p.ValidTo)
	return int(duration.Hours() / 24)
}

func (p *Promotion) IsExpiringSoon(days int) bool {
	daysLeft := p.DaysUntilExpiry()
	return daysLeft >= 0 && daysLeft <= days
}
