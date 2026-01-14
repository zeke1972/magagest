// pkg/auth/auth.go

package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"ricambi-manager/internal/domain"
)

var (
	ErrInvalidToken    = errors.New("invalid authentication token")
	ErrTokenExpired    = errors.New("authentication token expired")
	ErrSessionNotFound = errors.New("session not found")
	ErrUnauthorized    = errors.New("unauthorized access")
)

type AuthService struct {
	sessionTimeout time.Duration
	sessions       map[string]*Session
}

type Session struct {
	Token      string
	OperatorID string
	Username   string
	Profile    domain.ProfileType
	CreatedAt  time.Time
	ExpiresAt  time.Time
	IPAddress  string
	UserAgent  string
}

func NewAuthService(sessionTimeoutMinutes int) *AuthService {
	return &AuthService{
		sessionTimeout: time.Duration(sessionTimeoutMinutes) * time.Minute,
		sessions:       make(map[string]*Session),
	}
}

func (s *AuthService) CreateSession(operator *domain.Operator, ipAddress, userAgent string) (*Session, error) {
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		Token:      token,
		OperatorID: operator.ID.Hex(),
		Username:   operator.Username,
		Profile:    operator.Profile,
		CreatedAt:  now,
		ExpiresAt:  now.Add(s.sessionTimeout),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}

	s.sessions[token] = session
	operator.CreateSession(token, s.sessionTimeout)

	return session, nil
}

func (s *AuthService) ValidateSession(token string) (*Session, error) {
	session, exists := s.sessions[token]
	if !exists {
		return nil, ErrSessionNotFound
	}

	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, token)
		return nil, ErrTokenExpired
	}

	return session, nil
}

func (s *AuthService) RefreshSession(token string) error {
	session, err := s.ValidateSession(token)
	if err != nil {
		return err
	}

	session.ExpiresAt = time.Now().Add(s.sessionTimeout)
	return nil
}

func (s *AuthService) InvalidateSession(token string) {
	delete(s.sessions, token)
}

func (s *AuthService) InvalidateAllSessions(operatorID string) {
	for token, session := range s.sessions {
		if session.OperatorID == operatorID {
			delete(s.sessions, token)
		}
	}
}

func (s *AuthService) GetActiveSessions(operatorID string) []*Session {
	var sessions []*Session
	now := time.Now()

	for _, session := range s.sessions {
		if session.OperatorID == operatorID && now.Before(session.ExpiresAt) {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

func (s *AuthService) CleanupExpiredSessions() int {
	now := time.Now()
	count := 0

	for token, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, token)
			count++
		}
	}

	return count
}

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

type PermissionChecker struct{}

func NewPermissionChecker() *PermissionChecker {
	return &PermissionChecker{}
}

func (pc *PermissionChecker) CheckPermission(operator *domain.Operator, area domain.PermissionArea, action domain.PermissionAction) error {
	if !operator.IsActive {
		return errors.New("operator is not active")
	}

	if operator.IsLocked {
		return domain.ErrOperatorLocked
	}

	if !operator.HasPermission(area, action) {
		return domain.ErrInsufficientPermissions
	}

	return nil
}

func (pc *PermissionChecker) RequireAdmin(operator *domain.Operator) error {
	if !operator.IsActive {
		return errors.New("operator is not active")
	}

	if operator.IsLocked {
		return domain.ErrOperatorLocked
	}

	if !operator.IsAdmin() {
		return errors.New("admin privileges required")
	}

	return nil
}

func (pc *PermissionChecker) CanViewSensitiveData(operator *domain.Operator, dataType string) bool {
	if operator.IsAdmin() {
		return true
	}

	switch dataType {
	case "customer_financial":
		return operator.HasPermission(domain.AreaAccounting, domain.ActionView) ||
			operator.HasPermission(domain.AreaCustomers, domain.ActionView)
	case "pricing":
		return operator.HasPermission(domain.AreaCommercial, domain.ActionView) ||
			operator.HasPermission(domain.AreaArticles, domain.ActionView)
	case "supplier_costs":
		return operator.HasPermission(domain.AreaAccounting, domain.ActionView) ||
			operator.HasPermission(domain.AreaSuppliers, domain.ActionView)
	case "statistics":
		return operator.HasPermission(domain.AreaStatistics, domain.ActionView)
	default:
		return false
	}
}

