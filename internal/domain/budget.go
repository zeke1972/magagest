// internal/domain/budget.go

package domain

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BudgetType string

const (
	BudgetTypeCustomer BudgetType = "customer"
	BudgetTypeSupplier BudgetType = "supplier"
	BudgetTypeOperator BudgetType = "operator"
	BudgetTypeGlobal   BudgetType = "global"
)

type BudgetIncentive struct {
	ID               primitive.ObjectID `bson:"id"`
	MinRevenue       float64            `bson:"min_revenue"`
	MaxRevenue       float64            `bson:"max_revenue"`
	IncentivePercent float64            `bson:"incentive_percent"`
	IncentiveAmount  float64            `bson:"incentive_amount"`
	Description      string             `bson:"description"`
}

type Budget struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	EntityID  primitive.ObjectID `bson:"entity_id"`
	Type      BudgetType         `bson:"type"`
	Year      int                `bson:"year"`
	Quarter   int                `bson:"quarter"`
	StartDate time.Time          `bson:"start_date"`
	EndDate   time.Time          `bson:"end_date"`

	TargetRevenue float64 `bson:"target_revenue"`
	ActualRevenue float64 `bson:"actual_revenue"`

	TargetMargin float64 `bson:"target_margin"`
	ActualMargin float64 `bson:"actual_margin"`

	TargetOrders int `bson:"target_orders"`
	ActualOrders int `bson:"actual_orders"`

	Incentives []BudgetIncentive `bson:"incentives"`

	Notes     string    `bson:"notes"`
	CreatedBy string    `bson:"created_by"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

var (
	ErrBudgetNotFound        = errors.New("budget not found")
	ErrInvalidBudgetPeriod   = errors.New("invalid budget period")
	ErrBudgetAlreadyExists   = errors.New("budget already exists for this period")
	ErrInvalidIncentiveRange = errors.New("invalid incentive range")
)

func NewBudget(
	entityID primitive.ObjectID,
	budgetType BudgetType,
	year, quarter int,
	targetRevenue, targetMargin float64,
	targetOrders int,
	createdBy string,
) (*Budget, error) {
	if year < 2000 || year > 2100 {
		return nil, ErrInvalidBudgetPeriod
	}

	if quarter < 0 || quarter > 4 {
		return nil, ErrInvalidBudgetPeriod
	}

	startDate, endDate := calculateBudgetPeriod(year, quarter)

	budget := &Budget{
		ID:            primitive.NewObjectID(),
		EntityID:      entityID,
		Type:          budgetType,
		Year:          year,
		Quarter:       quarter,
		StartDate:     startDate,
		EndDate:       endDate,
		TargetRevenue: targetRevenue,
		TargetMargin:  targetMargin,
		TargetOrders:  targetOrders,
		ActualRevenue: 0,
		ActualMargin:  0,
		ActualOrders:  0,
		Incentives:    []BudgetIncentive{},
		CreatedBy:     createdBy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return budget, nil
}

func calculateBudgetPeriod(year, quarter int) (time.Time, time.Time) {
	if quarter == 0 {
		return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	startMonth := (quarter-1)*3 + 1
	endMonth := startMonth + 2

	startDate := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, time.Month(endMonth+1), 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)

	return startDate, endDate
}

func (b *Budget) UpdateActuals(revenue, margin float64, orders int) {
	b.ActualRevenue = revenue
	b.ActualMargin = margin
	b.ActualOrders = orders
	b.UpdatedAt = time.Now()
}

func (b *Budget) AddRevenue(amount float64) {
	b.ActualRevenue += amount
	b.UpdatedAt = time.Now()
}

func (b *Budget) AddMargin(amount float64) {
	b.ActualMargin += amount
	b.UpdatedAt = time.Now()
}

func (b *Budget) AddOrder() {
	b.ActualOrders++
	b.UpdatedAt = time.Now()
}

func (b *Budget) GetRevenueAchievement() float64 {
	if b.TargetRevenue == 0 {
		return 0
	}
	return (b.ActualRevenue / b.TargetRevenue) * 100
}

func (b *Budget) GetMarginAchievement() float64 {
	if b.TargetMargin == 0 {
		return 0
	}
	return (b.ActualMargin / b.TargetMargin) * 100
}

func (b *Budget) GetOrdersAchievement() float64 {
	if b.TargetOrders == 0 {
		return 0
	}
	return (float64(b.ActualOrders) / float64(b.TargetOrders)) * 100
}

func (b *Budget) GetOverallAchievement() float64 {
	revenueAch := b.GetRevenueAchievement()
	marginAch := b.GetMarginAchievement()
	ordersAch := b.GetOrdersAchievement()

	return (revenueAch + marginAch + ordersAch) / 3
}

func (b *Budget) GetRemainingRevenue() float64 {
	remaining := b.TargetRevenue - b.ActualRevenue
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (b *Budget) GetRemainingMargin() float64 {
	remaining := b.TargetMargin - b.ActualMargin
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (b *Budget) GetRemainingOrders() int {
	remaining := b.TargetOrders - b.ActualOrders
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (b *Budget) GetDaysRemaining() int {
	now := time.Now()
	if now.After(b.EndDate) {
		return 0
	}
	return int(b.EndDate.Sub(now).Hours() / 24)
}

func (b *Budget) GetElapsedDays() int {
	now := time.Now()
	if now.Before(b.StartDate) {
		return 0
	}
	if now.After(b.EndDate) {
		return int(b.EndDate.Sub(b.StartDate).Hours() / 24)
	}
	return int(now.Sub(b.StartDate).Hours() / 24)
}

func (b *Budget) GetTotalDays() int {
	return int(b.EndDate.Sub(b.StartDate).Hours() / 24)
}

func (b *Budget) GetTimeProgress() float64 {
	elapsed := b.GetElapsedDays()
	total := b.GetTotalDays()
	if total == 0 {
		return 0
	}
	progress := (float64(elapsed) / float64(total)) * 100
	if progress > 100 {
		return 100
	}
	return progress
}

func (b *Budget) IsOnTrack() bool {
	timeProgress := b.GetTimeProgress()
	revenueProgress := b.GetRevenueAchievement()

	return revenueProgress >= (timeProgress - 10)
}

func (b *Budget) AddIncentive(minRevenue, maxRevenue, incentivePercent, incentiveAmount float64, description string) error {
	if minRevenue >= maxRevenue {
		return ErrInvalidIncentiveRange
	}

	incentive := BudgetIncentive{
		ID:               primitive.NewObjectID(),
		MinRevenue:       minRevenue,
		MaxRevenue:       maxRevenue,
		IncentivePercent: incentivePercent,
		IncentiveAmount:  incentiveAmount,
		Description:      description,
	}

	b.Incentives = append(b.Incentives, incentive)
	b.UpdatedAt = time.Now()

	return nil
}

func (b *Budget) RemoveIncentive(incentiveID primitive.ObjectID) {
	for i, inc := range b.Incentives {
		if inc.ID == incentiveID {
			b.Incentives = append(b.Incentives[:i], b.Incentives[i+1:]...)
			b.UpdatedAt = time.Now()
			return
		}
	}
}

func (b *Budget) CalculateIncentive() float64 {
	for _, inc := range b.Incentives {
		if b.ActualRevenue >= inc.MinRevenue && b.ActualRevenue < inc.MaxRevenue {
			if inc.IncentiveAmount > 0 {
				return inc.IncentiveAmount
			}
			return b.ActualRevenue * (inc.IncentivePercent / 100)
		}
	}
	return 0
}

func (b *Budget) GetApplicableIncentive() *BudgetIncentive {
	for _, inc := range b.Incentives {
		if b.ActualRevenue >= inc.MinRevenue && b.ActualRevenue < inc.MaxRevenue {
			return &inc
		}
	}
	return nil
}

func (b *Budget) IsActive() bool {
	now := time.Now()
	return now.After(b.StartDate) && now.Before(b.EndDate)
}

func (b *Budget) Activate() error {
	if b.IsActive() {
		return errors.New("budget is already active")
	}
	b.UpdatedAt = time.Now()
	return nil
}

func (b *Budget) Deactivate() {
	b.UpdatedAt = time.Now()
}

func (b *Budget) IsExpired() bool {
	return time.Now().After(b.EndDate)
}

func (b *Budget) IsFuture() bool {
	return time.Now().Before(b.StartDate)
}

func (b *Budget) GetStatus() string {
	if b.IsFuture() {
		return "future"
	}
	if b.IsActive() {
		return "active"
	}
	if b.IsExpired() {
		return "expired"
	}
	return "unknown"
}

func (b *Budget) Validate() error {
	if b.EntityID.IsZero() {
		return errors.New("entity ID is required")
	}

	if b.TargetRevenue < 0 {
		return errors.New("target revenue cannot be negative")
	}

	if b.TargetMargin < 0 {
		return errors.New("target margin cannot be negative")
	}

	if b.TargetOrders < 0 {
		return errors.New("target orders cannot be negative")
	}

	if b.StartDate.After(b.EndDate) {
		return errors.New("start date must be before end date")
	}

	return nil
}
