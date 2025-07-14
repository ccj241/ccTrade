package utils

import (
	"errors"
	"regexp"
	"strings"
)

func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("邮箱不能为空")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("邮箱格式不正确")
	}

	return nil
}

func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("用户名不能为空")
	}

	if len(username) < 3 || len(username) > 50 {
		return errors.New("用户名长度必须在3-50字符之间")
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		return errors.New("用户名只能包含字母、数字、下划线和连字符")
	}

	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("密码不能为空")
	}

	if len(password) < 8 {
		return errors.New("密码长度至少8位")
	}

	if len(password) > 100 {
		return errors.New("密码长度不能超过100位")
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)

	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("密码必须包含大写字母、小写字母和数字")
	}

	return nil
}

func ValidateSymbol(symbol string) error {
	if symbol == "" {
		return errors.New("交易对不能为空")
	}

	symbol = strings.ToUpper(symbol)
	symbolRegex := regexp.MustCompile(`^[A-Z0-9]{6,20}$`)
	if !symbolRegex.MatchString(symbol) {
		return errors.New("交易对格式不正确")
	}

	return nil
}

func ValidateQuantity(quantity float64) error {
	if quantity <= 0 {
		return errors.New("数量必须大于0")
	}

	return nil
}

func ValidatePrice(price float64) error {
	if price <= 0 {
		return errors.New("价格必须大于0")
	}

	return nil
}

func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return errors.New("API密钥不能为空")
	}

	if len(apiKey) < 32 || len(apiKey) > 128 {
		return errors.New("API密钥长度不正确")
	}

	return nil
}

func ValidateSecretKey(secretKey string) error {
	if secretKey == "" {
		return errors.New("API密钥不能为空")
	}

	if len(secretKey) < 32 || len(secretKey) > 128 {
		return errors.New("API密钥长度不正确")
	}

	return nil
}
