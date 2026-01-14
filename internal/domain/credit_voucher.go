// internal/domain/credit_voucher.go

package domain

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VoucherStatus string

const (
	VoucherStatusIssued        VoucherStatus = "issued"
	VoucherStatusPartiallyUsed VoucherStatus = "partially_used"
	VoucherStatusUsed          VoucherStatus = "used"
	VoucherStatusExpired       VoucherStatus = "expired"
	VoucherStatusCancelled     VoucherStatus = "cancelled"
)

type VoucherUsage struct {
	ID           primitive.ObjectID `bson:"id"`
	DocumentID   string             `bson:"document_id"`
	DocumentType string             `bson:"document_type"`
	Amount       float64            `bson:"amount"`
	UsedBy       string             `bson:"used_by"`
	UsedAt       time.Time          `bson:"used_at"`
	Notes        string             `bson:"notes"`
}

type CreditVoucher struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	Code            string             `bson:"code"`
	CustomerID      primitive.ObjectID `bson:"customer_id"`
	OriginalAmount  float64            `bson:"original_amount"`
	RemainingAmount float64            `bson:"remaining_amount"`
	Status          VoucherStatus      `bson:"status"`
	Reason          string             `bson:"reason"`
	IssuedDate      time.Time          `bson:"issued_date"`
	ExpiryDate      time.Time          `bson:"expiry_date"`
	LastUsed        time.Time          `bson:"last_used"`
	UsageHistory    []VoucherUsage     `bson:"usage_history"`
	CreatedBy       string             `bson:"created_by"`
	CreatedAt       time.Time          `bson:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at"`
}

var (
	ErrVoucherNotFound      = errors.New("credit voucher not found")
	ErrVoucherExpired       = errors.New("credit voucher has expired")
	ErrVoucherUsed          = errors.New("credit voucher is already fully used")
	ErrVoucherCancelled     = errors.New("credit voucher has been cancelled")
	ErrInsufficientBalance  = errors.New("insufficient voucher balance")
	ErrInvalidVoucherAmount = errors.New("invalid voucher amount")
)

func NewCreditVoucher(
	customerID primitive.ObjectID,
	amount float64,
	reason string,
	expiryDays int,
	createdBy string,
) (*CreditVoucher, error) {
	if amount <= 0 {
		return nil, ErrInvalidVoucherAmount
	}

	code := generateVoucherCode()
	issuedDate := time.Now()
	var expiryDate time.Time
	if expiryDays > 0 {
		expiryDate = issuedDate.AddDate(0, 0, expiryDays)
	}

	voucher := &CreditVoucher{
		ID:              primitive.NewObjectID(),
		Code:            code,
		CustomerID:      customerID,
		OriginalAmount:  amount,
		RemainingAmount: amount,
		Status:          VoucherStatusIssued,
		Reason:          reason,
		IssuedDate:      issuedDate,
		ExpiryDate:      expiryDate,
		UsageHistory:    []VoucherUsage{},
		CreatedBy:       createdBy,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return voucher, nil
}

func generateVoucherCode() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("VC-%d", timestamp)
}

func (v *CreditVoucher) Use(amount float64, documentID, documentType, usedBy, notes string) error {
	if v.Status == VoucherStatusExpired {
		return errors.New("voucher has expired")
	}

	if v.Status == VoucherStatusUsed {
		return errors.New("voucher is already fully used")
	}

	if v.Status == VoucherStatusCancelled {
		return errors.New("voucher has been cancelled")
	}

	timestamp := time.Now()

	if !v.ExpiryDate.IsZero() && timestamp.After(v.ExpiryDate) {
		v.Status = VoucherStatusExpired
		return errors.New("voucher has expired")
	}

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	if amount > v.RemainingAmount {
		return errors.New("insufficient voucher balance")
	}

	usage := VoucherUsage{
		ID:           primitive.NewObjectID(),
		DocumentID:   documentID,
		DocumentType: documentType,
		Amount:       amount,
		UsedBy:       usedBy,
		UsedAt:       timestamp,
		Notes:        notes,
	}

	v.UsageHistory = append(v.UsageHistory, usage)
	v.RemainingAmount -= amount
	v.LastUsed = timestamp

	if v.RemainingAmount <= 0 {
		v.Status = VoucherStatusUsed
		v.RemainingAmount = 0
	} else if v.RemainingAmount < v.OriginalAmount {
		v.Status = VoucherStatusPartiallyUsed
	}

	v.UpdatedAt = timestamp

	return nil
}

