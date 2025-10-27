package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/application/services"
)

// MockEntryCounter for key validator tests
type MockEntryCounter struct {
	mock.Mock
}

func (m *MockEntryCounter) CountByOwnerAndType(ctx context.Context, ownerTaxID string, keyType services.KeyType) (int, error) {
	args := m.Called(ctx, ownerTaxID, keyType)
	return args.Int(0), args.Error(1)
}

// ===== CPF VALIDATION TESTS =====

// Test 1: TestKeyValidator_ValidateCPF_Success
func TestKeyValidator_ValidateCPF_Success(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	// Valid CPF: 123.456.789-09 (without formatting)
	validCPF := "12345678909"

	// Act
	err := validator.ValidateFormat(services.KeyTypeCPF, validCPF)

	// Assert
	require.NoError(t, err)
}

// Test 2: TestKeyValidator_ValidateCPF_InvalidLength
func TestKeyValidator_ValidateCPF_InvalidLength(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	// Act
	err := validator.ValidateFormat(services.KeyTypeCPF, "123456789")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "11 digits")
}

// Test 3: TestKeyValidator_ValidateCPF_InvalidPattern
func TestKeyValidator_ValidateCPF_InvalidPattern(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	testCases := []string{
		"00000000000",
		"11111111111",
		"22222222222",
		"99999999999",
	}

	for _, cpf := range testCases {
		// Act
		err := validator.ValidateFormat(services.KeyTypeCPF, cpf)

		// Assert
		require.Error(t, err, "CPF %s should be invalid", cpf)
		assert.Contains(t, err.Error(), "invalid")
	}
}

// Test 4: TestKeyValidator_ValidateCPF_InvalidCheckDigits
func TestKeyValidator_ValidateCPF_InvalidCheckDigits(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	// Invalid check digits
	invalidCPF := "12345678900"

	// Act
	err := validator.ValidateFormat(services.KeyTypeCPF, invalidCPF)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "check digits")
}

// ===== CNPJ VALIDATION TESTS =====

// Test 5: TestKeyValidator_ValidateCNPJ_Success
func TestKeyValidator_ValidateCNPJ_Success(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	// Valid CNPJ: 11.222.333/0001-81 (without formatting)
	validCNPJ := "11222333000181"

	// Act
	err := validator.ValidateFormat(services.KeyTypeCNPJ, validCNPJ)

	// Assert
	require.NoError(t, err)
}

// Test 6: TestKeyValidator_ValidateCNPJ_InvalidLength
func TestKeyValidator_ValidateCNPJ_InvalidLength(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	// Act
	err := validator.ValidateFormat(services.KeyTypeCNPJ, "1122233300018")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "14 digits")
}

// ===== EMAIL VALIDATION TESTS =====

// Test 7: TestKeyValidator_ValidateEmail_Success
func TestKeyValidator_ValidateEmail_Success(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	validEmails := []string{
		"user@example.com",
		"test.user@domain.co.br",
		"user+tag@example.com",
		"user123@test-domain.com",
	}

	for _, email := range validEmails {
		// Act
		err := validator.ValidateFormat(services.KeyTypeEmail, email)

		// Assert
		require.NoError(t, err, "Email %s should be valid", email)
	}
}

// Test 8: TestKeyValidator_ValidateEmail_Invalid
func TestKeyValidator_ValidateEmail_Invalid(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	invalidEmails := []string{
		"invalid",
		"@example.com",
		"user@",
		"user @example.com",
		"user@.com",
	}

	for _, email := range invalidEmails {
		// Act
		err := validator.ValidateFormat(services.KeyTypeEmail, email)

		// Assert
		require.Error(t, err, "Email %s should be invalid", email)
		assert.Contains(t, err.Error(), "invalid email")
	}
}

// ===== PHONE VALIDATION TESTS =====

// Test 9: TestKeyValidator_ValidatePhone_Success
func TestKeyValidator_ValidatePhone_Success(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	validPhones := []string{
		"+5511999998888",
		"+5521987654321",
		"+5585912345678",
	}

	for _, phone := range validPhones {
		// Act
		err := validator.ValidateFormat(services.KeyTypePhone, phone)

		// Assert
		require.NoError(t, err, "Phone %s should be valid", phone)
	}
}

// Test 10: TestKeyValidator_ValidatePhone_Invalid
func TestKeyValidator_ValidatePhone_Invalid(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	invalidPhones := []string{
		"11999998888",            // Missing +55
		"+55119999988",           // Too short
		"+551199999888888",       // Too long
		"+5501999998888",         // Invalid DDD (01)
	}

	for _, phone := range invalidPhones {
		// Act
		err := validator.ValidateFormat(services.KeyTypePhone, phone)

		// Assert
		require.Error(t, err, "Phone %s should be invalid", phone)
		assert.Contains(t, err.Error(), "invalid phone")
	}
}

// ===== EVP VALIDATION TESTS =====

// Test 11: TestKeyValidator_ValidateEVP_Success
func TestKeyValidator_ValidateEVP_Success(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	// Valid UUID v4
	validEVP := "550e8400-e29b-41d4-a716-446655440000"

	// Act
	err := validator.ValidateFormat(services.KeyTypeEVP, validEVP)

	// Assert
	require.NoError(t, err)
}

// Test 12: TestKeyValidator_ValidateEVP_Invalid
func TestKeyValidator_ValidateEVP_Invalid(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	invalidEVPs := []string{
		"not-a-uuid",
		"550e8400-e29b-31d4-a716-446655440000", // UUID v3, not v4
		"550e8400e29b41d4a716446655440000",     // Missing dashes
	}

	for _, evp := range invalidEVPs {
		// Act
		err := validator.ValidateFormat(services.KeyTypeEVP, evp)

		// Assert
		require.Error(t, err, "EVP %s should be invalid", evp)
		assert.Contains(t, err.Error(), "invalid EVP")
	}
}

// ===== LIMITS VALIDATION TESTS =====

// Test 13: TestKeyValidator_ValidateLimits_Success
func TestKeyValidator_ValidateLimits_Success(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	ownerTaxID := "12345678901"
	mockCounter.On("CountByOwnerAndType", mock.Anything, ownerTaxID, services.KeyTypeCPF).Return(4, nil)

	// Act
	err := validator.ValidateLimits(context.Background(), services.KeyTypeCPF, ownerTaxID)

	// Assert
	require.NoError(t, err)
	mockCounter.AssertExpectations(t)
}

// Test 14: TestKeyValidator_ValidateLimits_Exceeded
func TestKeyValidator_ValidateLimits_Exceeded(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	ownerTaxID := "12345678901"
	// CPF limit is 5, user already has 5
	mockCounter.On("CountByOwnerAndType", mock.Anything, ownerTaxID, services.KeyTypeCPF).Return(5, nil)

	// Act
	err := validator.ValidateLimits(context.Background(), services.KeyTypeCPF, ownerTaxID)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key limit exceeded")
	mockCounter.AssertExpectations(t)
}

// Test 15: TestKeyValidator_ValidateLimits_CountError
func TestKeyValidator_ValidateLimits_CountError(t *testing.T) {
	// Arrange
	mockCounter := new(MockEntryCounter)
	validator := services.NewKeyValidatorService(mockCounter)

	ownerTaxID := "12345678901"
	mockCounter.On("CountByOwnerAndType", mock.Anything, ownerTaxID, services.KeyTypeCPF).Return(0, errors.New("database error"))

	// Act
	err := validator.ValidateLimits(context.Background(), services.KeyTypeCPF, ownerTaxID)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to count")
	mockCounter.AssertExpectations(t)
}
