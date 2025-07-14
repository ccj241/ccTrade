package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func FormatFloat(f float64, precision int) string {
	return strconv.FormatFloat(f, 'f', precision, 64)
}

func RoundTo(value float64, decimals int) float64 {
	pow10 := math.Pow(10, float64(decimals))
	return math.Round(value*pow10) / pow10
}

func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func ParseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func ToUpper(s string) string {
	return strings.ToUpper(s)
}

func ToLower(s string) string {
	return strings.ToLower(s)
}

func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func GenerateUUID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

func GenerateRandomString(length int) string {
	bytes := make([]byte, length/2)
	_, err := rand.Read(bytes)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)[:length]
}

func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func GetCurrentTimestampMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func FormatTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}

func ParseTimestamp(timeStr string) (int64, error) {
	t, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func CalculatePercent(value, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (value / total) * 100
}

func CalculatePercentChange(oldValue, newValue float64) float64 {
	if oldValue == 0 {
		return 0
	}
	return ((newValue - oldValue) / oldValue) * 100
}

func MinFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func MaxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func AbsFloat64(f float64) float64 {
	return math.Abs(f)
}

func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}
