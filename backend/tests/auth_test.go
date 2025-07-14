package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/migrations"
	"github.com/ccj241/cctrade/routes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestEnvironment() *gin.Engine {
	gin.SetMode(gin.TestMode)

	appConfig := config.LoadConfig()
	appConfig.Database.Database = "binance_trading_test"

	config.InitDatabase(appConfig)
	config.InitRedis(appConfig)

	migrations.RunMigrations()

	r := gin.New()
	routes.SetupRoutes(r)

	return r
}

func TestUserRegistration(t *testing.T) {
	router := setupTestEnvironment()

	registerData := map[string]interface{}{
		"username": "testuser123",
		"email":    "testuser123@example.com",
		"password": "Test123456",
	}

	jsonData, _ := json.Marshal(registerData)

	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
}

func TestUserLogin(t *testing.T) {
	router := setupTestEnvironment()

	registerData := map[string]interface{}{
		"username": "testuser456",
		"email":    "testuser456@example.com",
		"password": "Test123456",
	}

	jsonData, _ := json.Marshal(registerData)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	config.DB.Exec("UPDATE users SET status = 'active' WHERE username = 'testuser456'")

	loginData := map[string]interface{}{
		"username": "testuser456",
		"password": "Test123456",
	}

	jsonData, _ = json.Marshal(loginData)
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])

	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"])
}

func TestHealthCheck(t *testing.T) {
	router := setupTestEnvironment()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}
