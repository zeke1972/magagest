// internal/domain/kit.go

package domain

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrKitNotFound           = errors.New("kit not found")
	ErrInvalidKitComponents  = errors.New("kit must have at least 2 components")
	ErrComponentNotAvailable = errors.New("component not available in sufficient quantity")
	ErrKitNotActive          = errors.New("kit is not active")
)

type PricingStrategy string

const (
	PricingStrategyCalculated PricingStrategy = "calculated"
	PricingStrategyCustom     PricingStrategy = "custom"
)

type Kit struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Code              string             `bson:"code" json:"code"`
	Name              string             `bson:"name" json:"name"`
	Description       string             `bson:"description" json:"description"`
	Category          string             `bson:"category" json:"category"`
	Components        []KitComponent     `bson:"components" json:"components"`
	PricingStrategy   PricingStrategy    `bson:"pricing_strategy" json:"pricing_strategy"`
	CalculatedPrice   float64            `bson:"calculated_price" json:"calculated_price"`
	CustomPrice       float64            `bson:"custom_price" json:"custom_price"`
	DiscountPercent   float64            `bson:"discount_percent" json:"discount_percent"`
	IsActive          bool               `bson:"is_active" json:"is_active"`
	AvailableQuantity float64            `bson:"available_quantity" json:"available_quantity"`
	Images            []string           `bson:"images" json:"images"`
	Tags              []string           `bson:"tags" json:"tags"`
	SalesCount        int                `bson:"sales_count" json:"sales_count"`
	LastSoldDate      time.Time          `bson:"last_sold_date" json:"last_sold_date"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedBy         string             `bson:"created_by" json:"created_by"`
	UpdatedBy         string             `bson:"updated_by" json:"updated_by"`
}

func NewKit(code, name string, createdBy string) (*Kit, error) {
	if strings.TrimSpace(code) == "" {
		return nil, errors.New("kit code cannot be empty")
	}
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("kit name cannot be empty")
	}

	now := time.Now()
	return &Kit{
		ID:              primitive.NewObjectID(),
		Code:            strings.ToUpper(strings.TrimSpace(code)),
		Name:            strings.TrimSpace(name),
		Components:      []KitComponent{},
		PricingStrategy: PricingStrategyCalculated,
		IsActive:        true,
		Images:          []string{},
		Tags:            []string{},
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       createdBy,
		UpdatedBy:       createdBy,
	}, nil
}

func (k *Kit) Validate() error {
	if strings.TrimSpace(k.Code) == "" {
		return errors.New("kit code cannot be empty")
	}
	if strings.TrimSpace(k.Name) == "" {
		return errors.New("kit name cannot be empty")
	}
	if len(k.Components) < 2 {
		return ErrInvalidKitComponents
	}
	return nil
}

func (k *Kit) AddComponent(articleID primitive.ObjectID, articleCode string, quantity float64) error {
	if quantity <= 0 {
		return errors.New("component quantity must be positive")
	}

	for i, comp := range k.Components {
		if comp.ArticleID == articleID {
			k.Components[i].Quantity = quantity
			k.UpdatedAt = time.Now()
			return nil
		}
	}

	component := KitComponent{
		ArticleID:   articleID,
		ArticleCode: articleCode,
		Quantity:    quantity,
	}

	k.Components = append(k.Components, component)
	k.UpdatedAt = time.Now()
	return nil
}

func (k *Kit) RemoveComponent(articleID primitive.ObjectID) error {
	for i, comp := range k.Components {
		if comp.ArticleID == articleID {
			k.Components = append(k.Components[:i], k.Components[i+1:]...)
			k.UpdatedAt = time.Now()
			return nil
		}
	}
	return errors.New("component not found in kit")
}

func (k *Kit) CalculatePrice(articles map[primitive.ObjectID]*Article) (float64, error) {
	if len(k.Components) == 0 {
		return 0, ErrInvalidKitComponents
	}

	totalPrice := 0.0
	for _, comp := range k.Components {
		article, exists := articles[comp.ArticleID]
		if !exists {
			return 0, errors.New("article not found for component: " + comp.ArticleCode)
		}
		totalPrice += article.Pricing.ListPrice * comp.Quantity
	}

	k.CalculatedPrice = totalPrice

	if k.DiscountPercent > 0 {
		k.CalculatedPrice = totalPrice * (1 - k.DiscountPercent/100)
	}

	k.UpdatedAt = time.Now()
	return k.CalculatedPrice, nil
}

func (k *Kit) GetFinalPrice() float64 {
	if k.PricingStrategy == PricingStrategyCustom && k.CustomPrice > 0 {
		return k.CustomPrice
	}
	return k.CalculatedPrice
}

func (k *Kit) SetCustomPrice(price float64) error {
	if price <= 0 {
		return errors.New("price must be positive")
	}
	k.CustomPrice = price
	k.PricingStrategy = PricingStrategyCustom
	k.UpdatedAt = time.Now()
	return nil
}

func (k *Kit) UseCalculatedPrice() {
	k.PricingStrategy = PricingStrategyCalculated
	k.UpdatedAt = time.Now()
}

func (k *Kit) CalculateAvailability(articles map[primitive.ObjectID]*Article) float64 {
	if len(k.Components) == 0 {
		k.AvailableQuantity = 0
		return 0
	}

	minAvailable := -1.0

	for _, comp := range k.Components {
		article, exists := articles[comp.ArticleID]
		if !exists {
			k.AvailableQuantity = 0
			return 0
		}

		if article.Stock.Available <= 0 {
			k.AvailableQuantity = 0
			return 0
		}

		availableKits := article.Stock.Available / comp.Quantity

		if minAvailable < 0 || availableKits < minAvailable {
			minAvailable = availableKits
		}
	}

	if minAvailable < 0 {
		minAvailable = 0
	}

	k.AvailableQuantity = minAvailable
	k.UpdatedAt = time.Now()
	return minAvailable
}

func (k *Kit) CanFulfill(quantity float64, articles map[primitive.ObjectID]*Article) (bool, []string) {
	var unavailableComponents []string

	for _, comp := range k.Components {
		article, exists := articles[comp.ArticleID]
		if !exists {
			unavailableComponents = append(unavailableComponents, comp.ArticleCode+" (not found)")
			continue
		}

		requiredQuantity := comp.Quantity * quantity
		if article.Stock.Available < requiredQuantity {
			unavailableComponents = append(unavailableComponents,
				comp.ArticleCode+" (need "+formatFloat(requiredQuantity)+", have "+formatFloat(article.Stock.Available)+")")
		}
	}

	return len(unavailableComponents) == 0, unavailableComponents
}

func formatFloat(val float64) string {
	if val == float64(int64(val)) {
		return string(rune(int64(val)))
	}
	return string(rune(int64(val*100))) + "/100"
}

func (k *Kit) ReserveComponents(quantity float64, articles map[primitive.ObjectID]*Article) error {
	canFulfill, unavailable := k.CanFulfill(quantity, articles)
	if !canFulfill {
		return errors.New("cannot fulfill kit: " + strings.Join(unavailable, ", "))
	}

	for _, comp := range k.Components {
		article := articles[comp.ArticleID]
		requiredQuantity := comp.Quantity * quantity
		if err := article.ReserveStock(requiredQuantity); err != nil {
			return err
		}
	}

	return nil
}

func (k *Kit) ReleaseComponents(quantity float64, articles map[primitive.ObjectID]*Article) error {
	for _, comp := range k.Components {
		article, exists := articles[comp.ArticleID]
		if !exists {
			continue
		}
		requiredQuantity := comp.Quantity * quantity
		if err := article.ReleaseReservedStock(requiredQuantity); err != nil {
			return err
		}
	}
	return nil
}

func (k *Kit) DecomposeKit(quantity float64, articles map[primitive.ObjectID]*Article) error {
	canFulfill, unavailable := k.CanFulfill(quantity, articles)
	if !canFulfill {
		return errors.New("cannot decompose kit: " + strings.Join(unavailable, ", "))
	}

	for _, comp := range k.Components {
		article := articles[comp.ArticleID]
		requiredQuantity := comp.Quantity * quantity
		if err := article.RemoveStock(requiredQuantity); err != nil {
			return err
		}
	}

	k.SalesCount++
	k.LastSoldDate = time.Now()
	k.UpdatedAt = time.Now()
	return nil
}

func (k *Kit) GetComponentsSummary() []string {
	summary := make([]string, len(k.Components))
	for i, comp := range k.Components {
		qtyStr := "1"
		if comp.Quantity != 1 {
			qtyStr = formatFloat(comp.Quantity)
		}
		summary[i] = qtyStr + "x " + comp.ArticleCode
	}
	return summary
}

func (k *Kit) AddTag(tag string) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return
	}

	for _, t := range k.Tags {
		if t == tag {
			return
		}
	}

	k.Tags = append(k.Tags, tag)
	k.UpdatedAt = time.Now()
}

func (k *Kit) RemoveTag(tag string) {
	for i, t := range k.Tags {
		if t == tag {
			k.Tags = append(k.Tags[:i], k.Tags[i+1:]...)
			k.UpdatedAt = time.Now()
			return
		}
	}
}

func (k *Kit) Activate() {
	k.IsActive = true
	k.UpdatedAt = time.Now()
}

func (k *Kit) Deactivate() {
	k.IsActive = false
	k.UpdatedAt = time.Now()
}

func (k *Kit) GetSavingsPercent(articles map[primitive.ObjectID]*Article) float64 {
	if k.CalculatedPrice == 0 {
		return 0
	}

	totalPrice := 0.0
	for _, comp := range k.Components {
		article, exists := articles[comp.ArticleID]
		if !exists {
			continue
		}
		totalPrice += article.Pricing.ListPrice * comp.Quantity
	}

	if totalPrice == 0 {
		return 0
	}

	finalPrice := k.GetFinalPrice()
	savings := ((totalPrice - finalPrice) / totalPrice) * 100

	if savings < 0 {
		savings = 0
	}

	return savings
}

func (k *Kit) IsPopular() bool {
	return k.SalesCount > 10
}

func (k *Kit) DaysSinceLastSale() int {
	if k.LastSoldDate.IsZero() {
		return -1
	}
	duration := time.Since(k.LastSoldDate)
	return int(duration.Hours() / 24)
}

func (k *Kit) UpdateComponentQuantity(articleID primitive.ObjectID, newQuantity float64) error {
	if newQuantity <= 0 {
		return errors.New("quantity must be positive")
	}

	for i, comp := range k.Components {
		if comp.ArticleID == articleID {
			k.Components[i].Quantity = newQuantity
			k.UpdatedAt = time.Now()
			return nil
		}
	}

	return errors.New("component not found in kit")
}

func (k *Kit) HasComponent(articleID primitive.ObjectID) bool {
	for _, comp := range k.Components {
		if comp.ArticleID == articleID {
			return true
		}
	}
	return false
}

func (k *Kit) GetComponentQuantity(articleID primitive.ObjectID) float64 {
	for _, comp := range k.Components {
		if comp.ArticleID == articleID {
			return comp.Quantity
		}
	}
	return 0
}