func (v *CreditVoucher) Cancel(reason string) error {
	if v.Status == VoucherStatusUsed {
		return errors.New("cannot cancel a fully used voucher")
	}

	if v.Status == VoucherStatusCancelled {
		return errors.New("voucher is already cancelled")
	}

	v.Status = VoucherStatusCancelled
	v.Reason = reason
	v.UpdatedAt = time.Now()

	return nil
}

func (v *CreditVoucher) IsExpired() bool {
	if v.ExpiryDate.IsZero() {
		return false
	}
	return time.Now().After(v.ExpiryDate)
}

func (v *CreditVoucher) IsValid() bool {
	if v.Status == VoucherStatusCancelled {
		return false
	}

	if v.Status == VoucherStatusExpired {
		return false
	}

	if v.Status == VoucherStatusUsed {
		return false
	}

	if v.IsExpired() {
		return false
	}

	return v.RemainingAmount > 0
}

func (v *CreditVoucher) GetUsagePercent() float64 {
	if v.OriginalAmount == 0 {
		return 0
	}
	used := v.OriginalAmount - v.RemainingAmount
	return (used / v.OriginalAmount) * 100
}

func (v *CreditVoucher) GetDaysUntilExpiry() int {
	if v.ExpiryDate.IsZero() {
		return -1
	}

	now := time.Now()
	if now.After(v.ExpiryDate) {
		return 0
	}

	return int(v.ExpiryDate.Sub(now).Hours() / 24)
}

func (v *CreditVoucher) IsExpiringSoon(days int) bool {
	daysRemaining := v.GetDaysUntilExpiry()
	if daysRemaining == -1 {
		return false
	}
	return daysRemaining <= days && daysRemaining > 0
}

func (v *CreditVoucher) Extend(days int) error {
	if v.ExpiryDate.IsZero() {
		v.ExpiryDate = time.Now().AddDate(0, 0, days)
	} else {
		v.ExpiryDate = v.ExpiryDate.AddDate(0, 0, days)
	}

	if v.Status == VoucherStatusExpired && v.RemainingAmount > 0 {
		v.Status = VoucherStatusPartiallyUsed
		if v.RemainingAmount == v.OriginalAmount {
			v.Status = VoucherStatusIssued
		}
	}

	v.UpdatedAt = time.Now()

	return nil
}

func (v *CreditVoucher) GetTotalUsed() float64 {
	return v.OriginalAmount - v.RemainingAmount
}

func (v *CreditVoucher) GetUsageCount() int {
	return len(v.UsageHistory)
}

func (v *CreditVoucher) GetLastUsage() *VoucherUsage {
	if len(v.UsageHistory) == 0 {
		return nil
	}
	return &v.UsageHistory[len(v.UsageHistory)-1]
}

func (v *CreditVoucher) CanBeUsed(amount float64) (bool, string) {
	if !v.IsValid() {
		return false, "voucher is not valid"
	}

	if amount > v.RemainingAmount {
		return false, fmt.Sprintf("insufficient balance (available: %.2f)", v.RemainingAmount)
	}

	return true, ""
}

func (v *CreditVoucher) Validate() error {
	if v.CustomerID.IsZero() {
		return errors.New("customer ID is required")
	}

	if v.OriginalAmount <= 0 {
		return errors.New("original amount must be positive")
	}

	if v.RemainingAmount < 0 {
		return errors.New("remaining amount cannot be negative")
	}

	if v.RemainingAmount > v.OriginalAmount {
		return errors.New("remaining amount cannot exceed original amount")
	}

	if v.Code == "" {
		return errors.New("voucher code is required")
	}

	return nil
}
