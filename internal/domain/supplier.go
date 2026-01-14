// internal/domain/supplier.go

package domain

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrSupplierNotFound    = errors.New("supplier not found")
	ErrInvalidSupplierData = errors.New("invalid supplier data")
)

type SupplierRating string

const (
	RatingExcellent SupplierRating = "excellent"
	RatingGood      SupplierRating = "good"
	RatingAverage   SupplierRating = "average"
	RatingPoor      SupplierRating = "poor"
)

type Supplier struct {
	ID                   primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Code                 string               `bson:"code" json:"code"`
	CompanyName          string               `bson:"company_name" json:"company_name"`
	VATNumber            string               `bson:"vat_number" json:"vat_number"`
	FiscalCode           string               `bson:"fiscal_code" json:"fiscal_code"`
	ContactInfo          ContactInfo          `bson:"contact_info" json:"contact_info"`
	Address              Address              `bson:"address" json:"address"`
	PaymentTerms         PaymentTerms         `bson:"payment_terms" json:"payment_terms"`
	Rating               SupplierRating       `bson:"rating" json:"rating"`
	IsActive             bool                 `bson:"is_active" json:"is_active"`
	IsPreferred          bool                 `bson:"is_preferred" json:"is_preferred"`
	Budgets              []primitive.ObjectID `bson:"budgets" json:"budgets"`
	DeliveryPerformance  DeliveryStats        `bson:"delivery_performance" json:"delivery_performance"`
	CommercialConditions CommercialConditions `bson:"commercial_conditions" json:"commercial_conditions"`
	BankDetails          BankDetails          `bson:"bank_details" json:"bank_details"`
	Notes                string               `bson:"notes" json:"notes"`
	Tags                 []string             `bson:"tags" json:"tags"`
	CreatedAt            time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time            `bson:"updated_at" json:"updated_at"`
	CreatedBy            string               `bson:"created_by" json:"created_by"`
	UpdatedBy            string               `bson:"updated_by" json:"updated_by"`
}

type DeliveryStats struct {
	OnTimeDeliveries int       `bson:"on_time_deliveries" json:"on_time_deliveries"`
	LateDeliveries   int       `bson:"late_deliveries" json:"late_deliveries"`
	AverageLeadDays  float64   `bson:"average_lead_days" json:"average_lead_days"`
	LastDeliveryDate time.Time `bson:"last_delivery_date" json:"last_delivery_date"`
	TotalOrders      int       `bson:"total_orders" json:"total_orders"`
	DefectiveRate    float64   `bson:"defective_rate" json:"defective_rate"`
}

type CommercialConditions struct {
	BaseDiscount     float64          `bson:"base_discount" json:"base_discount"`
	VolumeDiscounts  []VolumeDiscount `bson:"volume_discounts" json:"volume_discounts"`
	PaymentDiscount  float64          `bson:"payment_discount" json:"payment_discount"`
	MinOrderAmount   float64          `bson:"min_order_amount" json:"min_order_amount"`
	FreeShippingFrom float64          `bson:"free_shipping_from" json:"free_shipping_from"`
	ReturnPolicy     string           `bson:"return_policy" json:"return_policy"`
	WarrantyDays     int              `bson:"warranty_days" json:"warranty_days"`
}

type VolumeDiscount struct {
	MinAmount       float64 `bson:"min_amount" json:"min_amount"`
	MaxAmount       float64 `bson:"max_amount" json:"max_amount"`
	DiscountPercent float64 `bson:"discount_percent" json:"discount_percent"`
}

type BankDetails struct {
	BankName      string `bson:"bank_name" json:"bank_name"`
	IBAN          string `bson:"iban" json:"iban"`
	SWIFT         string `bson:"swift" json:"swift"`
	AccountHolder string `bson:"account_holder" json:"account_holder"`
}

func NewSupplier(code, companyName, createdBy string) (*Supplier, error) {
	if strings.TrimSpace(code) == "" {
		return nil, errors.New("supplier code cannot be empty")
	}
	if strings.TrimSpace(companyName) == "" {
		return nil, errors.New("company name cannot be empty")
	}

	now := time.Now()
	return &Supplier{
		ID:          primitive.NewObjectID(),
		Code:        strings.ToUpper(strings.TrimSpace(code)),
		CompanyName: strings.TrimSpace(companyName),
		Rating:      RatingAverage,
		IsActive:    true,
		IsPreferred: false,
		Budgets:     []primitive.ObjectID{},
		DeliveryPerformance: DeliveryStats{
			OnTimeDeliveries: 0,
			LateDeliveries:   0,
			TotalOrders:      0,
		},
		CommercialConditions: CommercialConditions{
			BaseDiscount:    0,
			VolumeDiscounts: []VolumeDiscount{},
			WarrantyDays:    365,
		},
		ContactInfo: ContactInfo{},
		Address:     Address{Country: "IT"},
		PaymentTerms: PaymentTerms{
			Method:  "bank_transfer",
			DaysNet: 60,
		},
		BankDetails: BankDetails{},
		Tags:        []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}, nil
}

