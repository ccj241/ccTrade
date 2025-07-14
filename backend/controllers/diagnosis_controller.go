package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/ccj241/cctrade/services"
	"github.com/ccj241/cctrade/utils"
	"strconv"
)

// DiagnoseBinanceAPI 诊断Binance API连接问题
func DiagnoseBinanceAPI(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	// 创建用户服务实例
	userService := services.NewUserService()
	
	// 获取用户的API密钥
	apiKey, secretKey, err := userService.GetUserAPIKeys(userID)
	if err != nil {
		utils.BadRequestResponse(c, "请先设置API密钥")
		return
	}
	
	// 创建Binance服务
	binanceService, err := services.NewBinanceService(apiKey, secretKey)
	if err != nil {
		utils.InternalServerErrorResponse(c, "创建Binance服务失败: "+err.Error())
		return
	}
	
	// 执行诊断
	diagnosis, err := binanceService.DiagnoseAPIConnection(c.Request.Context())
	if err != nil {
		utils.InternalServerErrorResponse(c, "诊断执行失败: "+err.Error())
		return
	}
	
	// 分析诊断结果
	recommendations := analyzeDiagnosis(diagnosis)
	
	response := map[string]interface{}{
		"diagnosis":       diagnosis,
		"recommendations": recommendations,
	}
	
	utils.SuccessResponse(c, response)
}

// analyzeDiagnosis 分析诊断结果并提供建议
func analyzeDiagnosis(diagnosis map[string]interface{}) []string {
	var recommendations []string

	// 检查加密密钥
	if valid, ok := diagnosis["encryption_key_valid"].(bool); ok && !valid {
		recommendations = append(recommendations, "加密密钥长度不正确，应为32个字符")
	}

	// 检查API密钥
	if present, ok := diagnosis["api_key_present"].(bool); ok && !present {
		recommendations = append(recommendations, "API密钥为空，请检查是否正确设置")
	}
	if present, ok := diagnosis["secret_key_present"].(bool); ok && !present {
		recommendations = append(recommendations, "Secret密钥为空，请检查是否正确设置")
	}

	// 检查解密错误
	if decryptErr, ok := diagnosis["credential_decrypt_error"].(string); ok && decryptErr != "" {
		recommendations = append(recommendations, "密钥解密失败: "+decryptErr)
		recommendations = append(recommendations, "请检查加密密钥是否与加密时使用的密钥一致")
	}

	// 检查网络连接
	if accessible, ok := diagnosis["network_accessible"].(bool); ok && !accessible {
		recommendations = append(recommendations, "无法连接到Binance服务器，请检查网络连接")
		if serverErr, ok := diagnosis["server_time_error"].(string); ok {
			recommendations = append(recommendations, "网络错误详情: "+serverErr)
		}
	}

	// 检查时间同步
	if timeDiff, ok := diagnosis["time_diff_ms"].(int64); ok {
		if timeDiff > 5000 || timeDiff < -5000 {
			recommendations = append(recommendations, "系统时间与Binance服务器时间相差过大，建议同步系统时间")
		}
	}

	// 检查认证测试结果
	authTestPassed := false
	for _, window := range []int64{5000, 30000, 60000} {
		key := "auth_test_recv_window_" + strconv.FormatInt(window, 10)
		if passed, ok := diagnosis[key].(bool); ok && passed {
			authTestPassed = true
			break
		}
	}

	if !authTestPassed {
		recommendations = append(recommendations, "API认证失败，可能的原因：")
		recommendations = append(recommendations, "- API密钥或Secret密钥不正确")
		recommendations = append(recommendations, "- API密钥未激活或已过期")
		recommendations = append(recommendations, "- IP地址未加入白名单（如果启用了IP限制）")
		recommendations = append(recommendations, "- 测试网/主网设置与API密钥不匹配")
		
		// 检查具体的认证错误
		for _, window := range []int64{5000, 30000, 60000} {
			errKey := "auth_error_recv_window_" + strconv.FormatInt(window, 10)
			if authErr, ok := diagnosis[errKey].(string); ok && authErr != "" {
				recommendations = append(recommendations, "认证错误 (recvWindow="+strconv.FormatInt(window, 10)+"): "+authErr)
			}
		}
	}

	// 检查测试网设置
	if testnet, ok := diagnosis["testnet_enabled"].(bool); ok && testnet {
		recommendations = append(recommendations, "当前使用测试网，请确保API密钥是从测试网生成的")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "API连接正常，所有测试通过")
	}

	return recommendations
}