package tests

import (
	"testing"

	"github.com/ccj241/cctrade/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user123@domain.org", true},
		{"invalid-email", false},
		{"@domain.com", false},
		{"", false},
	}

	for _, test := range tests {
		err := utils.ValidateEmail(test.email)
		if test.expected {
			assert.NoError(t, err, "Email %s should be valid", test.email)
		} else {
			assert.Error(t, err, "Email %s should be invalid", test.email)
		}
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		username string
		expected bool
	}{
		{"testuser", true},
		{"test_user123", true},
		{"test-user", true},
		{"a", false},
		{"", false},
		{"test user", false},
		{"test@user", false},
	}

	for _, test := range tests {
		err := utils.ValidateUsername(test.username)
		if test.expected {
			assert.NoError(t, err, "Username %s should be valid", test.username)
		} else {
			assert.Error(t, err, "Username %s should be invalid", test.username)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		expected bool
	}{
		{"Test123456", true},
		{"MyPassword1", true},
		{"test123", false},
		{"TEST123", false},
		{"Test", false},
		{"", false},
	}

	for _, test := range tests {
		err := utils.ValidatePassword(test.password)
		if test.expected {
			assert.NoError(t, err, "Password should be valid")
		} else {
			assert.Error(t, err, "Password should be invalid")
		}
	}
}

func TestEncryptDecryptAES(t *testing.T) {
	key := "your-32-character-key-here!!"
	plaintext := "test secret message"

	encrypted, err := utils.EncryptAES(plaintext, key)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := utils.DecryptAES(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestJWTOperations(t *testing.T) {
	token, err := utils.GenerateJWT(1, "testuser", "user")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := utils.ValidateJWT(token)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user", claims.Role)
}

func TestHelperFunctions(t *testing.T) {
	assert.Equal(t, "123.45", utils.FormatFloat(123.45, 2))
	assert.Equal(t, 123.46, utils.RoundTo(123.456, 2))

	uuid := utils.GenerateUUID()
	assert.NotEmpty(t, uuid)
	assert.Contains(t, uuid, "-")

	randomStr := utils.GenerateRandomString(10)
	assert.Len(t, randomStr, 10)

	assert.Equal(t, 50.0, utils.CalculatePercent(1, 2))
	assert.Equal(t, 100.0, utils.CalculatePercentChange(1, 2))
}
