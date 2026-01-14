// internal/domain/operator.go

package domain

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrOperatorNotFound        = errors.New("operator not found")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrOperatorLocked          = errors.New("operator account is locked")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrInvalidPassword         = errors.New("invalid password format")
)

type ProfileType string

const (
	ProfileAdmin      ProfileType = "admin"
	ProfileWarehouse  ProfileType = "warehouse"
	ProfileSales      ProfileType = "sales"
	ProfileAccounting ProfileType = "accounting"
)

type PermissionArea string

const (
	AreaAccounting PermissionArea = "accounting"
	AreaStatistics PermissionArea = "statistics"
	AreaCommercial PermissionArea = "commercial"
	AreaWarehouse  PermissionArea = "warehouse"
	AreaCustomers  PermissionArea = "customers"
	AreaSuppliers  PermissionArea = "suppliers"
	AreaArticles   PermissionArea = "articles"
	AreaOrders     PermissionArea = "orders"
	AreaDocuments  PermissionArea = "documents"
	AreaReports    PermissionArea = "reports"
	AreaSettings   PermissionArea = "settings"
)

type PermissionAction string

const (
	ActionView    PermissionAction = "view"
	ActionCreate  PermissionAction = "create"
	ActionEdit    PermissionAction = "edit"
	ActionDelete  PermissionAction = "delete"
	ActionExport  PermissionAction = "export"
	ActionApprove PermissionAction = "approve"
)

