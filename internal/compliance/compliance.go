package compliance

import (
	"fmt"
	"time"
)

// ComplianceLevel represents the level of KYC verification
type ComplianceLevel string

const (
	LevelNone     ComplianceLevel = "none"
	LevelBasic    ComplianceLevel = "basic"
	LevelEnhanced ComplianceLevel = "enhanced"
)

// UserVerification represents user KYC information
type UserVerification struct {
	UserID          string          `json:"user_id"`
	Level           ComplianceLevel `json:"level"`
	VerifiedAt      time.Time       `json:"verified_at"`
	ExpiresAt       time.Time       `json:"expires_at"`
	DocumentType    string          `json:"document_type,omitempty"`
	DocumentNumber  string          `json:"document_number,omitempty"`
	DocumentCountry string          `json:"document_country,omitempty"`
	RiskScore       float64         `json:"risk_score"`
}

// TransactionRisk represents risk assessment for a transaction
type TransactionRisk struct {
	TransactionID string    `json:"transaction_id"`
	RiskScore     float64   `json:"risk_score"`
	RiskFactors   []string  `json:"risk_factors"`
	AssessedAt    time.Time `json:"assessed_at"`
}

// ComplianceManager handles compliance-related operations
type ComplianceManager struct {
	verifications map[string]*UserVerification
	transactions  map[string]*TransactionRisk
}

// NewComplianceManager creates a new compliance manager
func NewComplianceManager() *ComplianceManager {
	return &ComplianceManager{
		verifications: make(map[string]*UserVerification),
		transactions:  make(map[string]*TransactionRisk),
	}
}

// VerifyUser performs KYC verification
func (cm *ComplianceManager) VerifyUser(userID string, level ComplianceLevel, documentInfo map[string]string) (*UserVerification, error) {
	verification := &UserVerification{
		UserID:          userID,
		Level:           level,
		VerifiedAt:      time.Now(),
		ExpiresAt:       time.Now().AddDate(1, 0, 0), // 1 year validity
		DocumentType:    documentInfo["type"],
		DocumentNumber:  documentInfo["number"],
		DocumentCountry: documentInfo["country"],
		RiskScore:       calculateRiskScore(level, documentInfo),
	}

	cm.verifications[userID] = verification
	return verification, nil
}

// AssessTransactionRisk evaluates transaction risk
func (cm *ComplianceManager) AssessTransactionRisk(txID string, amount float64, senderID, receiverID string) (*TransactionRisk, error) {
	risk := &TransactionRisk{
		TransactionID: txID,
		AssessedAt:    time.Now(),
	}

	// Get user risk scores
	senderRisk := 0.0
	if sender, exists := cm.verifications[senderID]; exists {
		senderRisk = sender.RiskScore
	}

	receiverRisk := 0.0
	if receiver, exists := cm.verifications[receiverID]; exists {
		receiverRisk = receiver.RiskScore
	}

	// Calculate transaction risk
	risk.RiskScore = calculateTransactionRisk(amount, senderRisk, receiverRisk)
	risk.RiskFactors = identifyRiskFactors(amount, senderRisk, receiverRisk)

	cm.transactions[txID] = risk
	return risk, nil
}

// GetUserVerification retrieves user verification status
func (cm *ComplianceManager) GetUserVerification(userID string) (*UserVerification, error) {
	if verification, exists := cm.verifications[userID]; exists {
		return verification, nil
	}
	return nil, fmt.Errorf("no verification found for user %s", userID)
}

// GetTransactionRisk retrieves transaction risk assessment
func (cm *ComplianceManager) GetTransactionRisk(txID string) (*TransactionRisk, error) {
	if risk, exists := cm.transactions[txID]; exists {
		return risk, nil
	}
	return nil, fmt.Errorf("no risk assessment found for transaction %s", txID)
}

// Helper functions

func calculateRiskScore(level ComplianceLevel, documentInfo map[string]string) float64 {
	baseScore := 0.0
	switch level {
	case LevelNone:
		baseScore = 1.0
	case LevelBasic:
		baseScore = 0.5
	case LevelEnhanced:
		baseScore = 0.2
	}

	// Additional risk factors
	if documentInfo["country"] == "high_risk_country" {
		baseScore += 0.3
	}

	return baseScore
}

func calculateTransactionRisk(amount float64, senderRisk, receiverRisk float64) float64 {
	// Base risk from amount
	amountRisk := amount / 10000.0 // Normalize to 0-1 scale

	// Combine risks
	combinedRisk := (amountRisk + senderRisk + receiverRisk) / 3.0

	// Cap at 1.0
	if combinedRisk > 1.0 {
		combinedRisk = 1.0
	}

	return combinedRisk
}

func identifyRiskFactors(amount float64, senderRisk, receiverRisk float64) []string {
	factors := make([]string, 0)

	if amount > 10000 {
		factors = append(factors, "high_value_transaction")
	}
	if senderRisk > 0.7 {
		factors = append(factors, "high_risk_sender")
	}
	if receiverRisk > 0.7 {
		factors = append(factors, "high_risk_receiver")
	}

	return factors
}
