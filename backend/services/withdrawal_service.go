package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/ccj241/cctrade/config"
	"github.com/ccj241/cctrade/models"
	"github.com/ccj241/cctrade/utils"
	"gorm.io/gorm"
)

type WithdrawalService struct {
	db          *gorm.DB
	userService *UserService
}

func NewWithdrawalService() *WithdrawalService {
	return &WithdrawalService{
		db:          config.DB,
		userService: NewUserService(),
	}
}

func (ws *WithdrawalService) CreateWithdrawalRule(userID uint, withdrawalData map[string]interface{}) (*models.Withdrawal, error) {
	withdrawal := &models.Withdrawal{
		UserID: userID,
	}

	if asset, ok := withdrawalData["asset"].(string); ok {
		withdrawal.Asset = utils.ToUpper(asset)
	} else {
		return nil, errors.New("币种不能为空")
	}

	if address, ok := withdrawalData["address"].(string); ok {
		withdrawal.Address = address
	} else {
		return nil, errors.New("提币地址不能为空")
	}

	if network, ok := withdrawalData["network"].(string); ok {
		withdrawal.Network = network
	}

	if amount, ok := withdrawalData["amount"].(float64); ok {
		if amount <= 0 {
			return nil, errors.New("提币数量必须大于0")
		}
		withdrawal.Amount = amount
	} else {
		return nil, errors.New("提币数量不能为空")
	}

	if minBalance, ok := withdrawalData["min_balance"].(float64); ok && minBalance >= 0 {
		withdrawal.MinBalance = minBalance
	}

	if triggerPrice, ok := withdrawalData["trigger_price"].(float64); ok && triggerPrice > 0 {
		withdrawal.TriggerPrice = triggerPrice
	}

	if autoWithdraw, ok := withdrawalData["auto_withdraw"].(bool); ok {
		withdrawal.AutoWithdraw = autoWithdraw
	}

	if description, ok := withdrawalData["description"].(string); ok {
		withdrawal.Description = description
	}

	if err := ws.db.Create(withdrawal).Error; err != nil {
		return nil, err
	}

	return withdrawal, nil
}