type Operator struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username           string             `bson:"username" json:"username"`
	PasswordHash       string             `bson:"password_hash" json:"-"`
	FullName           string             `bson:"full_name" json:"full_name"`
	Email              string             `bson:"email" json:"email"`
	Profile            ProfileType        `bson:"profile" json:"profile"`
	Permissions        []Permission       `bson:"permissions" json:"permissions"`
	IsActive           bool               `bson:"is_active" json:"is_active"`
	IsLocked           bool               `bson:"is_locked" json:"is_locked"`
	FailedAttempts     int                `bson:"failed_attempts" json:"failed_attempts"`
	LastFailedAttempt  time.Time          `bson:"last_failed_attempt" json:"last_failed_attempt"`
	LastLogin          time.Time          `bson:"last_login" json:"last_login"`
	LastPasswordChange time.Time          `bson:"last_password_change" json:"last_password_change"`
	SessionToken       string             `bson:"session_token" json:"-"`
	SessionExpiry      time.Time          `bson:"session_expiry" json:"session_expiry"`
	AuditLog           []AuditEntry       `bson:"audit_log" json:"audit_log"`
	Settings           OperatorSettings   `bson:"settings" json:"settings"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedBy          string             `bson:"created_by" json:"created_by"`
}

type Permission struct {
	Area    PermissionArea     `bson:"area" json:"area"`
	Actions []PermissionAction `bson:"actions" json:"actions"`
}

type AuditEntry struct {
	ID         primitive.ObjectID `bson:"id" json:"id"`
	Action     string             `bson:"action" json:"action"`
	Area       string             `bson:"area" json:"area"`
	ResourceID string             `bson:"resource_id" json:"resource_id"`
	Details    string             `bson:"details" json:"details"`
	IPAddress  string             `bson:"ip_address" json:"ip_address"`
	Timestamp  time.Time          `bson:"timestamp" json:"timestamp"`
}

type OperatorSettings struct {
	Theme              string `bson:"theme" json:"theme"`
	Language           string `bson:"language" json:"language"`
	PageSize           int    `bson:"page_size" json:"page_size"`
	NotificationsEmail bool   `bson:"notifications_email" json:"notifications_email"`
}

func NewOperator(username, fullName, email, password string, profile ProfileType, createdBy string) (*Operator, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username cannot be empty")
	}
	if strings.TrimSpace(password) == "" || len(password) < 8 {
		return nil, ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	operator := &Operator{
		ID:             primitive.NewObjectID(),
		Username:       strings.ToLower(strings.TrimSpace(username)),
		PasswordHash:   string(hashedPassword),
		FullName:       strings.TrimSpace(fullName),
		Email:          strings.ToLower(strings.TrimSpace(email)),
		Profile:        profile,
		Permissions:    GetDefaultPermissions(profile),
		IsActive:       true,
		IsLocked:       false,
		FailedAttempts: 0,
		AuditLog:       []AuditEntry{},
		Settings: OperatorSettings{
			Theme:              "default",
			Language:           "it",
			PageSize:           20,
			NotificationsEmail: true,
		},
		CreatedAt:          now,
		UpdatedAt:          now,
		LastPasswordChange: now,
		CreatedBy:          createdBy,
	}

	return operator, nil
}

func GetDefaultPermissions(profile ProfileType) []Permission {
	switch profile {
	case ProfileAdmin:
		return []Permission{
			{Area: AreaAccounting, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit, ActionDelete, ActionExport, ActionApprove}},
			{Area: AreaStatistics, Actions: []PermissionAction{ActionView, ActionExport}},
			{Area: AreaCommercial, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit, ActionDelete}},
			{Area: AreaWarehouse, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit, ActionDelete}},
			{Area: AreaCustomers, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit, ActionDelete, ActionExport}},
			{Area: AreaSuppliers, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit, ActionDelete, ActionExport}},
			{Area: AreaArticles, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit, ActionDelete, ActionExport}},
			{Area: AreaOrders, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit, ActionDelete, ActionApprove}},
			{Area: AreaDocuments, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit, ActionDelete}},
			{Area: AreaReports, Actions: []PermissionAction{ActionView, ActionExport}},
			{Area: AreaSettings, Actions: []PermissionAction{ActionView, ActionEdit}},
		}
	case ProfileWarehouse:
		return []Permission{
			{Area: AreaWarehouse, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit}},
			{Area: AreaArticles, Actions: []PermissionAction{ActionView, ActionEdit}},
			{Area: AreaSuppliers, Actions: []PermissionAction{ActionView}},
			{Area: AreaOrders, Actions: []PermissionAction{ActionView}},
		}
	case ProfileSales:
		return []Permission{
			{Area: AreaCommercial, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit}},
			{Area: AreaCustomers, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit}},
			{Area: AreaArticles, Actions: []PermissionAction{ActionView}},
			{Area: AreaOrders, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit}},
			{Area: AreaWarehouse, Actions: []PermissionAction{ActionView}},
			{Area: AreaStatistics, Actions: []PermissionAction{ActionView}},
		}
	case ProfileAccounting:
		return []Permission{
			{Area: AreaAccounting, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit, ActionExport}},
			{Area: AreaCustomers, Actions: []PermissionAction{ActionView, ActionEdit}},
			{Area: AreaSuppliers, Actions: []PermissionAction{ActionView, ActionEdit}},
			{Area: AreaDocuments, Actions: []PermissionAction{ActionView, ActionCreate, ActionEdit}},
			{Area: AreaReports, Actions: []PermissionAction{ActionView, ActionExport}},
			{Area: AreaStatistics, Actions: []PermissionAction{ActionView, ActionExport}},
		}
	default:
		return []Permission{}
	}
}

func (o *Operator) Validate() error {
	if strings.TrimSpace(o.Username) == "" {
		return errors.New("username cannot be empty")
	}
	if strings.TrimSpace(o.FullName) == "" {
		return errors.New("full name cannot be empty")
	}
	return nil
}

func (o *Operator) CheckPassword(password string) error {
	if o.IsLocked {
		return ErrOperatorLocked
	}
	if !o.IsActive {
		return errors.New("operator is not active")
	}

	err := bcrypt.CompareHashAndPassword([]byte(o.PasswordHash), []byte(password))
	if err != nil {
		o.FailedAttempts++
		o.LastFailedAttempt = time.Now()
		if o.FailedAttempts >= 5 {
			o.IsLocked = true
		}
		o.UpdatedAt = time.Now()
		return ErrInvalidCredentials
	}

	o.FailedAttempts = 0
	o.LastLogin = time.Now()
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Operator) ChangePassword(oldPassword, newPassword string) error {
	if err := o.CheckPassword(oldPassword); err != nil {
		return err
	}

	if len(newPassword) < 8 {
		return ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	o.PasswordHash = string(hashedPassword)
	o.LastPasswordChange = time.Now()
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Operator) ResetPassword(newPassword string) error {
	if len(newPassword) < 8 {
		return ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	o.PasswordHash = string(hashedPassword)
	o.LastPasswordChange = time.Now()
	o.FailedAttempts = 0
	o.IsLocked = false
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Operator) HasPermission(area PermissionArea, action PermissionAction) bool {
	for _, perm := range o.Permissions {
		if perm.Area == area {
			for _, act := range perm.Actions {
				if act == action {
					return true
				}
			}
		}
	}
	return false
}

func (o *Operator) GrantPermission(area PermissionArea, action PermissionAction) {
	for i, perm := range o.Permissions {
		if perm.Area == area {
			for _, act := range perm.Actions {
				if act == action {
					return
				}
			}
			o.Permissions[i].Actions = append(o.Permissions[i].Actions, action)
			o.UpdatedAt = time.Now()
			return
		}
	}

	o.Permissions = append(o.Permissions, Permission{
		Area:    area,
		Actions: []PermissionAction{action},
	})
	o.UpdatedAt = time.Now()
}

func (o *Operator) RevokePermission(area PermissionArea, action PermissionAction) {
	for i, perm := range o.Permissions {
		if perm.Area == area {
			for j, act := range perm.Actions {
				if act == action {
					o.Permissions[i].Actions = append(perm.Actions[:j], perm.Actions[j+1:]...)
					o.UpdatedAt = time.Now()
					return
				}
			}
		}
	}
}

func (o *Operator) AddAuditEntry(action, area, resourceID, details, ipAddress string) {
	entry := AuditEntry{
		ID:         primitive.NewObjectID(),
		Action:     action,
		Area:       area,
		ResourceID: resourceID,
		Details:    details,
		IPAddress:  ipAddress,
		Timestamp:  time.Now(),
	}

	o.AuditLog = append(o.AuditLog, entry)

	if len(o.AuditLog) > 1000 {
		o.AuditLog = o.AuditLog[len(o.AuditLog)-1000:]
	}
}

func (o *Operator) GetRecentAuditLog(limit int) []AuditEntry {
	if limit <= 0 || limit > len(o.AuditLog) {
		limit = len(o.AuditLog)
	}

	result := make([]AuditEntry, limit)
	copy(result, o.AuditLog[len(o.AuditLog)-limit:])

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

func (o *Operator) Lock() {
	o.IsLocked = true
	o.UpdatedAt = time.Now()
}

func (o *Operator) Unlock() {
	o.IsLocked = false
	o.FailedAttempts = 0
	o.UpdatedAt = time.Now()
}

func (o *Operator) Deactivate() {
	o.IsActive = false
	o.SessionToken = ""
	o.SessionExpiry = time.Time{}
	o.UpdatedAt = time.Now()
}

func (o *Operator) Activate() {
	o.IsActive = true
	o.UpdatedAt = time.Now()
}

func (o *Operator) CreateSession(token string, duration time.Duration) {
	o.SessionToken = token
	o.SessionExpiry = time.Now().Add(duration)
	o.LastLogin = time.Now()
	o.UpdatedAt = time.Now()
}

func (o *Operator) IsSessionValid() bool {
	if o.SessionToken == "" {
		return false
	}
	return time.Now().Before(o.SessionExpiry)
}

func (o *Operator) InvalidateSession() {
	o.SessionToken = ""
	o.SessionExpiry = time.Time{}
	o.UpdatedAt = time.Now()
}

func (o *Operator) IsAdmin() bool {
	return o.Profile == ProfileAdmin
}

func (o *Operator) CanOverrideFido() bool {
	return o.Profile == ProfileAdmin
}

func (o *Operator) CanApproveSottocosto() bool {
	return o.Profile == ProfileAdmin || o.Profile == ProfileSales
}