func (s *Supplier) Validate() error {
	if strings.TrimSpace(s.Code) == "" {
		return errors.New("supplier code cannot be empty")
	}
	if strings.TrimSpace(s.CompanyName) == "" {
		return errors.New("company name cannot be empty")
	}
	return nil
}

func (s *Supplier) RecordDelivery(isOnTime bool, leadDays int) {
	s.DeliveryPerformance.TotalOrders++

	if isOnTime {
		s.DeliveryPerformance.OnTimeDeliveries++
	} else {
		s.DeliveryPerformance.LateDeliveries++
	}

	currentTotal := s.DeliveryPerformance.AverageLeadDays * float64(s.DeliveryPerformance.TotalOrders-1)
	s.DeliveryPerformance.AverageLeadDays = (currentTotal + float64(leadDays)) / float64(s.DeliveryPerformance.TotalOrders)
	s.DeliveryPerformance.LastDeliveryDate = time.Now()

	s.UpdateRating()
	s.UpdatedAt = time.Now()
}

func (s *Supplier) GetOnTimeDeliveryRate() float64 {
	if s.DeliveryPerformance.TotalOrders == 0 {
		return 0
	}
	return (float64(s.DeliveryPerformance.OnTimeDeliveries) / float64(s.DeliveryPerformance.TotalOrders)) * 100
}

func (s *Supplier) UpdateRating() {
	onTimeRate := s.GetOnTimeDeliveryRate()

	if onTimeRate >= 95 && s.DeliveryPerformance.DefectiveRate < 1 {
		s.Rating = RatingExcellent
	} else if onTimeRate >= 85 && s.DeliveryPerformance.DefectiveRate < 3 {
		s.Rating = RatingGood
	} else if onTimeRate >= 70 && s.DeliveryPerformance.DefectiveRate < 5 {
		s.Rating = RatingAverage
	} else {
		s.Rating = RatingPoor
	}
}

func (s *Supplier) AddVolumeDiscount(volumeDiscount VolumeDiscount) error {
	if volumeDiscount.DiscountPercent < 0 || volumeDiscount.DiscountPercent > 100 {
		return errors.New("invalid discount percentage")
	}
	if volumeDiscount.MinAmount < 0 {
		return errors.New("minimum amount cannot be negative")
	}

	s.CommercialConditions.VolumeDiscounts = append(s.CommercialConditions.VolumeDiscounts, volumeDiscount)
	s.UpdatedAt = time.Now()
	return nil
}

func (s *Supplier) GetApplicableDiscount(orderAmount float64) float64 {
	totalDiscount := s.CommercialConditions.BaseDiscount

	for _, vd := range s.CommercialConditions.VolumeDiscounts {
		if orderAmount >= vd.MinAmount && (vd.MaxAmount == 0 || orderAmount <= vd.MaxAmount) {
			if vd.DiscountPercent > totalDiscount {
				totalDiscount = vd.DiscountPercent
			}
		}
	}

	return totalDiscount
}

func (s *Supplier) CalculateNetPrice(grossPrice, orderAmount float64) float64 {
	discount := s.GetApplicableDiscount(orderAmount)
	return grossPrice * (1 - discount/100)
}

func (s *Supplier) MeetsMinimumOrder(orderAmount float64) bool {
	return orderAmount >= s.CommercialConditions.MinOrderAmount
}

func (s *Supplier) QualifiesForFreeShipping(orderAmount float64) bool {
	if s.CommercialConditions.FreeShippingFrom == 0 {
		return false
	}
	return orderAmount >= s.CommercialConditions.FreeShippingFrom
}

func (s *Supplier) AddTag(tag string) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return
	}

	for _, t := range s.Tags {
		if t == tag {
			return
		}
	}

	s.Tags = append(s.Tags, tag)
	s.UpdatedAt = time.Now()
}

func (s *Supplier) RemoveTag(tag string) {
	for i, t := range s.Tags {
		if t == tag {
			s.Tags = append(s.Tags[:i], s.Tags[i+1:]...)
			s.UpdatedAt = time.Now()
			return
		}
	}
}

func (s *Supplier) SetPreferred(preferred bool) {
	s.IsPreferred = preferred
	s.UpdatedAt = time.Now()
}

func (s *Supplier) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = time.Now()
}

func (s *Supplier) Activate() {
	s.IsActive = true
	s.UpdatedAt = time.Now()
}

func (s *Supplier) GetReliabilityScore() float64 {
	if s.DeliveryPerformance.TotalOrders == 0 {
		return 50.0
	}

	onTimeRate := s.GetOnTimeDeliveryRate()
	defectiveImpact := (100 - s.DeliveryPerformance.DefectiveRate*10)

	if defectiveImpact < 0 {
		defectiveImpact = 0
	}

	return (onTimeRate*0.7 + defectiveImpact*0.3)
}

func (s *Supplier) IsReliable() bool {
	return s.GetReliabilityScore() >= 70
}