func (pc *PermissionChecker) CanApproveDiscount(operator *domain.Operator, discountPercent float64) bool {
	if operator.IsAdmin() {
		return true
	}

	if operator.Profile == domain.ProfileSales && discountPercent <= 20 {
		return true
	}

	return false
}

func (pc *PermissionChecker) CanOverrideFido(operator *domain.Operator) bool {
	return operator.CanOverrideFido()
}

func (pc *PermissionChecker) CanApproveSottocosto(operator *domain.Operator) bool {
	return operator.CanApproveSottocosto()
}

type AuditLogger struct{}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}

func (al *AuditLogger) LogAction(operator *domain.Operator, action, area, resourceID, details, ipAddress string) {
	if operator == nil {
		return
	}
	operator.AddAuditEntry(action, area, resourceID, details, ipAddress)
}

func (al *AuditLogger) LogSensitiveAction(operator *domain.Operator, action, area, resourceID, details, ipAddress string) {
	if operator == nil {
		return
	}

	sensitiveDetails := "[SENSITIVE] " + details
	operator.AddAuditEntry(action, area, resourceID, sensitiveDetails, ipAddress)
}

func (al *AuditLogger) LogFailedAccess(username, action, area, reason, ipAddress string) {
}

type RateLimiter struct {
	attempts       map[string][]time.Time
	maxAttempts    int
	windowDuration time.Duration
}

func NewRateLimiter(maxAttempts int, windowMinutes int) *RateLimiter {
	return &RateLimiter{
		attempts:       make(map[string][]time.Time),
		maxAttempts:    maxAttempts,
		windowDuration: time.Duration(windowMinutes) * time.Minute,
	}
}

func (rl *RateLimiter) CheckLimit(identifier string) (bool, int) {
	now := time.Now()
	windowStart := now.Add(-rl.windowDuration)

	attempts, exists := rl.attempts[identifier]
	if !exists {
		rl.attempts[identifier] = []time.Time{now}
		return true, 1
	}

	var recentAttempts []time.Time
	for _, attempt := range attempts {
		if attempt.After(windowStart) {
			recentAttempts = append(recentAttempts, attempt)
		}
	}

	if len(recentAttempts) >= rl.maxAttempts {
		rl.attempts[identifier] = recentAttempts
		return false, len(recentAttempts)
	}

	recentAttempts = append(recentAttempts, now)
	rl.attempts[identifier] = recentAttempts
	return true, len(recentAttempts)
}

func (rl *RateLimiter) Reset(identifier string) {
	delete(rl.attempts, identifier)
}

func (rl *RateLimiter) Cleanup() {
	now := time.Now()
	windowStart := now.Add(-rl.windowDuration)

	for identifier, attempts := range rl.attempts {
		var recentAttempts []time.Time
		for _, attempt := range attempts {
			if attempt.After(windowStart) {
				recentAttempts = append(recentAttempts, attempt)
			}
		}

		if len(recentAttempts) == 0 {
			delete(rl.attempts, identifier)
		} else {
			rl.attempts[identifier] = recentAttempts
		}
	}
}

type PasswordValidator struct {
	minLength        int
	requireUppercase bool
	requireLowercase bool
	requireDigit     bool
	requireSpecial   bool
}

func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		minLength:        8,
		requireUppercase: true,
		requireLowercase: true,
		requireDigit:     true,
		requireSpecial:   false,
	}
}

func (pv *PasswordValidator) Validate(password string) error {
	if len(password) < pv.minLength {
		return errors.New("password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= '!' && char <= '/' || char >= ':' && char <= '@' || char >= '[' && char <= '`' || char >= '{' && char <= '~':
			hasSpecial = true
		}
	}

	if pv.requireUppercase && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if pv.requireLowercase && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	if pv.requireDigit && !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	if pv.requireSpecial && !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

func (pv *PasswordValidator) GenerateRandomPassword(length int) (string, error) {
	if length < pv.minLength {
		length = pv.minLength
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)

	if _, err := rand.Read(password); err != nil {
		return "", err
	}

	for i := range password {
		password[i] = charset[int(password[i])%len(charset)]
	}

	return string(password), nil
}
