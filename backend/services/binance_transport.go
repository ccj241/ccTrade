package services

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
)

// BinanceTransport 自定义的HTTP传输层，用于添加apiAgentCode参数
type BinanceTransport struct {
	apiAgentCode  string
	baseTransport http.RoundTripper
	isFutures     bool
}

// NewBinanceTransport 创建新的币安传输层
func NewBinanceTransport(apiAgentCode string, isFutures bool) *BinanceTransport {
	return &BinanceTransport{
		apiAgentCode:  apiAgentCode,
		baseTransport: http.DefaultTransport,
		isFutures:     isFutures,
	}
}

// RoundTrip 实现http.RoundTripper接口
func (t *BinanceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 币安API的大多数请求（包括POST）都使用URL查询参数
	// apiAgentCode应该作为查询参数添加到需要的端点
	if t.apiAgentCode != "" {
		// 检查是否是需要apiAgentCode的端点
		needsAgentCode := false
		path := req.URL.Path

		// 根据币安文档，以下端点需要apiAgentCode参数
		// - 订单相关的端点
		// - 某些账户操作端点
		if strings.Contains(path, "/api/v3/order") ||
			strings.Contains(path, "/fapi/v1/order") ||
			strings.Contains(path, "/sapi/v1/") ||
			(req.Method == http.MethodPost && (strings.Contains(path, "order") || strings.Contains(path, "withdraw"))) {
			needsAgentCode = true
		}

		if needsAgentCode {
			// 克隆请求以避免修改原始请求
			newReq := req.Clone(req.Context())

			// 添加apiAgentCode到查询参数
			q := newReq.URL.Query()
			q.Set("apiAgentCode", t.apiAgentCode)
			newReq.URL.RawQuery = q.Encode()

			// 如果有请求体，需要保持不变
			if req.Body != nil {
				bodyBytes, _ := ioutil.ReadAll(req.Body)
				req.Body.Close()

				// 恢复原始请求的body
				req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
				newReq.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
			}

			req = newReq
		}
	}

	return t.baseTransport.RoundTrip(req)
}
