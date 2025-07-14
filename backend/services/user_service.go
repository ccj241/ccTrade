package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/models"
	"github.com/ccj241/cctrade/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{db: config.DB}
}

func (us *UserService) CreateUser(username, email, password string) (*models.User, error) {
	if err := utils.ValidateUsername(username); err != nil {
		return nil, err
	}
	if err := utils.ValidateEmail(email); err != nil {
		return nil, err
	}
	if err := utils.ValidatePassword(password); err != nil {
		return nil, err
	}

	var existingUser models.User
	if err := us.db.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名或邮箱已存在")
	}

	user := &models.User{
		Username: username,
		Email:    email,
		Role:     models.RoleUser,
		Status:   models.StatusPending,
	}

	if err := user.HashPassword(password); err != nil {
		return nil, err
	}

	if err := us.db.Create(user).Error; err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (us *UserService) LoginUser(username, password string) (*models.User, string, error) {
	var user models.User
	if err := us.db.Where("username = ? OR email = ?", username, username).First(&user).Error; err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	if user.Status != models.StatusActive {
		return nil, "", errors.New("账户未激活或已被禁用")
	}

	if err := user.CheckPassword(password); err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	token, err := utils.GenerateJWT(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	user.LastLoginAt = &now
	us.db.Save(&user)

	user.Password = ""
	user.HasAPIKey = user.APIKey != "" && user.SecretKey != ""
	return &user, token, nil
}

func (us *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := us.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	
	// 设置HasAPIKey字段，用于前端显示
	// 如果API密钥已加密，只要密钥字段不为空就表示已配置
	if user.IsEncrypted {
		user.HasAPIKey = user.APIKey != "" && user.SecretKey != ""
	} else {
		// 未加密的情况下，直接判断
		user.HasAPIKey = user.APIKey != "" && user.SecretKey != ""
	}
	
	// 清除敏感信息
	user.Password = ""
	user.APIKey = ""
	user.SecretKey = ""
	
	return &user, nil
}

func (us *UserService) UpdateUserAPIKeys(userID uint, apiKey, secretKey string) error {
	// 清理API密钥中的空白字符和换行符
	apiKey = strings.TrimSpace(apiKey)
	secretKey = strings.TrimSpace(secretKey)
	
	// 移除所有换行符和回车符
	apiKey = strings.ReplaceAll(apiKey, "\n", "")
	apiKey = strings.ReplaceAll(apiKey, "\r", "")
	secretKey = strings.ReplaceAll(secretKey, "\n", "")
	secretKey = strings.ReplaceAll(secretKey, "\r", "")
	
	if err := utils.ValidateAPIKey(apiKey); err != nil {
		return err
	}
	if err := utils.ValidateSecretKey(secretKey); err != nil {
		return err
	}

	encryptedAPIKey, err := utils.EncryptAES(apiKey, config.AppConfig.Security.EncryptionKey)
	if err != nil {
		return err
	}

	encryptedSecretKey, err := utils.EncryptAES(secretKey, config.AppConfig.Security.EncryptionKey)
	if err != nil {
		return err
	}

	// 验证步骤是可选的，先尝试验证，如果失败也允许保存
	// 用户可以通过单独的验证接口来验证密钥
	shouldValidate := false // 暂时禁用自动验证，避免保存时出错
	
	if shouldValidate {
		binanceService, err := NewBinanceService(encryptedAPIKey, encryptedSecretKey)
		if err != nil {
			// 如果创建服务失败，记录错误但继续保存
			logrus.WithError(err).Warn("Failed to create Binance service for validation")
		} else {
			// 尝试验证，但不阻止保存
			if err := binanceService.ValidateAPICredentials(context.Background()); err != nil {
				// 记录验证失败，但仍然允许保存密钥
				logrus.WithError(err).Warn("API validation failed. Keys will be saved but may not work")
			} else {
				logrus.Info("API keys validated successfully")
			}
		}
	}

	updates := map[string]interface{}{
		"api_key":      encryptedAPIKey,
		"secret_key":   encryptedSecretKey,
		"is_encrypted": true,
	}

	return us.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (us *UserService) GetUserAPIKeys(userID uint) (string, string, error) {
	var user models.User
	if err := us.db.First(&user, userID).Error; err != nil {
		return "", "", err
	}

	if !user.IsEncrypted || user.APIKey == "" || user.SecretKey == "" {
		return "", "", errors.New("API密钥未设置")
	}

	apiKey, err := utils.DecryptAES(user.APIKey, config.AppConfig.Security.EncryptionKey)
	if err != nil {
		return "", "", err
	}

	secretKey, err := utils.DecryptAES(user.SecretKey, config.AppConfig.Security.EncryptionKey)
	if err != nil {
		return "", "", err
	}

	return apiKey, secretKey, nil
}

// ValidateUserAPIKeys 验证用户的API密钥是否有效
func (us *UserService) ValidateUserAPIKeys(userID uint) error {
	var user models.User
	if err := us.db.First(&user, userID).Error; err != nil {
		return err
	}

	if !user.IsEncrypted || user.APIKey == "" || user.SecretKey == "" {
		return errors.New("API密钥未设置")
	}

	// 创建币安服务实例
	binanceService, err := NewBinanceService(user.APIKey, user.SecretKey)
	if err != nil {
		return fmt.Errorf("创建币安服务失败: %w", err)
	}

	// 验证API凭证
	if err := binanceService.ValidateAPICredentials(context.Background()); err != nil {
		return fmt.Errorf("API密钥验证失败: %w", err)
	}

	return nil
}

func (us *UserService) UpdateUserProfile(userID uint, updates map[string]interface{}) error {
	allowedFields := []string{"username", "email"}
	filteredUpdates := make(map[string]interface{})

	for field, value := range updates {
		if utils.Contains(allowedFields, field) {
			strValue, ok := value.(string)
			if !ok {
				return errors.New("字段值必须是字符串类型")
			}
			switch field {
			case "username":
				if err := utils.ValidateUsername(strValue); err != nil {
					return err
				}
			case "email":
				if err := utils.ValidateEmail(strValue); err != nil {
					return err
				}
			}
			filteredUpdates[field] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return errors.New("没有有效的更新字段")
	}

	return us.db.Model(&models.User{}).Where("id = ?", userID).Updates(filteredUpdates).Error
}

func (us *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := us.db.First(&user, userID).Error; err != nil {
		return err
	}

	if err := user.CheckPassword(oldPassword); err != nil {
		return errors.New("原密码错误")
	}

	if err := utils.ValidatePassword(newPassword); err != nil {
		return err
	}

	if err := user.HashPassword(newPassword); err != nil {
		return err
	}

	return us.db.Model(&user).Update("password", user.Password).Error
}

func (us *UserService) GetAllUsers(page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 先计算总数
	if err := us.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 然后查询分页数据
	offset := (page - 1) * limit
	if err := us.db.Select("*").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	for i := range users {
		// 先保存API密钥状态
		hasAPIKey := users[i].APIKey != "" && users[i].SecretKey != ""
		
		// 设置HasAPIKey字段，用于前端显示
		users[i].HasAPIKey = hasAPIKey
		
		// 清除敏感信息
		users[i].Password = ""
		users[i].APIKey = ""
		users[i].SecretKey = ""
	}

	return users, total, nil
}

func (us *UserService) UpdateUserStatus(userID uint, status models.UserStatus) error {
	return us.db.Model(&models.User{}).Where("id = ?", userID).Update("status", status).Error
}

func (us *UserService) UpdateUserRole(userID uint, role models.UserRole) error {
	return us.db.Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error
}

func (us *UserService) DeleteUser(userID uint) error {
	return us.db.Delete(&models.User{}, userID).Error
}

func (us *UserService) GetUserStats(userID uint) (map[string]interface{}, error) {
	var strategiesCount int64
	var ordersCount int64
	var activeStrategiesCount int64

	us.db.Model(&models.Strategy{}).Where("user_id = ?", userID).Count(&strategiesCount)
	us.db.Model(&models.Order{}).Where("user_id = ?", userID).Count(&ordersCount)
	us.db.Model(&models.Strategy{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&activeStrategiesCount)

	stats := map[string]interface{}{
		"strategies_count":        strategiesCount,
		"orders_count":            ordersCount,
		"active_strategies_count": activeStrategiesCount,
	}

	return stats, nil
}