func (ws *WithdrawalService) GetUserWithdrawals(userID uint, page, limit int) ([]models.Withdrawal, int64, error) {
	var withdrawals []models.Withdrawal
	var total int64

	query := ws.db.Model(&models.Withdrawal{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&withdrawals).Error; err != nil {
		return nil, 0, err
	}

	return withdrawals, total, nil
}

func (ws *WithdrawalService) GetWithdrawalByID(userID, withdrawalID uint) (*models.Withdrawal, error) {
	var withdrawal models.Withdrawal
	if err := ws.db.Where("id = ? AND user_id = ?", withdrawalID, userID).First(&withdrawal).Error; err != nil {
		return nil, err
	}
	return &withdrawal, nil
}

func (ws *WithdrawalService) UpdateWithdrawal(userID, withdrawalID uint, updates map[string]interface{}) error {
	allowedFields := []string{"amount", "min_balance", "trigger_price", "auto_withdraw", "is_active", "description"}
	filteredUpdates := make(map[string]interface{})

	for field, value := range updates {
		if utils.Contains(allowedFields, field) {
			if field == "amount" && value.(float64) <= 0 {
				return errors.New("提币数量必须大于0")
			}
			filteredUpdates[field] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return errors.New("没有有效的更新字段")
	}

	return ws.db.Model(&models.Withdrawal{}).Where("id = ? AND user_id = ?", withdrawalID, userID).Updates(filteredUpdates).Error
}

func (ws *WithdrawalService) ToggleWithdrawal(userID, withdrawalID uint) error {
	var withdrawal models.Withdrawal
	if err := ws.db.Where("id = ? AND user_id = ?", withdrawalID, userID).First(&withdrawal).Error; err != nil {
		return err
	}

	return ws.db.Model(&withdrawal).Update("is_active", !withdrawal.IsActive).Error
}

func (ws *WithdrawalService) DeleteWithdrawal(userID, withdrawalID uint) error {
	return ws.db.Where("id = ? AND user_id = ?", withdrawalID, userID).Delete(&models.Withdrawal{}).Error
}

func (ws *WithdrawalService) ExecuteWithdrawal(withdrawal *models.Withdrawal) error {
	apiKey, secretKey, err := ws.userService.GetUserAPIKeys(withdrawal.UserID)
	if err != nil {
		return err
	}

	binanceService, err := NewBinanceService(apiKey, secretKey)
	if err != nil {
		return err
	}

	account, err := binanceService.GetAccountInfo(context.Background())
	if err != nil {
		return err
	}

	var currentBalance float64
	for _, balance := range account.Balances {
		if balance.Asset == withdrawal.Asset {
			free, _ := strconv.ParseFloat(balance.Free, 64)
			currentBalance = free
			break
		}
	}

	if currentBalance < withdrawal.MinBalance+withdrawal.Amount {
		return nil
	}

	if withdrawal.TriggerPrice > 0 {
		symbol := withdrawal.Asset + "USDT"
		currentPrice, err := binanceService.GetPrice(context.Background(), symbol)
		if err != nil {
			log.Printf("获取价格失败: %v", err)
			return nil
		}

		if currentPrice < withdrawal.TriggerPrice {
			return nil
		}
	}

	resp, err := binanceService.Withdraw(context.Background(), withdrawal.Asset, withdrawal.Address, withdrawal.Network, withdrawal.Amount, "")
	if err != nil {
		return err
	}

	history := &models.WithdrawalHistory{
		UserID:       withdrawal.UserID,
		WithdrawalID: &withdrawal.ID,
		Asset:        withdrawal.Asset,
		Amount:       withdrawal.Amount,
		Address:      withdrawal.Address,
		Network:      withdrawal.Network,
		TxID:         resp.ID,
		Status:       "PENDING",
		ApplyTime:    utils.GetCurrentTimestamp(),
	}

	if err := ws.db.Create(history).Error; err != nil {
		log.Printf("保存提币历史失败: %v", err)
	}

	log.Printf("提币成功: %s %.8f %s", withdrawal.Asset, withdrawal.Amount, withdrawal.Address)

	return nil
}

func (ws *WithdrawalService) CheckWithdrawalRules() error {
	var withdrawals []models.Withdrawal
	if err := ws.db.Where("is_active = ? AND auto_withdraw = ?", true, true).Find(&withdrawals).Error; err != nil {
		return err
	}

	for _, withdrawal := range withdrawals {
		if err := ws.ExecuteWithdrawal(&withdrawal); err != nil {
			log.Printf("执行提币规则失败: %v", err)
		}
	}

	return nil
}

func (ws *WithdrawalService) GetUserWithdrawalHistory(userID uint, page, limit int) ([]models.WithdrawalHistory, int64, error) {
	var histories []models.WithdrawalHistory
	var total int64

	query := ws.db.Model(&models.WithdrawalHistory{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

func (ws *WithdrawalService) SyncWithdrawalHistory(userID uint) error {
	apiKey, secretKey, err := ws.userService.GetUserAPIKeys(userID)
	if err != nil {
		return err
	}

	binanceService, err := NewBinanceService(apiKey, secretKey)
	if err != nil {
		return err
	}
	withdrawals, err := binanceService.GetWithdrawHistory(context.Background(), "", 100)
	if err != nil {
		return err
	}

	for _, withdraw := range withdrawals {
		amount, _ := strconv.ParseFloat(withdraw.Amount, 64)
		fee, _ := strconv.ParseFloat(withdraw.TransactionFee, 64)
		applyTime, _ := strconv.ParseInt(withdraw.ApplyTime, 10, 64)
		completeTime, _ := strconv.ParseInt(withdraw.CompleteTime, 10, 64)

		history := &models.WithdrawalHistory{
			UserID:       userID,
			Asset:        withdraw.Coin,
			Amount:       amount,
			Fee:          fee,
			Address:      withdraw.Address,
			Network:      withdraw.Network,
			TxID:         withdraw.TxID,
			Status:       fmt.Sprintf("%d", withdraw.Status),
			ApplyTime:    applyTime / 1000,
			CompleteTime: completeTime / 1000,
		}

		var existing models.WithdrawalHistory
		if err := ws.db.Where("user_id = ? AND tx_id = ?", userID, withdraw.TxID).First(&existing).Error; err != nil {
			if err := ws.db.Create(history).Error; err != nil {
				log.Printf("保存提币历史失败: %v", err)
			}
		} else {
			if err := ws.db.Model(&existing).Updates(history).Error; err != nil {
				log.Printf("更新提币历史失败: %v", err)
			}
		}
	}

	return nil
}

func (ws *WithdrawalService) GetWithdrawalStats(userID uint) (map[string]interface{}, error) {
	var totalRules int64
	var activeRules int64
	var totalWithdrawals int64
	var totalAmount float64

	ws.db.Model(&models.Withdrawal{}).Where("user_id = ?", userID).Count(&totalRules)
	ws.db.Model(&models.Withdrawal{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&activeRules)
	ws.db.Model(&models.WithdrawalHistory{}).Where("user_id = ?", userID).Count(&totalWithdrawals)

	var histories []models.WithdrawalHistory
	ws.db.Where("user_id = ? AND status = ?", userID, "COMPLETED").Find(&histories)
	for _, history := range histories {
		totalAmount += history.Amount
	}

	stats := map[string]interface{}{
		"total_rules":       totalRules,
		"active_rules":      activeRules,
		"total_withdrawals": totalWithdrawals,
		"total_amount":      totalAmount,
	}

	return stats, nil
}
