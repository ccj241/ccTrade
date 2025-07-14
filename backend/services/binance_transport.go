package services

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// BinanceTransport 自定义的HTTP传输层，用于添加apiAgentCode参数
type BinanceTransport struct {
	apiAgentCode string
	baseTransport http.RoundTripper
	isFutures    bool
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
	// 只对POST请求添加apiAgentCode
	if req.Method == http.MethodPost && t.apiAgentCode != "" {
		// 克隆请求以避免修改原始请求
		newReq := req.Clone(req.Context())
		
		// 读取请求体
		var bodyBytes []byte
		if req.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(req.Body)
			req.Body.Close()
		}
		
		// 根据Content-Type处理请求体
		contentType := req.Header.Get("Content-Type")
		
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			// 处理表单数据
			values, _ := url.ParseQuery(string(bodyBytes))
			values.Set("apiAgentCode", t.apiAgentCode)
			newBody := values.Encode()
			
			newReq.Body = ioutil.NopCloser(strings.NewReader(newBody))
			newReq.ContentLength = int64(len(newBody))
			
			// 恢复原始请求的body
			req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
		} else if strings.Contains(contentType, "application/json") {
			// 对于JSON请求，币安API通常不使用POST的JSON body，而是使用查询参数
			// 所以我们需要在URL中添加参数
			q := newReq.URL.Query()
			q.Set("apiAgentCode", t.apiAgentCode)
			newReq.URL.RawQuery = q.Encode()
			
			// 恢复原始请求的body
			if len(bodyBytes) > 0 {
				newReq.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
				req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
			}
		} else {
			// 对于其他类型的请求，在URL查询参数中添加
			q := newReq.URL.Query()
			q.Set("apiAgentCode", t.apiAgentCode)
			newReq.URL.RawQuery = q.Encode()
			
			// 恢复原始请求的body
			if len(bodyBytes) > 0 {
				newReq.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
				req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
			}
		}
		
		// 使用新请求
		req = newReq
	}
	
	// 对于币安API，大多数POST请求的参数都在URL查询字符串中
	// 所以我们也需要检查GET请求（某些API可能使用GET但实际上是"查询"操作）
	if req.Method == http.MethodGet && t.apiAgentCode != "" {
		// 某些特定的端点可能需要apiAgentCode
		needsAgentCode := false
		path := req.URL.Path
		
		// 检查是否是需要apiAgentCode的端点
		// 根据币安文档，某些端点可能需要这个参数
		if strings.Contains(path, "/sapi/") || strings.Contains(path, "/api/v3/order") {
			needsAgentCode = true
		}
		
		if needsAgentCode {
			newReq := req.Clone(req.Context())
			q := newReq.URL.Query()
			q.Set("apiAgentCode", t.apiAgentCode)
			newReq.URL.RawQuery = q.Encode()
			req = newReq
		}
	}
	
	return t.baseTransport.RoundTrip(req)
}