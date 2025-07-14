package services

import (
	"context"
	"errors"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/stretchr/testify/assert"
)

func TestHandleBinanceError(t *testing.T) {
	bs := &BinanceService{}

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "timestamp error -1021",
			err:      &binance.APIError{Code: -1021, Message: "Timestamp for this request is outside of the recvWindow."},
			expected: "timestamp error: request timestamp outside recvWindow (server time diff too large)",
		},
		{
			name:     "invalid signature -1022",
			err:      &binance.APIError{Code: -1022, Message: "Signature for this request is not valid."},
			expected: "invalid signature: API secret key may be incorrect",
		},
		{
			name:     "invalid API key -2008",
			err:      &binance.APIError{Code: -2008, Message: "Invalid Api-Key ID."},
			expected: "invalid API key: API key format is incorrect",
		},
		{
			name:     "API key not exist -2014",
			err:      &binance.APIError{Code: -2014, Message: "API-key format invalid."},
			expected: "invalid API key: API key does not exist",
		},
		{
			name:     "API key not activated -2015",
			err:      &binance.APIError{Code: -2015, Message: "Invalid API-key, IP, or permissions for action."},
			expected: "invalid API key: API key is not activated or IP not whitelisted",
		},
		{
			name:     "unknown error -1000",
			err:      &binance.APIError{Code: -1000, Message: "An unknown error occurred while processing the request."},
			expected: "unknown error: please check Binance API status",
		},
		{
			name:     "rate limit -1003",
			err:      &binance.APIError{Code: -1003, Message: "Too many requests."},
			expected: ErrRateLimitExceeded.Error(),
		},
		{
			name:     "insufficient balance -2010",
			err:      &binance.APIError{Code: -2010, Message: "Account has insufficient balance for requested action."},
			expected: ErrInsufficientBalance.Error(),
		},
		{
			name:     "order not found -2011",
			err:      &binance.APIError{Code: -2011, Message: "Order was not found."},
			expected: ErrOrderNotFound.Error(),
		},
		{
			name:     "generic API error",
			err:      &binance.APIError{Code: -9999, Message: "Some other error"},
			expected: "binance API error: Some other error (code: -9999)",
		},
		{
			name:     "string error with APIError tag",
			err:      errors.New("<APIError> code=-1022"),
			expected: "binance API error: <APIError> code=-1022 (check API credentials and permissions)",
		},
		{
			name:     "generic error",
			err:      errors.New("network timeout"),
			expected: "binance API error: network timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bs.handleBinanceError(tt.err)
			if tt.expected == "" {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, result.Error())
			}
		})
	}
}

func TestDiagnoseAPIConnection(t *testing.T) {
	// 这个测试需要模拟BinanceService，这里仅作为示例
	t.Skip("Requires mock implementation")

	bs := &BinanceService{
		testNet: true,
	}

	ctx := context.Background()
	diagnosis, err := bs.DiagnoseAPIConnection(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, diagnosis)

	// 验证诊断结果包含必要的字段
	requiredFields := []string{
		"encryption_key_length",
		"encryption_key_valid",
		"api_key_present",
		"secret_key_present",
		"testnet_enabled",
	}

	for _, field := range requiredFields {
		assert.Contains(t, diagnosis, field, "Missing required field: %s", field)
	}
}