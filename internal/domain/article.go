// internal/domain/article.go

package domain

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrArticleNotFound      = errors.New("article not found")
	ErrInvalidArticleCode   = errors.New("invalid article code")
	ErrInvalidPrice         = errors.New("invalid price")
	ErrInsufficientStock    = errors.New("insufficient stock")
	ErrDuplicateBarcode     = errors.New("duplicate barcode")
	ErrInvalidApplicability = errors.New("invalid applicability")
)

type Article struct {
	ID                 primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Code               string                 `bson:"code" json:"code"`
	Description        string                 `bson:"description" json:"description"`
	ExtendedDesc       string                 `bson:"extended_desc" json:"extended_desc"`
	Brand              string                 `bson:"brand" json:"brand"`
	Category           string                 `bson:"category" json:"category"`
	Subcategory        string                 `bson:"subcategory" json:"subcategory"`
	Precodice          string                 `bson:"precodice" json:"precodice"`
	Family             string                 `bson:"family" json:"family"`
	Classification     []string               `bson:"classification" json:"classification"`
	Barcodes           []string               `bson:"barcodes" json:"barcodes"`
	Stock              StockInfo              `bson:"stock" json:"stock"`
	Pricing            PricingInfo            `bson:"pricing" json:"pricing"`
	Suppliers          []ArticleSupplier      `bson:"suppliers" json:"suppliers"`
	Applicability      []VehicleApplicability `bson:"applicability" json:"applicability"`
	Images             []string               `bson:"images" json:"images"`
	TechnicalSpecs     map[string]string      `bson:"technical_specs" json:"technical_specs"`
	Weight             float64                `bson:"weight" json:"weight"`
	Dimensions         Dimensions             `bson:"dimensions" json:"dimensions"`
	IsActive           bool                   `bson:"is_active" json:"is_active"`
	IsKit              bool                   `bson:"is_kit" json:"is_kit"`
	KitComponents      []KitComponent         `bson:"kit_components,omitempty" json:"kit_components,omitempty"`
	ReplacedBy         string                 `bson:"replaced_by,omitempty" json:"replaced_by,omitempty"`
	ReplacementHistory []Replacement          `bson:"replacement_history" json:"replacement_history"`
	CreatedAt          time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time              `bson:"updated_at" json:"updated_at"`
	CreatedBy          string                 `bson:"created_by" json:"created_by"`
	UpdatedBy          string                 `bson:"updated_by" json:"updated_by"`
}

type StockInfo struct {
	Quantity         float64   `bson:"quantity" json:"quantity"`
	MinStock         float64   `bson:"min_stock" json:"min_stock"`
	MaxStock         float64   `bson:"max_stock" json:"max_stock"`
	ReorderPoint     float64   `bson:"reorder_point" json:"reorder_point"`
	Location         string    `bson:"location" json:"location"`
	LastRestockDate  time.Time `bson:"last_restock_date" json:"last_restock_date"`
	LastMovementDate time.Time `bson:"last_movement_date" json:"last_movement_date"`
	Reserved         float64   `bson:"reserved" json:"reserved"`
	Available        float64   `bson:"available" json:"available"`
}

type PricingInfo struct {
	ListPrice        float64              `bson:"list_price" json:"list_price"`
	Currency         string               `bson:"currency" json:"currency"`
	LastPurchaseCost float64              `bson:"last_purchase_cost" json:"last_purchase_cost"`
	AverageCost      float64              `bson:"average_cost" json:"average_cost"`
	NetPrices        []NetPrice           `bson:"net_prices" json:"net_prices"`
	ActivePromotions []primitive.ObjectID `bson:"active_promotions" json:"active_promotions"`
	VAT              float64              `bson:"vat" json:"vat"`
}

type NetPrice struct {
	CustomerID primitive.ObjectID `bson:"customer_id" json:"customer_id"`
	Price      float64            `bson:"price" json:"price"`
	ValidFrom  time.Time          `bson:"valid_from" json:"valid_from"`
	ValidTo    time.Time          `bson:"valid_to" json:"valid_to"`
	CreatedBy  string             `bson:"created_by" json:"created_by"`
}

