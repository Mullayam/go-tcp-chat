package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// OTPData holds OTP information
type OTPData struct {
	Code      string
	Email     string
	ExpiresAt time.Time
	Attempts  int
}

// OTPService manages OTP generation and validation
type OTPService struct {
	otps              map[string]*OTPData // key: email
	expirationMinutes int
	maxRetries        int
	mu                sync.RWMutex
}

// NewOTPService creates a new OTP service
func NewOTPService(expirationMinutes, maxRetries int) *OTPService {
	service := &OTPService{
		otps:              make(map[string]*OTPData),
		expirationMinutes: expirationMinutes,
		maxRetries:        maxRetries,
	}

	// Start cleanup goroutine
	go service.cleanupExpired()

	return service
}

// Generate generates a new 6-digit OTP for an email
func (s *OTPService) Generate(email string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate random 6-digit code
	code, err := s.generateCode()
	if err != nil {
		return "", err
	}

	// Store OTP data
	s.otps[email] = &OTPData{
		Code:      code,
		Email:     email,
		ExpiresAt: time.Now().Add(time.Duration(s.expirationMinutes) * time.Minute),
		Attempts:  0,
	}

	return code, nil
}

// Validate validates an OTP for an email
func (s *OTPService) Validate(email, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	otpData, exists := s.otps[email]
	if !exists {
		return fmt.Errorf("no OTP found for this email")
	}

	// Check expiration
	if time.Now().After(otpData.ExpiresAt) {
		delete(s.otps, email)
		return fmt.Errorf("OTP has expired")
	}

	// Check max attempts
	if otpData.Attempts >= s.maxRetries {
		delete(s.otps, email)
		return fmt.Errorf("maximum verification attempts exceeded")
	}

	// Increment attempts
	otpData.Attempts++

	// Validate code
	if otpData.Code != code {
		return fmt.Errorf("invalid OTP code")
	}

	// Success - delete OTP (one-time use)
	delete(s.otps, email)
	return nil
}

// generateCode generates a random 6-digit code
func (s *OTPService) generateCode() (string, error) {
	max := big.NewInt(1000000) // 0-999999
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// cleanupExpired periodically removes expired OTPs
func (s *OTPService) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for email, otpData := range s.otps {
			if now.After(otpData.ExpiresAt) {
				delete(s.otps, email)
			}
		}
		s.mu.Unlock()
	}
}

// HasPendingOTP checks if an email has a pending OTP
func (s *OTPService) HasPendingOTP(email string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	otpData, exists := s.otps[email]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(otpData.ExpiresAt) {
		return false
	}

	return true
}

// Clear removes OTP for an email
func (s *OTPService) Clear(email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.otps, email)
}