type ArticleSupplier struct {
	SupplierID    primitive.ObjectID `bson:"supplier_id" json:"supplier_id"`
	SupplierCode  string             `bson:"supplier_code" json:"supplier_code"`
	PurchasePrice float64            `bson:"purchase_price" json:"purchase_price"`
	Discount      float64            `bson:"discount" json:"discount"`
	LeadTimeDays  int                `bson:"lead_time_days" json:"lead_time_days"`
	MOQ           float64            `bson:"moq" json:"moq"`
	LastOrderDate time.Time          `bson:"last_order_date" json:"last_order_date"`
	IsPreferred   bool               `bson:"is_preferred" json:"is_preferred"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

type VehicleApplicability struct {
	Make         string `bson:"make" json:"make"`
	Model        string `bson:"model" json:"model"`
	YearFrom     int    `bson:"year_from" json:"year_from"`
	YearTo       int    `bson:"year_to" json:"year_to"`
	Engine       string `bson:"engine" json:"engine"`
	Displacement string `bson:"displacement" json:"displacement"`
	FuelType     string `bson:"fuel_type" json:"fuel_type"`
	PowerHP      int    `bson:"power_hp" json:"power_hp"`
	Notes        string `bson:"notes" json:"notes"`
}

type Dimensions struct {
	Length float64 `bson:"length" json:"length"`
	Width  float64 `bson:"width" json:"width"`
	Height float64 `bson:"height" json:"height"`
	Unit   string  `bson:"unit" json:"unit"`
}

type KitComponent struct {
	ArticleID   primitive.ObjectID `bson:"article_id" json:"article_id"`
	ArticleCode string             `bson:"article_code" json:"article_code"`
	Quantity    float64            `bson:"quantity" json:"quantity"`
}

type Replacement struct {
	OldCode    string    `bson:"old_code" json:"old_code"`
	NewCode    string    `bson:"new_code" json:"new_code"`
	Reason     string    `bson:"reason" json:"reason"`
	ReplacedAt time.Time `bson:"replaced_at" json:"replaced_at"`
	ReplacedBy string    `bson:"replaced_by" json:"replaced_by"`
}

func NewArticle(code, description string, createdBy string) (*Article, error) {
	if strings.TrimSpace(code) == "" {
		return nil, ErrInvalidArticleCode
	}

	now := time.Now()
	return &Article{
		ID:                 primitive.NewObjectID(),
		Code:               strings.ToUpper(strings.TrimSpace(code)),
		Description:        strings.TrimSpace(description),
		IsActive:           true,
		Stock:              StockInfo{Available: 0, Quantity: 0},
		Pricing:            PricingInfo{Currency: "EUR", VAT: 22.0},
		Barcodes:           []string{},
		Suppliers:          []ArticleSupplier{},
		Applicability:      []VehicleApplicability{},
		Classification:     []string{},
		Images:             []string{},
		TechnicalSpecs:     make(map[string]string),
		ReplacementHistory: []Replacement{},
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedBy:          createdBy,
		UpdatedBy:          createdBy,
	}, nil
}

func (a *Article) Validate() error {
	if strings.TrimSpace(a.Code) == "" {
		return ErrInvalidArticleCode
	}
	if a.Pricing.ListPrice < 0 {
		return ErrInvalidPrice
	}
	if a.Stock.Quantity < 0 {
		return errors.New("stock quantity cannot be negative")
	}
	return nil
}

func (a *Article) AddBarcode(barcode string) error {
	barcode = strings.TrimSpace(barcode)
	if barcode == "" {
		return errors.New("barcode cannot be empty")
	}

	for _, bc := range a.Barcodes {
		if bc == barcode {
			return ErrDuplicateBarcode
		}
	}

	a.Barcodes = append(a.Barcodes, barcode)
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Article) RemoveBarcode(barcode string) {
	for i, bc := range a.Barcodes {
		if bc == barcode {
			a.Barcodes = append(a.Barcodes[:i], a.Barcodes[i+1:]...)
			a.UpdatedAt = time.Now()
			break
		}
	}
}

func (a *Article) UpdateStock(quantity float64, reserved float64) {
	a.Stock.Quantity = quantity
	a.Stock.Reserved = reserved
	a.Stock.Available = quantity - reserved
	a.Stock.LastMovementDate = time.Now()
	a.UpdatedAt = time.Now()
}

func (a *Article) AddStock(quantity float64) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	a.Stock.Quantity += quantity
	a.Stock.Available = a.Stock.Quantity - a.Stock.Reserved
	a.Stock.LastRestockDate = time.Now()
	a.Stock.LastMovementDate = time.Now()
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Article) RemoveStock(quantity float64) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if a.Stock.Available < quantity {
		return ErrInsufficientStock
	}
	a.Stock.Quantity -= quantity
	a.Stock.Available = a.Stock.Quantity - a.Stock.Reserved
	a.Stock.LastMovementDate = time.Now()
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Article) ReserveStock(quantity float64) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if a.Stock.Available < quantity {
		return ErrInsufficientStock
	}
	a.Stock.Reserved += quantity
	a.Stock.Available = a.Stock.Quantity - a.Stock.Reserved
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Article) ReleaseReservedStock(quantity float64) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if a.Stock.Reserved < quantity {
		return errors.New("cannot release more than reserved")
	}
	a.Stock.Reserved -= quantity
	a.Stock.Available = a.Stock.Quantity - a.Stock.Reserved
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Article) IsLowStock() bool {
	return a.Stock.Available <= a.Stock.ReorderPoint
}

func (a *Article) CalculateMargin(sellingPrice float64) float64 {
	if a.Pricing.LastPurchaseCost == 0 {
		return 0
	}
	return ((sellingPrice - a.Pricing.LastPurchaseCost) / sellingPrice) * 100
}

func (a *Article) IsSottocosto(sellingPrice float64, threshold float64) bool {
	margin := a.CalculateMargin(sellingPrice)
	return margin < threshold
}

func (a *Article) AddSupplier(supplier ArticleSupplier) {
	for i, s := range a.Suppliers {
		if s.SupplierID == supplier.SupplierID {
			a.Suppliers[i] = supplier
			a.UpdatedAt = time.Now()
			return
		}
	}
	a.Suppliers = append(a.Suppliers, supplier)
	a.UpdatedAt = time.Now()
}

func (a *Article) GetBestSupplier() *ArticleSupplier {
	var best *ArticleSupplier
	for i, s := range a.Suppliers {
		if s.IsPreferred {
			return &a.Suppliers[i]
		}
		if best == nil || s.PurchasePrice*(1-s.Discount/100) < best.PurchasePrice*(1-best.Discount/100) {
			best = &a.Suppliers[i]
		}
	}
	return best
}

func (a *Article) AddApplicability(applicability VehicleApplicability) error {
	if applicability.Make == "" || applicability.Model == "" {
		return ErrInvalidApplicability
	}
	a.Applicability = append(a.Applicability, applicability)
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Article) IsApplicableTo(make, model string, year int) bool {
	make = strings.ToLower(strings.TrimSpace(make))
	model = strings.ToLower(strings.TrimSpace(model))

	for _, app := range a.Applicability {
		if strings.ToLower(app.Make) == make && strings.ToLower(app.Model) == model {
			if year >= app.YearFrom && (app.YearTo == 0 || year <= app.YearTo) {
				return true
			}
		}
	}
	return false
}

func (a *Article) ReplaceWith(newCode, reason, replacedBy string) {
	replacement := Replacement{
		OldCode:    a.Code,
		NewCode:    newCode,
		Reason:     reason,
		ReplacedAt: time.Now(),
		ReplacedBy: replacedBy,
	}
	a.ReplacementHistory = append(a.ReplacementHistory, replacement)
	a.ReplacedBy = newCode
	a.IsActive = false
	a.UpdatedAt = time.Now()
	a.UpdatedBy = replacedBy
}

func (a *Article) GetAvailableForKitProduction() float64 {
	if !a.IsKit || len(a.KitComponents) == 0 {
		return a.Stock.Available
	}
	return 0
}

func (a *Article) AddNetPrice(netPrice NetPrice) {
	for i, np := range a.Pricing.NetPrices {
		if np.CustomerID == netPrice.CustomerID {
			a.Pricing.NetPrices[i] = netPrice
			a.UpdatedAt = time.Now()
			return
		}
	}
	a.Pricing.NetPrices = append(a.Pricing.NetPrices, netPrice)
	a.UpdatedAt = time.Now()
}

func (a *Article) GetNetPrice(customerID primitive.ObjectID) *NetPrice {
	now := time.Now()
	for _, np := range a.Pricing.NetPrices {
		if np.CustomerID == customerID {
			if now.After(np.ValidFrom) && (np.ValidTo.IsZero() || now.Before(np.ValidTo)) {
				return &np
			}
		}
	}
	return nil
}

func (a *Article) GetExpiredNetPrices() []NetPrice {
	now := time.Now()
	var expired []NetPrice
	for _, np := range a.Pricing.NetPrices {
		if !np.ValidTo.IsZero() && now.After(np.ValidTo) {
			expired = append(expired, np)
		}
	}
	return expired
}
